package cache

import (
	"encoding/json"
	"errors"

	"github.com/bradfitz/gomemcache/memcache"
)

var Client *memcache.Client

// Initialize initializes the cache client with Memcache servers.
func Initialize(serverList ...string) {
	Client = memcache.New(serverList...)
}

const DefaultTTL = 24 * 60 * 60 // 24 hours in seconds

// SetCache sets data in the cache. If ttl is provided, it overrides the default TTL.
func SetCache(key string, value interface{}, ttl ...int32) error {
	// Serialize value to a cache-compatible format (e.g., JSON).
	serializedValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	// Determine TTL: use the first ttl argument if provided, otherwise use DefaultTTL.
	finalTTL := int32(DefaultTTL)
	if len(ttl) > 0 && ttl[0] > 0 {
		finalTTL = ttl[0]
	}

	// Use the Memcache client to set the value with the determined TTL.
	return Client.Set(&memcache.Item{Key: key, Value: serializedValue, Expiration: finalTTL})
}

// GetCache retrieves data from the cache, deserializing it into the target object.
func GetCache(key string, target interface{}) error {
	// Use the Memcache client to get the value.
	item, err := Client.Get(key)
	if err != nil {
		if err == memcache.ErrCacheMiss {
			return errors.New("cache miss")
		}
		return err
	}

	// Deserialize the value into the target object.
	return json.Unmarshal(item.Value, target)
}

// DeleteCache removes the data associated with the given key from the cache.
func DeleteCache(key string) error {
	err := Client.Delete(key)
	if err == memcache.ErrCacheMiss {
		// If it's a cache miss, consider it already deleted.
		return nil
	}
	return err
}
