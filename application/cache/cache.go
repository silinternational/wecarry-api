package cache

import (
	"context"
	"encoding/json"
	"errors"
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
func GetVisibleRequests(ctx context.Context, orgs []models.Organization) ([]api.RequestAbridged, error) {
	// get and de-duplicate private requests for all organizations a user belongs to
	privateRequestsMap := make(map[string]api.RequestAbridged)
	for _, organization := range orgs {
		_, privateRequestsMapPartial, err := getOrCreateCacheEntryPrivate(ctx, organization)
		if err != nil {
			return nil, errors.New("error in cache get visible requests: " + err.Error())
		}
		for requestID, request := range privateRequestsMapPartial {
			privateRequestsMap[requestID] = request
		}
	}

	// get public requests
	_, publicRequestsMap, err := getOrCreateCacheEntryPublic(ctx)
	if err != nil {
		return nil, errors.New("error in cache get visible requests: " + err.Error())
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

// Rebuild cache after a request is created
// We cache non-finished public requests publicly with cache key "requests-allusers"
// We cache non-finished private caches privately with cache key "requests-orgname-affiliated_"
// We remove a finished request from its respective cache
func CacheRebuildOnNewRequest(ctx context.Context, request models.Request) error {
	organization := request.Organization
	// add to public cache only
	if request.Visibility == models.RequestVisibilityAll {
		newEntryCreated, requestsMap, err := getOrCreateCacheEntryPublic(ctx)
		if err != nil {
			return errors.New("error in cache rebuild (creation): " + err.Error())
		}
		if !newEntryCreated {
			if err := updateRequestPublicCache(ctx, request, requestsMap); err != nil {
				return errors.New("error in cache rebuild (creation): " + err.Error())
			}
		}
	} else {
		// add to private cache only
		newEntryCreated, requestsMap, err := getOrCreateCacheEntryPrivate(ctx, organization)
		if err != nil {
			return errors.New("error in cache rebuild: " + err.Error())
		}
		if !newEntryCreated {
			if err := updateRequestPrivateCache(ctx, request, requestsMap); err != nil {
				return errors.New("error in cache rebuild: " + err.Error())
			}
		}
	}
	return nil
}

// Rebuild cache after a request is updated
// We cache non-finished public requests publicly with cache key "requests-allusers"
// We cache non-finished private caches privately with cache key "requests-orgname-affiliated_"
// We remove a finished request from its respective cache
func CacheRebuildOnChangedRequest(ctx context.Context, request models.Request) error {
	organization := request.Organization
	// update private cache
	newEntryCreated, requestsMap, err := getOrCreateCacheEntryPrivate(ctx, organization)
	if err != nil {
		return errors.New("error in cache rebuild: " + err.Error())
	}
	if !newEntryCreated {
		if err := updateRequestPrivateCache(ctx, request, requestsMap); err != nil {
			return errors.New("error in cache rebuild: " + err.Error())
		}
	}

	// update public cache
	newEntryCreated, requestsMap, err = getOrCreateCacheEntryPublic(ctx)
	if err != nil {
		return errors.New("error in cache rebuild: " + err.Error())
	}
	if !newEntryCreated {
		if err := updateRequestPublicCache(ctx, request, requestsMap); err != nil {
			return errors.New("error in cache rebuild: " + err.Error())
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

		// RequestFilterParams is currently empty because the UI is not using it
		filter := models.RequestFilterParams{}

		requests := models.Requests{}
		if err := requests.FindByOrganization(tx, organization, filter); err != nil {
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

func updateRequestPrivateCache(ctx context.Context, request models.Request, requestsMap map[string]api.RequestAbridged) error {
	organization := request.Organization
	_, cachedPrivately := requestsMap[request.UUID.String()]
	// was cached privately, but request has been finished. Just need to remove from private cache.
	if cachedPrivately && request.IsFinished() {
		delete(requestsMap, request.UUID.String())
		return cacheWrite(ctx, PrivateRequestKeyPrefix+organization.Name, requestsMap)
	}
	// was cached privately, but request visibility has changed to public.
	// Need to remove from private and add to public cache.
	if cachedPrivately && request.IsPublic() {
		delete(requestsMap, request.UUID.String())
		if err := cacheWrite(ctx, PrivateRequestKeyPrefix+organization.Name, requestsMap); err != nil {
			return err
		}
	}

	// should be cached privately. Need to update or create an entry in the private cache.
	if request.IsPrivate() {
		requestAbridged, err := models.ConvertRequestToAPITypeAbridged(ctx, request)
		if err != nil {
			return err
		}
		requestsMap[request.UUID.String()] = requestAbridged
		if err := cacheWrite(ctx, PrivateRequestKeyPrefix+organization.Name, requestsMap); err != nil {
			return err
		}
	}
	return nil
}

func updateRequestPublicCache(ctx context.Context, request models.Request, requestsMap map[string]api.RequestAbridged) error {
	_, cachedPublicly := requestsMap[request.UUID.String()]
	// was cached publicly, but request has been finished. Just need to remove from public cache.
	if cachedPublicly && request.IsFinished() {
		delete(requestsMap, request.UUID.String())
		return cacheWrite(ctx, PublicRequestKey, requestsMap)
	}
	// was cached publicly, but request visibility has changed to private.
	// Need to remove from public cache.
	if cachedPublicly && request.IsPrivate() {
		delete(requestsMap, request.UUID.String())
		if err := cacheWrite(ctx, PublicRequestKey, requestsMap); err != nil {
			return err
		}
	}

	// should be cached publicly, need to update or create an entry in the public cache.
	if request.IsPublic() {
		requestAbridged, err := models.ConvertRequestToAPITypeAbridged(ctx, request)
		if err != nil {
			return err
		}
		requestsMap[request.UUID.String()] = requestAbridged
		return cacheWrite(ctx, PublicRequestKey, requestsMap)
	}
	return nil
}
