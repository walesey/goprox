package cache

import (
	"bytes"
	"fmt"
	"io"
	"time"
)

type memoryCacheItem struct {
	buffer    *bytes.Buffer
	data      []byte
	ttl       int
	createdAt time.Time
}

func (item *memoryCacheItem) Close() error {
	item.data = item.buffer.Bytes()
	item.buffer = nil
	return nil
}

// MemoryCache - implementation of a key value store using ram
type MemoryCache struct {
	store map[string]*memoryCacheItem
}

// NewMemoryCache - create a new instance of FileCache
func NewMemoryCache() Cache {
	return &MemoryCache{
		store: make(map[string]*memoryCacheItem),
	}
}

// Input - store a key value pair
func (mCache *MemoryCache) Input(key string) (io.Writer, io.Closer, error) {
	buffer := bytes.NewBuffer([]byte{})
	item := &memoryCacheItem{
		buffer:    buffer,
		ttl:       -1,
		createdAt: time.Now().UTC(),
	}
	mCache.store[key] = item
	return buffer, item, nil
}

// Output - get the value stored as key, returns error if ttl is expired or if nothing is found
func (mCache *MemoryCache) Output(key string) (io.Reader, io.Closer, error) {
	item, ok := mCache.store[key]
	if ok {
		if item.ttl >= 0 && time.Since(item.createdAt).Seconds() > float64(item.ttl) {
			return nil, nil, fmt.Errorf("Value Has Expired")
		}
		return bytes.NewBuffer(item.data), nil, nil
	}
	return nil, nil, fmt.Errorf("Value Not Found")

}

// OutputLastGoodCopy - Like Output, but returns the value even if ttl has expired
func (mCache *MemoryCache) OutputLastGoodCopy(key string) (io.Reader, io.Closer, error) {
	item, ok := mCache.store[key]
	if ok {
		return bytes.NewBuffer(item.data), nil, nil
	}
	return nil, nil, fmt.Errorf("Value Not Found")
}

// Expire set an expiry time for a cache entry with a ttl in seconds, if ttl < 0 ttl will be forever
func (mCache *MemoryCache) Expire(key string, ttl int) {
	item, ok := mCache.store[key]
	if ok {
		item.ttl = ttl
	}
}
