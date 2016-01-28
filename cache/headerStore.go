package cache

import "net/http"

type responseHeaders struct {
	statusCode int
	headers    http.Header
}

type headerStore struct {
	entries map[string]responseHeaders
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
	store.entries[key] = headers
}

func newHeaderStore() *headerStore {
	return &headerStore{entries: make(map[string]responseHeaders)}
}
