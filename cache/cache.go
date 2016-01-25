package cache

import "fmt"

type Cache interface {
	Set(key, value string, ttl int) error
	Get(key string) (string, error)
	GetLastGoodCopy(key string) (string, error)
}

type SimpleCache struct {
	store map[string]string
}

func NewSimpleCache() Cache {
	return &SimpleCache{
		store: make(map[string]string),
	}
}

func (sCache *SimpleCache) Set(key, value string, ttl int) error {

}

func (sCache *SimpleCache) Get(key string) (string, error) {
	value, ok := sCache.store[key]
	if ok {
		return value, nil
	}
	return value, fmt.Errorf("Value Not Found")
}

func (sCache *SimpleCache) GetLastGoodCopy(key string) (string, error) {

}
