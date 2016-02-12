package proxy

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/walesey/goprox/middleware"

	"github.com/walesey/goprox/cache"
)

// ServerConfig - config for a singe proxy server
type ServerConfig struct {
	ServerType string   `json:"serverType"`
	Path       string   `json:"path"`
	Mapping    string   `json:"mapping"`
	Mappings   []string `json:"mappings"`
}

// Config - main config
type Config struct {
	Servers    []ServerConfig `json:"servers"`
	CacheType  string         `json:"cacheType"`
	DefaultTTL int            `json:"defaultTTL"`
	ColorLogs  bool           `json:"colorLogs"`
}

// ProxyServer - main server with config
type ProxyServer struct {
	config Config
}

// NewProxyServer - Create new instance of ProxyServer
func NewProxyServer(config Config) *ProxyServer {
	return &ProxyServer{
		config: config,
	}
}

func parseURL(urlString string) *url.URL {
	result, err := url.Parse(urlString)
	if err != nil {
		log.Printf("Error parsing mapping to url: %v", err)
	}
	return result
}

// Listen - start the http server
func (server *ProxyServer) Listen() {
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "80"
	}

	router := http.NewServeMux()
	for _, serverConfig := range server.config.Servers {
		if serverConfig.ServerType == "loadBallancer" {
			urls := make([]*url.URL, len(serverConfig.Mappings))
			for index, mapping := range serverConfig.Mappings {
				urls[index] = parseURL(mapping)
			}
			router.Handle(serverConfig.Path, newLoadBallancerReverseProxy(serverConfig.Path, urls...))
		} else {
			router.Handle(serverConfig.Path, newSingleProxy(serverConfig.Path, parseURL(serverConfig.Mapping)))
		}
	}

	var c cache.Cache
	if server.config.CacheType == "fileSystem" {
		// c = cache.NewFileCache()
	} else {
		c = cache.NewMemoryCache()
	}
	requestCache := cache.NewRequestCache(c, server.config.DefaultTTL)

	loggerConf := middleware.LoggerConfig{
		Output:      os.Stdout,
		EnableColor: server.config.ColorLogs,
	}

	s := &http.Server{
		Addr:    fmt.Sprintf(":%v", port),
		Handler: middleware.Logger(loggerConf, requestCache.Handler(router)),
	}
	log.Printf("Listening on port: %v", port)
	log.Fatal(s.ListenAndServe())
}
