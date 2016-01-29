package proxy

import (
	"net/http"
	"sync"
)

// LockedHandler - ensures single threaded http handling
func LockedHandler(next http.Handler) http.HandlerFunc {
	mutex := sync.Mutex{}
	return func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		next.ServeHTTP(w, r)
		mutex.Unlock()
	}
}
