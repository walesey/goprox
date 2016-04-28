package cache

import (
	"net/http"
	"sync"
)

type responseHeaders struct {
	statusCode int
	headers    http.Header
}

type headerStore struct {
	entries map[string]responseHeaders
	mutex   *sync.Mutex
}

// Headers - get the stored header values
func (store *headerStore) Headers(key string) responseHeaders {
	entry, ok := store.entries[key]
	if ok {
		return entry
	}
	return responseHeaders{
		statusCode: 200,
		headers:    http.Header{},
	}
}

//StoreHeaders - store the header values in an existing cached response
func (store *headerStore) StoreHeaders(key string, headers responseHeaders) {
	store.mutex.Lock()
	store.entries[key] = headers
	store.mutex.Unlock()
}

func newHeaderStore() *headerStore {
	return &headerStore{
		entries: make(map[string]responseHeaders),
		mutex:   &sync.Mutex{},
	}
}
