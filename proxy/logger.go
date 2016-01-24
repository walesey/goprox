package proxy

import (
	"log"
	"net/http"
	"time"
)

type loggerData struct {
	w          http.ResponseWriter
	statusCode int
}

func (ld *loggerData) Header() http.Header {
	return ld.w.Header()
}

func (ld *loggerData) Write(data []byte) (int, error) {
	return ld.w.Write(data)
}

func (ld *loggerData) WriteHeader(statusCode int) {
	ld.w.WriteHeader(statusCode)
	ld.statusCode = statusCode
}

// Logger - writes information about http requests to the log
func Logger(next http.Handler) http.HandlerFunc {
	ld := &loggerData{}
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ld.w = w
		ld.statusCode = 200
		next.ServeHTTP(ld, r)
		method := r.Method
		statusCode := ld.statusCode
		log.Printf("[REQUEST] %v - %v - %v %v", statusCode, time.Since(start), method, r.URL.Path)
	}
}
