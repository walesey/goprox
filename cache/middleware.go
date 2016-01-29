package cache

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/walesey/goprox/util"
)

type RequestCache struct {
	DefaultTTL int
	cache      Cache
	hStore     *headerStore
	rci        *requestCacheInterceptor
}

type requestCacheInterceptor struct {
	cacheKey   string
	cache      Cache
	w          http.ResponseWriter
	writer     io.Writer
	closer     io.Closer
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

func (rci *requestCacheInterceptor) Header() http.Header {
	return rci.w.Header()
}

func (rci *requestCacheInterceptor) Write(data []byte) (int, error) {
	if rci.writer != nil {
		rci.writer.Write(data)
	}
	if statusInValidRange(rci.statusCode) {
		return rci.w.Write(data)
	}
	return len(data), nil
}

func (rci *requestCacheInterceptor) WriteHeader(statusCode int) {
	rci.statusCode = statusCode
	if statusInValidRange(rci.statusCode) {
		writer, closer, err := rci.cache.Input(rci.cacheKey)
		if err != nil {
			fmt.Printf("Error getting cache input: %v", err)
		}
		rci.writer = writer
		rci.closer = closer
	} else {
		rci.writer = nil
		rci.closer = nil
	}

	if rci.statusCode < 500 {
		rci.w.WriteHeader(statusCode)
	}
}

// NewRequestCache - create a new instance of NewRequestCache
func NewRequestCache(cache Cache) *RequestCache {
	return &RequestCache{
		DefaultTTL: 60,
		cache:      cache,
		hStore:     newHeaderStore(),
		rci:        &requestCacheInterceptor{cache: cache},
	}
}

func (rc *RequestCache) handleCaching(w http.ResponseWriter, r *http.Request, next http.Handler) {
	// etag handling
	key := fmt.Sprintf("%v%v", r.URL.Path, r.URL.RawQuery)
	headers := rc.hStore.Headers(key)
	etag := headers.headers.Get("etag")
	ifNoneMatch := r.Header.Get("If-None-Match")

	// check for cached value
	reader, closer, err := rc.cache.Output(key)
	if err == nil {
		if len(ifNoneMatch) > 0 && ifNoneMatch == etag {
			util.CopyHeaders(headers.headers, w.Header())
			w.WriteHeader(304)
			w.Write([]byte(""))
		} else {
			// return cached value
			util.CopyHeaders(headers.headers, w.Header())
			w.WriteHeader(headers.statusCode)
			io.Copy(w, reader)
			if closer != nil {
				closer.Close()
			}
		}
	} else {
		// no cached value - request a new value
		rc.rci.w = w
		rc.rci.cacheKey = key
		next.ServeHTTP(rc.rci, r)
		if rc.rci.writer != nil {
			headers := make(map[string][]string)
			util.CopyHeaders(w.Header(), headers)
			rc.hStore.StoreHeaders(key, responseHeaders{
				statusCode: rc.rci.statusCode,
				headers:    w.Header(),
			})
		}
		if rc.rci.closer != nil {
			rc.rci.closer.Close()
		}

		// set ttl based on cache-control headers
		cacheControl := w.Header().Get("cache-control")
		if rc.rci.statusCode >= 304 {
			rc.cache.Refresh(key)
		} else if len(cacheControl) == 0 {
			rc.cache.Expire(key, rc.DefaultTTL)
		} else {
			cacheSpecs := getCacheSpecs(cacheControl)
			if cacheSpecs.nocache {
				rc.cache.Expire(key, 0)
			} else {
				rc.cache.Expire(key, cacheSpecs.maxage)
			}
		}

		// error requesting - attempt to serve last good copy
		if rc.rci.statusCode >= 500 {
			reader, closer, err := rc.cache.OutputLastGoodCopy(key)
			if err == nil {
				headers := rc.hStore.Headers(key)
				util.CopyHeaders(headers.headers, w.Header())
				w.WriteHeader(headers.statusCode)
				io.Copy(w, reader)
				if closer != nil {
					closer.Close()
				}
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
