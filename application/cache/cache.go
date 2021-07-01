package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"

	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/conversions"
	"github.com/silinternational/wecarry-api/models"
)

var (
	RequestsCache         *cache.Cache
	PublicRequestsKeyName string = "public"
)

func init() {
	ring := redis.NewRing(&redis.RingOptions{
		Addrs: map[string]string{
			"redis": "redis:6379",
		},
	})

	RequestsCache = cache.New(&cache.Options{
		Redis:      ring,
		LocalCache: cache.NewTinyLFU(1000, time.Minute),
		// use json in lieu of msgpack for encoding/decoding
		Marshal:   json.Marshal,
		Unmarshal: json.Unmarshal,
	})
}

// Wrapper for cache write
func CacheWrite(ctx context.Context, organization string, requestsMap interface{}) error {
	return RequestsCache.Set(&cache.Item{
		Ctx:   ctx,
		Key:   organization,
		Value: requestsMap,
		TTL:   time.Hour,
	})
}

// Get all requests visible to user in specified organization from cache, fetching data from database if needed
func GetVisibleRequests(ctx context.Context, orgaziation models.Organization) ([]api.RequestAbridged, error) {
	// get private requests
	_, privateRequestsMap, err := getOrCreateCacheEntryPrivate(ctx, orgaziation)
	if err != nil {
		return nil, err
	}

	// get public requests
	_, publicRequestsMap, err := getOrCreateCacheEntryPublic(ctx)
	if err != nil {
		return nil, err
	}

	// compose them to get list of all requests visible to the current user
	visibleRequestsList := make([]api.RequestAbridged, len(publicRequestsMap)+len(privateRequestsMap))
	for _, privateRequest := range privateRequestsMap {
		visibleRequestsList = append(visibleRequestsList, privateRequest)
	}
	for _, publicRequest := range publicRequestsMap {
		visibleRequestsList = append(visibleRequestsList, publicRequest)
	}

	return visibleRequestsList, nil
}

// Wrapper for cache read
func CacheRead(ctx context.Context, organization string, requestsMap interface{}) error {
	return RequestsCache.Get(ctx, organization, requestsMap)
}

// Rebuild cache after a request is created or updated
func CacheRebuildOnChangedRequest(ctx context.Context, organization models.Organization, request models.Request) error {
	// update private cache
	newEntryCreated, requestsMap, err := getOrCreateCacheEntryPrivate(ctx, organization)
	if err != nil {
		return err
	}
	// may need to update existing cache entry with new data
	if !newEntryCreated {
		_, cachedPrivately := requestsMap[request.UUID.String()]
		// was cached privately, but request visbility has completed. Just need to remove from private cache.
		if cachedPrivately && isCompleted(request) {
			delete(requestsMap, request.UUID.String())
			return CacheWrite(ctx, organization.Name, requestsMap)
		}
		// was cached privately, but request visbility has changed to public.
		// Need to remove from private and add to public cache.
		if cachedPrivately && isPublic(request) {
			delete(requestsMap, request.UUID.String())
			if err := CacheWrite(ctx, organization.Name, requestsMap); err != nil {
				return err
			}
		}

		// should be cached privately. Need to update or create an entry in the private cache.
		// Cannot return yet in case the request was previously public (must remove from public cache)
		if isPrivate(request) {
			requestAbridged, err := conversions.ConvertRequestToAPITypeAbridged(ctx, request)
			if err != nil {
				return err
			}
			requestsMap[request.UUID.String()] = requestAbridged
			if err := CacheWrite(ctx, organization.Name, requestsMap); err != nil {
				return err
			}
		}

		// if request was and remains public and active, we update it below

	}
	// update public cache
	newEntryCreated, requestsMap, err = getOrCreateCacheEntryPublic(ctx)
	if err != nil {
		return err
	}
	// may need to update existing cache entry with new data
	if !newEntryCreated {
		_, cachedPublicly := requestsMap[request.UUID.String()]
		// was cached publicly, but request visbility has completed. Just need to remove from public cache.
		if cachedPublicly && isCompleted(request) {
			delete(requestsMap, request.UUID.String())
			return CacheWrite(ctx, organization.Name, requestsMap)
		}
		// was cached publicly, but request visbility has changed to private.
		// Already added to private cache, now need to add to public cache.
		if cachedPublicly && isPrivate(request) {
			delete(requestsMap, request.UUID.String())
			if err := CacheWrite(ctx, organization.Name, requestsMap); err != nil {
				return err
			}
		}

		// should be cached publicly. Just need to update or create an entry in the public cache,
		// as we've already dealt with the case where visibility was originally private
		if isPublic(request) {
			requestAbridged, err := conversions.ConvertRequestToAPITypeAbridged(ctx, request)
			if err != nil {
				return err
			}
			requestsMap[request.UUID.String()] = requestAbridged
			return CacheWrite(ctx, organization.Name, requestsMap)
		}
	}
	return nil
}

// Gets cache entry for organization, or creates one if none exists
// Returns true if new cache entry was created
func getOrCreateCacheEntryPrivate(ctx context.Context, organization models.Organization) (bool, map[string]api.RequestAbridged, error) {
	var requestsMap map[string]api.RequestAbridged
	// if a cache value does not exist, we create it
	if err := CacheRead(ctx, organization.Name, requestsMap); err != nil {
		tx := models.Tx(ctx)

		var org models.Organization
		if err := org.FindByUUID(tx, organization.UUID.String()); err != nil {
			return false, nil, err
		}

		// RequestFilterParams is currently empty because the UI is not using it
		filter := models.RequestFilterParams{}

		requests := models.Requests{}
		if err := requests.FindByOrganization(tx, org, filter); err != nil {
			return false, nil, err
		}

		requestsList, err := conversions.ConvertRequestsAbridged(ctx, requests)
		if err != nil {
			return false, nil, err
		}
		requestsMap := make(map[string]api.RequestAbridged)
		for _, requestEntry := range requestsList {
			requestsMap[requestEntry.ID.String()] = requestEntry
		}

		return true, requestsMap, CacheWrite(ctx, organization.Name, requestsMap)
	}

	return false, requestsMap, nil
}

// Gets cache entry for organization, or creates one if none exists.
// Returns true if new cache entry was created
func getOrCreateCacheEntryPublic(ctx context.Context) (bool, map[string]api.RequestAbridged, error) {
	var requestsMap map[string]api.RequestAbridged
	// if a cache value does not exist, we create it
	if err := CacheRead(ctx, PublicRequestsKeyName, requestsMap); err != nil {
		tx := models.Tx(ctx)

		// RequestFilterParams is currently empty because the UI is not using it
		filter := models.RequestFilterParams{}

		requests := models.Requests{}
		if err := requests.FindPublic(tx, filter); err != nil {
			return false, nil, err
		}
		requestsList, err := conversions.ConvertRequestsAbridged(ctx, requests)
		if err != nil {
			return false, nil, err
		}
		requestsMap := make(map[string]api.RequestAbridged)
		for _, requestEntry := range requestsList {
			requestsMap[requestEntry.ID.String()] = requestEntry
		}

		return true, requestsMap, CacheWrite(ctx, PublicRequestsKeyName, requestsMap)
	}
	return false, requestsMap, nil
}

// we should cache non-completed public requests publicly
func isPublic(request models.Request) bool {
	return request.Visibility == models.RequestVisibilityAll
}

// we should cache non-completed private requests privately (at the organizational level)
func isPrivate(request models.Request) bool {
	return request.Visibility != models.RequestVisibilityAll
}

// we should remove completed requests from the cache
func isCompleted(request models.Request) bool {
	return request.Status.String() == models.RequestStatusRemoved.String() ||
		request.Status.String() == models.RequestStatusCompleted.String()
}
