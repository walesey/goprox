package cache

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

type fileCacheItem struct {
	ttl       int
	createdAt time.Time
}

// FileCache - implementation of a key value store using the file system
type FileCache struct {
	store   map[string]*fileCacheItem
	current *os.File
}

// NewFileCache - create a new instance of FileCache
func NewFileCache() Cache {
	return &FileCache{
		store: make(map[string]*fileCacheItem),
	}
}

func getFilename(key string) string {
	hash := md5.Sum([]byte(key))
	return fmt.Sprintf("%v%v", ".cachedata/", hex.EncodeToString(hash[:]))
}

// Input - store a key value pair
func (fCache *FileCache) Input(key string) (io.Writer, io.Closer, error) {
	file, err := os.Create(getFilename(key))
	if err != nil {
		log.Printf("Error creating file for filecache: %v", err)
		return nil, nil, err
	}
	fCache.current = file

	fCache.store[key] = &fileCacheItem{
		ttl:       -1,
		createdAt: time.Now().UTC(),
	}
	return fCache.current, fCache.current, nil
}

// Output - get the value stored as key, returns error if ttl is expired or if nothing is found
func (fCache *FileCache) Output(key string) (io.Reader, io.Closer, error) {
	value, ok := fCache.store[key]
	if ok {
		if value.ttl >= 0 && time.Since(value.createdAt).Seconds() > float64(value.ttl) {
			return nil, nil, fmt.Errorf("Value Has Expired")
		}
		file, err := os.Open(getFilename(key))
		if err == nil {
			return file, file, nil
		}
	}
	return nil, nil, fmt.Errorf("Value Not Found")

}

// OutputLastGoodCopy - Like Output, but returns the value even if ttl has expired
func (fCache *FileCache) OutputLastGoodCopy(key string) (io.Reader, io.Closer, error) {
	_, ok := fCache.store[key]
	if ok {
		file, err := os.Open(getFilename(key))
		if err == nil {
			return file, file, nil
		}
	}
	return nil, nil, fmt.Errorf("Value Not Found")
}

// Expire set an expiry time for a cache entry with a ttl in seconds, if ttl < 0 ttl will be forever
func (fCache *FileCache) Expire(key string, ttl int) {
	entry, ok := fCache.store[key]
	if ok {
		entry.ttl = ttl
	}
}
