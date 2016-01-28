package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/walesey/goprox/proxy"
)

type ServerConfig struct {
}

type Config struct {
	servers []ServerConfig
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
	err = json.Unmarshal(configJson, config)
	if err != nil {
		log.Printf("Error Unmarshaling config file: %v", err)
	}

	proxy.NewProxyServer(
		proxy.Proxy{
			Path:    "/server1/",
			Mapping: proxy.DefaultMapping{"http://localhost:3000/"},
		},
		proxy.Proxy{
			Path: "/server2/",
			Mapping: proxy.NewLoadBallancer(
				"http://localhost:3000/",
				"http://localhost:3001/",
				"http://localhost:3002/",
			),
		},
	).Listen()
}
