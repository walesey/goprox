package proxy

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

// Mapping - maps to the next url
type Mapping interface {
	MakeRequest(url string, w http.ResponseWriter, r *http.Request) error
}

// DefaultMapping - simple mapping without state
type DefaultMapping struct {
	URL string
}

// MakeRequest - simple mapping without state
func (mapping DefaultMapping) MakeRequest(url string, w http.ResponseWriter, r *http.Request) error {
	return makeHttpRequest(fmt.Sprintf("%v%v", mapping.URL, url), w, r)
}

// Proxy - Map a path to a proxymapping
type Proxy struct {
	Path    string
	Mapping Mapping
}

// ProxyServer - a server that passes all request to another server
type ProxyServer struct {
	proxyList []Proxy
}

// NewProxyServer - Create new instance of ProxyServer
func NewProxyServer(proxyList ...Proxy) *ProxyServer {
	return &ProxyServer{
		proxyList: proxyList,
	}
}

func handleProxy(proxy Proxy) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		proxyURL := fmt.Sprintf("%v?%v", r.URL.Path[len(proxy.Path):], r.URL.RawQuery)
		err := proxy.Mapping.MakeRequest(proxyURL, w, r)
		if err != nil {
			log.Printf("Error making proxy request: %v", err)
			w.WriteHeader(500)
			w.Write([]byte("Proxy Error"))
		}
	}
}

func (server *ProxyServer) Listen() {
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "80"
	}

	router := http.NewServeMux()
	for _, proxy := range server.proxyList {
		router.HandleFunc(proxy.Path, handleProxy(proxy))
	}

	s := &http.Server{
		Addr:    fmt.Sprintf(":%v", port),
		Handler: Logger(router),
	}
	log.Printf("Listening on port: %v", port)
	log.Fatal(s.ListenAndServe())
}
