package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/walesey/goprox/proxy"
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
	Servers []ServerConfig `json:"servers"`
}

func main() {
	configFile, err := os.Open("config.json")
	if err != nil {
		log.Printf("Error opening config file: %v", err)
	}

	configJson, err := ioutil.ReadAll(configFile)
	if err != nil {
		log.Printf("Error reading config file: %v", err)
	}

	var config Config
	err = json.Unmarshal(configJson, &config)
	if err != nil {
		log.Printf("Error Unmarshaling config file: %v", err)
	}

	proxies := make([]proxy.Proxy, len(config.Servers))
	for index, serverConfig := range config.Servers {
		if serverConfig.ServerType == "loadBallancer" {
			log.Printf("Configuring loadballancer at %v", serverConfig.Path)
			proxies[index] = proxy.Proxy{
				Path:    serverConfig.Path,
				Mapping: proxy.NewLoadBallancer(serverConfig.Mappings...),
			}
		} else {
			log.Printf("Configuring proxy at %v", serverConfig.Path)
			proxies[index] = proxy.Proxy{
				Path:    serverConfig.Path,
				Mapping: proxy.DefaultMapping{serverConfig.Mapping},
			}
		}
	}

	proxy.NewProxyServer(proxies...).Listen()
}
