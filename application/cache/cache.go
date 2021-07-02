package cache

import (
	"context"
	"encoding/json"
	"time"

	rediscache "github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"

	"github.com/silinternational/wecarry-api/api"
	"github.com/silinternational/wecarry-api/domain"
	"github.com/silinternational/wecarry-api/models"
)

var (
	RequestsCache           *rediscache.Cache
	PrivateRequestKeyPrefix string = "requests-orgname-affiliated_"
	PublicRequestKey        string = "requests-allusers"
)

func init() {
	ring := redis.NewRing(&redis.RingOptions{
		Addrs: map[string]string{
			domain.Env.RedisInstanceName: domain.Env.RedisInstanceHostPort,
		},
	})

	RequestsCache = rediscache.New(&rediscache.Options{
		Redis:      ring,
		LocalCache: rediscache.NewTinyLFU(1000, time.Minute),
		// use json in lieu of msgpack for encoding/decoding
		Marshal:   json.Marshal,
		Unmarshal: json.Unmarshal,
	})
}

// Wrapper for cache write
func cacheWrite(ctx context.Context, key string, data interface{}) error {
	return RequestsCache.Set(&rediscache.Item{
		Ctx:   ctx,
		Key:   key,
		Value: data,
		TTL:   time.Hour,
	})
}

// Wrapper for cache read
func cacheRead(ctx context.Context, organization string, requestsMap interface{}) error {
	return RequestsCache.Get(ctx, organization, requestsMap)
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
	visibleRequestsList := []api.RequestAbridged{}
	for _, privateRequest := range privateRequestsMap {
		visibleRequestsList = append(visibleRequestsList, privateRequest)
	}
	for _, publicRequest := range publicRequestsMap {
		visibleRequestsList = append(visibleRequestsList, publicRequest)
	}

	return visibleRequestsList, nil
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
			return cacheWrite(ctx, PrivateRequestKeyPrefix+organization.Name, requestsMap)
		}
		// was cached privately, but request visbility has changed to public.
		// Need to remove from private and add to public cache.
		if cachedPrivately && isPublic(request) {
			delete(requestsMap, request.UUID.String())
			if err := cacheWrite(ctx, PrivateRequestKeyPrefix+organization.Name, requestsMap); err != nil {
				return err
			}
		}

		// should be cached privately. Need to update or create an entry in the private cache.
		// Cannot return yet in case the request was previously public (must remove from public cache)
		if isPrivate(request) {
			requestAbridged, err := models.ConvertRequestToAPITypeAbridged(ctx, request)
			if err != nil {
				return err
			}
			requestsMap[request.UUID.String()] = requestAbridged
			if err := cacheWrite(ctx, PrivateRequestKeyPrefix+organization.Name, requestsMap); err != nil {
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
			return cacheWrite(ctx, PublicRequestKey, requestsMap)
		}
		// was cached publicly, but request visbility has changed to private.
		// Already added to private cache, now need to add to public cache.
		if cachedPublicly && isPrivate(request) {
			delete(requestsMap, request.UUID.String())
			if err := cacheWrite(ctx, PublicRequestKey, requestsMap); err != nil {
				return err
			}
		}

		// should be cached publicly. Just need to update or create an entry in the public cache,
		// as we've already dealt with the case where visibility was originally private
		if isPublic(request) {
			requestAbridged, err := models.ConvertRequestToAPITypeAbridged(ctx, request)
			if err != nil {
				return err
			}
			requestsMap[request.UUID.String()] = requestAbridged
			return cacheWrite(ctx, PublicRequestKey, requestsMap)
		}
	}
	return nil
}

// Gets cache entry for organization, or creates one if none exists
// Returns true if new cache entry was created
func getOrCreateCacheEntryPrivate(ctx context.Context, organization models.Organization) (bool, map[string]api.RequestAbridged, error) {
	var requestsMap map[string]api.RequestAbridged
	// if a cache value does not exist, we create it
	if err := cacheRead(ctx, PrivateRequestKeyPrefix+organization.Name, requestsMap); err != nil {
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

		requestsList, err := models.ConvertRequestsAbridged(ctx, requests)
		if err != nil {
			return false, nil, err
		}
		requestsMap := make(map[string]api.RequestAbridged)
		for _, requestEntry := range requestsList {
			requestsMap[requestEntry.ID.String()] = requestEntry
		}

		return true, requestsMap, cacheWrite(ctx, PrivateRequestKeyPrefix+organization.Name, requestsMap)
	}

	return false, requestsMap, nil
}

// Gets cache entry for organization, or creates one if none exists.
// Returns true if new cache entry was created
func getOrCreateCacheEntryPublic(ctx context.Context) (bool, map[string]api.RequestAbridged, error) {
	var requestsMap map[string]api.RequestAbridged
	// if a cache value does not exist, we create it
	if err := cacheRead(ctx, PublicRequestKey, requestsMap); err != nil {
		tx := models.Tx(ctx)

		// RequestFilterParams is currently empty because the UI is not using it
		filter := models.RequestFilterParams{}

		requests := models.Requests{}
		if err := requests.FindPublic(tx, filter); err != nil {
			return false, nil, err
		}
		requestsList, err := models.ConvertRequestsAbridged(ctx, requests)
		if err != nil {
			return false, nil, err
		}
		requestsMap := make(map[string]api.RequestAbridged)
		for _, requestEntry := range requestsList {
			requestsMap[requestEntry.ID.String()] = requestEntry
		}

		return true, requestsMap, cacheWrite(ctx, PublicRequestKey, requestsMap)
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
