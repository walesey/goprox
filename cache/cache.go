package cache

import (
	"fmt"
	"time"
)

// Cache - interface for a key value store
type Cache interface {
	Set(key, value string, ttl int) error
	Get(key string) (string, error)
	GetLastGoodCopy(key string) (string, error)
}

type simpleCacheItem struct {
	value     string
	ttl       int
	createdAt time.Time
}

// SimpleCache - implementation of a key value store using a go map
type SimpleCache struct {
	store map[string]simpleCacheItem
}

// NewSimpleCache - create a new instance of SimpleCache
func NewSimpleCache() Cache {
	return &SimpleCache{
		store: make(map[string]simpleCacheItem),
	}
}

// Set - store a key value pair with a ttl in seconds, if ttl <= 0 ttl will be forever
func (sCache *SimpleCache) Set(key, value string, ttl int) error {
	sCache.store[key] = simpleCacheItem{
		value:     value,
		ttl:       ttl,
		createdAt: time.Now().UTC(),
	}
	return nil
}

// Get - get the value stored as key, returns error if ttl is expired or if nothing is found
func (sCache *SimpleCache) Get(key string) (string, error) {
	value, ok := sCache.store[key]
	if ok {
		if value.ttl > 0 && time.Since(value.createdAt).Seconds() > float64(value.ttl) {
			return "", fmt.Errorf("Value Has Expired")
		}
		return value.value, nil
	}
	return "", fmt.Errorf("Value Not Found")
}

// GetLastGoodCopy - Like Get, but returns the value even if ttl has expired
func (sCache *SimpleCache) GetLastGoodCopy(key string) (string, error) {
	value, ok := sCache.store[key]
	if ok {
		return value.value, nil
	}
	return "", fmt.Errorf("Value Not Found")
}
