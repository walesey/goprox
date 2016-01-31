package cache

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type RequestCache struct {
	defaultTTL int
	cache      Cache
	hStore     *headerStore
}

type requestCacheInterceptor struct {
	buffer     bytes.Buffer
	writer     http.ResponseWriter
	statusCode int
}

type cacheSpecs struct {
	maxage  int
	nocache bool
	private bool
}

func getCacheSpecs(cacheControl string) cacheSpecs {
	specs := cacheSpecs{}
	items := strings.Split(strings.Replace(cacheControl, " ", "", -1), ",")
	for _, item := range items {
		if strings.Contains(item, "max-age=") {
			parts := strings.Split(item, "=")
			if len(parts) == 2 {
				specs.maxage, _ = strconv.Atoi(parts[1])
			}
		} else if item == "no-cache" {
			specs.nocache = true
		} else if item == "private" {
			specs.private = true
		}
	}
	return specs
}

func statusInValidRange(statusCode int) bool {
	return (statusCode >= 200 && statusCode < 300) || (statusCode >= 400 && statusCode < 500)
}

func copyHeaders(src, dest http.Header) {
	for key, _ := range dest {
		dest.Del(key)
	}
	for key, values := range src {
		for _, value := range values {
			dest.Add(key, value)
		}
	}
}

func (rci *requestCacheInterceptor) Header() http.Header {
	return rci.writer.Header()
}

func (rci *requestCacheInterceptor) Write(data []byte) (int, error) {
	if statusInValidRange(rci.statusCode) {
		rci.buffer.Write(data)
		return rci.writer.Write(data)
	}
	return len(data), nil
}

func (rci *requestCacheInterceptor) WriteHeader(statusCode int) {
	rci.statusCode = statusCode
	if statusInValidRange(rci.statusCode) {
		rci.writer.WriteHeader(statusCode)
	}
}

// NewRequestCache - create a new instance of NewRequestCache
func NewRequestCache(cache Cache, defaultTTL int) *RequestCache {
	return &RequestCache{
		defaultTTL: defaultTTL,
		cache:      cache,
		hStore:     newHeaderStore(),
	}
}

func (rc *RequestCache) handleCaching(w http.ResponseWriter, r *http.Request, next http.Handler) {
	// etag handling
	key := fmt.Sprintf("%v%v", r.URL.Path, r.URL.RawQuery)
	storedHeaders := rc.hStore.Headers(key)
	etag := storedHeaders.headers.Get("etag")
	ifNoneMatch := r.Header.Get("If-None-Match")

	// check for cached value
	value, err := rc.cache.Get(key)
	if err == nil {
		if len(ifNoneMatch) > 0 && ifNoneMatch == etag {
			copyHeaders(storedHeaders.headers, w.Header())
			w.WriteHeader(304)
			w.Write([]byte(""))
		} else {
			// return cached value
			copyHeaders(storedHeaders.headers, w.Header())
			w.WriteHeader(storedHeaders.statusCode)
			w.Write(value)
		}
	} else {
		// no cached value - request a new value
		rci := &requestCacheInterceptor{
			buffer: bytes.Buffer{},
			writer: w,
		}
		r.Header.Del("If-None-Match")
		if len(etag) > 0 {
			r.Header.Add("If-None-Match", etag)
		}

		next.ServeHTTP(rci, r)

		if statusInValidRange(rci.statusCode) {
			headers := make(map[string][]string)
			copyHeaders(w.Header(), headers)
			rc.hStore.StoreHeaders(key, responseHeaders{
				statusCode: rci.statusCode,
				headers:    headers,
			})
			// set ttl based on cache-control headers
			cacheControl := w.Header().Get("cache-control")
			if len(cacheControl) == 0 {
				rc.cache.Expire(key, rc.defaultTTL)
			} else {
				cacheSpecs := getCacheSpecs(cacheControl)
				if cacheSpecs.nocache {
					rc.cache.Expire(key, 0)
				} else {
					rc.cache.Expire(key, cacheSpecs.maxage)
				}
			}
		}

		if rci.statusCode == 304 {
			rc.cache.Refresh(key)
			value, err := rc.cache.Get(key)
			if err == nil {
				if len(ifNoneMatch) > 0 && ifNoneMatch == etag {
					copyHeaders(storedHeaders.headers, w.Header())
					w.WriteHeader(304)
					w.Write([]byte(""))
				} else {
					// return cached value
					copyHeaders(storedHeaders.headers, w.Header())
					w.WriteHeader(storedHeaders.statusCode)
					w.Write(value)
				}
			}
		}

		if rci.statusCode >= 500 {
			// error requesting - attempt to serve last good copy
			value, err := rc.cache.GetLastGoodCopy(key)
			if err == nil {
				copyHeaders(storedHeaders.headers, w.Header())
				w.WriteHeader(storedHeaders.statusCode)
				w.Write(value)
			} else {
				w.WriteHeader(500)
				w.Write([]byte("Internal Server Error"))
			}
		}

	}
}

// Handler - Middleware for caching responses per request
func (rc *RequestCache) Handler(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rc.handleCaching(w, r, next)
	}
}
