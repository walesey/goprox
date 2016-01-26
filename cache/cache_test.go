package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSimpleCache_Set(t *testing.T) {
	sc := &SimpleCache{store: make(map[string]simpleCacheItem)}
	sc.Set("key", "value", -1)
	assert.EqualValues(t, sc.store["key"].value, "value", "SimpleCache.Set should store a value in the map")
}

func TestSimpleCache_Get(t *testing.T) {
	sc := &SimpleCache{store: make(map[string]simpleCacheItem)}
	sc.store["key"] = simpleCacheItem{
		value:     "value",
		ttl:       5,
		createdAt: time.Now().UTC(),
	}
	value, err := sc.Get("key")
	assert.Nil(t, err, "SimpleCache.Get shouldn't return an error")
	assert.EqualValues(t, value, "value", "SimpleCache.Get should return the value stored in the map")
}

func TestSimpleCache_Get_Expired(t *testing.T) {
	sc := &SimpleCache{store: make(map[string]simpleCacheItem)}
	sc.store["key"] = simpleCacheItem{
		value:     "value",
		ttl:       5,
		createdAt: time.Now().UTC().Add(-6 * time.Second),
	}
	_, err := sc.Get("key")
	assert.NotNil(t, err, "SimpleCache.Get should return an error when the value is expired")
}

func TestSimpleCache_Get_Expired_NoTTL(t *testing.T) {
	sc := &SimpleCache{store: make(map[string]simpleCacheItem)}
	sc.store["key"] = simpleCacheItem{
		value:     "value",
		ttl:       0,
		createdAt: time.Now().UTC().Add(-6 * time.Second),
	}
	value, err := sc.Get("key")
	assert.Nil(t, err, "SimpleCache.Get shouldn't return an error")
	assert.EqualValues(t, value, "value", "SimpleCache.Get should return the value stored in the map when ttl < 0")
}

func TestSimpleCache_GetLastGoodCopy(t *testing.T) {
	sc := &SimpleCache{store: make(map[string]simpleCacheItem)}
	sc.store["key"] = simpleCacheItem{
		value:     "value",
		ttl:       5,
		createdAt: time.Now().UTC().Add(-6 * time.Second),
	}
	value, err := sc.GetLastGoodCopy("key")
	assert.Nil(t, err, "SimpleCache.Get shouldn't return an error")
	assert.EqualValues(t, value, "value", "SimpleCache.Get should return an error when the value is expired")
}
