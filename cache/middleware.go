package cache

import (
	"net/http"
	"strconv"
	"strings"
)

type cacheSpecs struct {
	maxage  int
	nocache bool
	private bool
}

func getCacheSpecs(cacheControl string) cacheSpecs {
	specs := cacheSpecs{}
	items := strings.Split(cacheControl, ",")
	for _, item := range items {
		if strings.Contains(item, "max-age=") {
			specs.maxage, _ = strconv.Atoi(strings.Replace(item, "=", "", 1))
		} else if item == "no-cache" {
			specs.nocache = true
		} else if item == "private" {
			specs.private = true
		}
	}
	return specs
}

type requestCacheMiddleware struct {
	sc Cache
	w  http.ResponseWriter
}

func (rcm *requestCacheMiddleware) Header() http.Header {
	return rcm.w.Header()
}

func (rcm *requestCacheMiddleware) Write(data []byte) (int, error) {
	cacheSpecs := getCacheSpecs(rcm.w.Header().Get("cache-control"))
	if !cacheSpecs.nocache {
		rcm.sc.Set("key", "value", cacheSpecs.maxage)
	}
	return rcm.w.Write(data)
}

func (rcm *requestCacheMiddleware) WriteHeader(statusCode int) {
	rcm.w.WriteHeader(statusCode)
}

// RequestCache - Middleware for caching responses per request
func RequestCache(next http.Handler) http.HandlerFunc {
	sc := NewSimpleCache()
	rcm := &requestCacheMiddleware{sc: sc}
	return func(w http.ResponseWriter, r *http.Request) {
		if sc.Get() {
			rcm.w = w
			next.ServeHTTP(rcm, r)
		}
	}
}
