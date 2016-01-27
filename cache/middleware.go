package cache

import (
	"fmt"
	"io"
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

type requestCacheMiddleware struct {
	c          Cache
	cacheKey   string
	w          http.ResponseWriter
	writer     io.Writer
	closer     io.Closer
	statusCode int
}

func statusInValidRange(statusCode int) bool {
	return (statusCode >= 200 && statusCode < 300) || (statusCode >= 400 && statusCode < 500)
}

func (rcm *requestCacheMiddleware) Header() http.Header {
	return rcm.w.Header()
}

func (rcm *requestCacheMiddleware) Write(data []byte) (int, error) {
	if rcm.writer != nil {
		rcm.writer.Write(data)
	}
	if rcm.statusCode < 500 {
		return rcm.w.Write(data)
	}
	return len(data), nil
}

func (rcm *requestCacheMiddleware) WriteHeader(statusCode int) {
	rcm.statusCode = statusCode
	if statusInValidRange(rcm.statusCode) {
		writer, closer, err := rcm.c.Input(rcm.cacheKey)
		if err != nil {
			fmt.Printf("Error geting cache input: %v", err)
		}
		rcm.writer = writer
		rcm.closer = closer
	} else {
		rcm.writer = nil
		rcm.closer = nil
	}

	if rcm.statusCode < 500 {
		rcm.w.WriteHeader(statusCode)
	}
}

// RequestCache - Middleware for caching responses per request
func RequestCache(next http.Handler) http.HandlerFunc {
	fc := NewFileCache()
	rcm := &requestCacheMiddleware{c: fc}
	return func(w http.ResponseWriter, r *http.Request) {
		key := fmt.Sprintf("%v%v", r.URL.RawPath, r.URL.RawQuery)
		reader, closer, err := fc.Output(key)
		if err == nil {
			w.WriteHeader(200)
			io.Copy(w, reader)
			closer.Close()
		} else {
			rcm.w = w
			rcm.cacheKey = key
			next.ServeHTTP(rcm, r)
			if rcm.closer != nil {
				rcm.closer.Close()
			}
			cacheSpecs := getCacheSpecs(w.Header().Get("cache-control"))
			if cacheSpecs.nocache {
				fc.Expire(key, 0)
			} else {
				fc.Expire(key, cacheSpecs.maxage)
			}
			if rcm.statusCode >= 500 {
				reader, closer, err := fc.OutputLastGoodCopy(key)
				if err == nil {
					w.WriteHeader(200)
					io.Copy(w, reader)
					closer.Close()
				} else {
					w.WriteHeader(500)
					w.Write([]byte("Internal Server Error"))
				}
			}
		}
	}
}
