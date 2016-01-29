package proxy

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/walesey/goprox/cache"
	"github.com/walesey/goprox/util"
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
	req, err := util.CopyRequest(r)
	if err != nil {
		log.Printf("Error copying proxy request: %v", err)
		return err
	}
	return util.ProxyHttpRequest(fmt.Sprintf("%v%v", mapping.URL, url), req, w)
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
		proxyURL := r.URL.Path[len(proxy.Path):]
		if len(r.URL.RawQuery) > 0 {
			proxyURL = fmt.Sprintf("%v?%v", proxyURL, r.URL.RawQuery)
		}
		err := proxy.Mapping.MakeRequest(proxyURL, w, r)
		if err != nil {
			log.Printf("Error making proxy request: %v", err)
			w.WriteHeader(500)
			w.Write([]byte("Proxy Error"))
		}
	}
}

// Listen - start the http server
func (server *ProxyServer) Listen() {
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "80"
	}

	router := http.NewServeMux()
	for _, proxy := range server.proxyList {
		router.HandleFunc(proxy.Path, handleProxy(proxy))
	}

	requestCache := cache.NewRequestCache(cache.NewMemoryCache())

	s := &http.Server{
		Addr:    fmt.Sprintf(":%v", port),
		Handler: LockedHandler(Logger(requestCache.Handler(router))),
	}
	log.Printf("Listening on port: %v", port)
	log.Fatal(s.ListenAndServe())
}
