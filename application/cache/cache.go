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

var RequestsCache *cache.Cache

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
func CacheWrite(ctx context.Context, organization string, requestsList interface{}) error {
	return RequestsCache.Set(&cache.Item{
		Ctx:   ctx,
		Key:   organization,
		Value: requestsList,
		TTL:   time.Hour,
	})
}

// Wrapper for cache read
func CacheRead(ctx context.Context, organization string, requestsList interface{}) error {
	return RequestsCache.Get(ctx, organization, requestsList)
}

// TODO: error, must also rebuild the public requests list IF request is public
// Rebuild cache after a request is created or updated
func CacheRebuildOnCreate(ctx context.Context, organization string, request api.RequestAbridged) error {
	requestsList, err := createCacheEntryIfEmpty(ctx, organization, request)
	if err != nil {
		return err
	}
	// new cache entry was created
	if requestsList == nil {
		return nil
	}
	// if a cache value exists, then we only need to append the new request
	requestsList = append(requestsList, request)
	return CacheWrite(ctx, organization, requestsList)
}

// TODO: finish implementing
// Rebuild cache after a request is created or updated
func CacheRebuildOnUpdate(ctx context.Context, organization string, request api.RequestAbridged) error {
	requestsList, err := createCacheEntryIfEmpty(ctx, organization, request)
	if err != nil {
		return err
	}
	// new cache entry was created
	if requestsList == nil {
		return nil
	}
	// TODO see whether there is a better way to do this
	//		An alternative would be to query the database with a new function like models.FindByOrganization (see below),
	//		but that would add a significant extra load on the database every time a request is updated,
	// We need to modify the existing cache value, so we linear search through the list of requests to find and update it
	return CacheWrite(ctx, organization, requestsList)
}

// Create cache entry if none exists
func createCacheEntryIfEmpty(ctx context.Context, organization string, request api.RequestAbridged) ([]api.RequestAbridged, error) {
	var requestsList []api.RequestAbridged
	// if a cache value does not exist, we create it
	if err := CacheRead(ctx, organization, requestsList); err != nil {
		tx := models.Tx(ctx)

		var org models.Organization
		if err := org.FindByUUID(tx, request.Organization.ID.String()); err != nil {
			return nil, err
		}

		// RequestFilterParams is currently empty because the UI is not using it
		filter := models.RequestFilterParams{}

		requests := models.Requests{}
		if err := requests.FindByOrganization(tx, org, filter); err != nil {
			return nil, err
		}
		requestsList, err := conversions.ConvertRequestsAbridged(ctx, requests)
		if err != nil {
			return nil, err
		}
		return nil, CacheWrite(ctx, organization, requestsList)
	}
	return requestsList, nil
}
