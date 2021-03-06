package cache

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type RequestCache struct {
	defaultTTL, maxTTL int
	cache              Cache
	hStore             *headerStore
}

type requestCacheInterceptor struct {
	buffer     bytes.Buffer
	header     http.Header
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

func minInt(i1, i2 int) int {
	if i1 < i2 {
		return i1
	}
	return i2
}

func (rci *requestCacheInterceptor) Header() http.Header {
	return rci.header
}

func (rci *requestCacheInterceptor) Write(data []byte) (int, error) {
	if statusInValidRange(rci.statusCode) {
		return rci.buffer.Write(data)
	}
	return len(data), nil
}

func (rci *requestCacheInterceptor) WriteHeader(statusCode int) {
	rci.statusCode = statusCode
}

// NewRequestCache - create a new instance of NewRequestCache
func NewRequestCache(cache Cache, defaultTTL, maxTTL int) *RequestCache {
	return &RequestCache{
		defaultTTL: defaultTTL,
		maxTTL:     maxTTL,
		cache:      cache,
		hStore:     newHeaderStore(),
	}
}

func (rc *RequestCache) handleCaching(w http.ResponseWriter, r *http.Request, next http.Handler) {
	// etag handling
	key := fmt.Sprintf("%v%v", r.URL.Path, r.URL.RawQuery)
	storedHeaders := rc.hStore.Headers(key)
	header := storedHeaders.headers
	statusCode := storedHeaders.statusCode
	etag := header.Get("etag")
	ifNoneMatch := r.Header.Get("If-None-Match")
	r.Header.Del("If-None-Match")

	// check for cached value
	value, err := rc.cache.Get(key)
	if err != nil {
		// no cached value - request a new value
		header = make(map[string][]string)
		rci := &requestCacheInterceptor{
			buffer: bytes.Buffer{},
			header: header,
		}
		if len(etag) > 0 {
			r.Header.Add("If-None-Match", etag)
		}

		//make a request to the server
		next.ServeHTTP(rci, r)

		//check if response is cacheable
		if statusInValidRange(rci.statusCode) {
			value = rci.buffer.Bytes()
			statusCode = rci.statusCode
			// add etag if none exists
			if len(header.Get("etag")) == 0 {
				header.Set("etag", fmt.Sprintf("W/\"%x\"", md5.Sum(value)))
			}
			rc.hStore.StoreHeaders(key, responseHeaders{
				statusCode: rci.statusCode,
				headers:    header,
			})
			rc.cache.Set(key, value)
			// set ttl based on cache-control headers
			cacheControl := header.Get("cache-control")
			if len(cacheControl) == 0 {
				rc.cache.Expire(key, rc.defaultTTL)
			} else {
				cacheSpecs := getCacheSpecs(cacheControl)
				if cacheSpecs.nocache {
					rc.cache.Expire(key, 0)
				} else {
					rc.cache.Expire(key, minInt(cacheSpecs.maxage, rc.maxTTL))
				}
			}
		}

		if rci.statusCode == 304 || rci.statusCode >= 500 {
			// error requesting - attempt to serve last good copy
			rc.cache.Refresh(key)
			value, err = rc.cache.GetLastGoodCopy(key)
			if err != nil {
				w.WriteHeader(500)
				w.Write([]byte("Internal Server Error"))
				return
			}
		}
	}

	copyHeaders(header, w.Header())
	if len(ifNoneMatch) > 0 && ifNoneMatch == etag {
		w.WriteHeader(304)
		w.Write([]byte(""))
	} else {
		// return cached value
		w.WriteHeader(statusCode)
		w.Write(value)
	}
}

// Handler - Middleware for caching responses per request
func (rc *RequestCache) Handler(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rc.handleCaching(w, r, next)
	}
}
