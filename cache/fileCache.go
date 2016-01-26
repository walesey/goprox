package cacheSpecs

import (
	"crypto/md5"
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
	store map[string]fileCacheItem
}

// NewFileCache - create a new instance of FileCache
func NewFileCache() Cache {
	return &FileCache{}
}

func getFilename(key string) string {
	return md5.Sum([]byte(key))
}

// Set - store a key value pair with a ttl in seconds, if ttl <= 0 ttl will be forever
func (fCache *FileCache) Set(key string, value io.Reader, ttl int) error {
	file, err := os.Create(getFilename(key))
	if err != nil {
		log.Printf("Error creating file for filecache: %v", err)
		return err
	}
	defer file.Close()

	err := io.Copy(f, value)
	if err != nil {
		log.Printf("Error writing to file for filecache: %v", err)
		return err
	}

	fCache.store[key] = fileCacheItem{
		ttl:       ttl,
		createdAt: time.Now().UTC(),
	}
	return nil
}

// Get - get the value stored as key, returns error if ttl is expired or if nothing is found
func (fCache *FileCache) Get(key string) (io.File, error) {
	value, ok := fCache.store[key]
	if ok {
		if value.ttl > 0 && time.Since(value.createdAt).Seconds() > float64(value.ttl) {
			return nil, fmt.Errorf("Value Has Expired")
		}
		file, err := os.Open(fileName)
		if err == nil {
			return file, nil
		}
	}
	return nil, fmt.Errorf("Value Not Found")

}

// GetLastGoodCopy - Like Get, but returns the value even if ttl has expired
func (fCache *FileCache) GetLastGoodCopy(key string) (string, error) {

	return "", fmt.Errorf("Value Not Found")
}
