package cache

import (
	"fmt"
	"sync"
	"time"
)

type memoryCacheItem struct {
	value     []byte
	ttl       int
	createdAt time.Time
}

// MemoryCache - implementation of a key value store using ram
type MemoryCache struct {
	store map[string]*memoryCacheItem
	mutex *sync.Mutex
}

// NewMemoryCache - create a new instance of FileCache
func NewMemoryCache() Cache {
	return &MemoryCache{
		store: make(map[string]*memoryCacheItem),
		mutex: &sync.Mutex{},
	}
}

// Set - store a key value pair
func (mCache *MemoryCache) Set(key string, value []byte) {
	item := &memoryCacheItem{
		value: value,
		ttl:   -1,
	}
	mCache.mutex.Lock()
	mCache.store[key] = item
	mCache.mutex.Unlock()

	mCache.Refresh(key)
}

// Get - get the value stored as key, returns error if ttl is expired or if nothing is found
func (mCache *MemoryCache) Get(key string) ([]byte, error) {
	item, ok := mCache.store[key]
	if ok {
		if item.ttl >= 0 && time.Since(item.createdAt).Seconds() > float64(item.ttl) {
			return nil, fmt.Errorf("Value Has Expired")
		}
		return item.value, nil
	}
	return nil, fmt.Errorf("Value Not Found")

}

// GetLastGoodCopy - Like Get, but returns the value even if ttl has expired
func (mCache *MemoryCache) GetLastGoodCopy(key string) ([]byte, error) {
	item, ok := mCache.store[key]
	if ok {
		return item.value, nil
	}
	return nil, fmt.Errorf("Value Not Found")
}

// Refresh - reset the expiry timer
func (mCache *MemoryCache) Refresh(key string) {
	mCache.mutex.Lock()
	item, ok := mCache.store[key]
	if ok {
		item.createdAt = time.Now().UTC()
	}
	mCache.mutex.Unlock()
}

// Expire - set the expiry time
func (mCache *MemoryCache) Expire(key string, ttl int) {
	mCache.mutex.Lock()
	item, ok := mCache.store[key]
	if ok {
		item.ttl = ttl
	}
	mCache.mutex.Unlock()
}
