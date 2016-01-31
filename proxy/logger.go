package proxy

import (
	"log"
	"net/http"
	"time"
)

type loggerInterptor struct {
	w          http.ResponseWriter
	statusCode int
}

func (li *loggerInterptor) Header() http.Header {
	return li.w.Header()
}

func (li *loggerInterptor) Write(data []byte) (int, error) {
	return li.w.Write(data)
}

func (li *loggerInterptor) WriteHeader(statusCode int) {
	li.w.WriteHeader(statusCode)
	li.statusCode = statusCode
}

// Logger - writes information about http requests to the log
func Logger(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		path := r.URL.Path
		li := &loggerInterptor{w: w}
		next.ServeHTTP(li, r)
		method := r.Method
		statusCode := li.statusCode
		log.Printf("[REQUEST] %v - %v - %v %v", statusCode, time.Since(start), method, path)
	}
}
