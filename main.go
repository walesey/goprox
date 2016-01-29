package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/walesey/goprox/proxy"
)

func main() {
	configFile, err := os.Open("config.json")
	if err != nil {
		log.Printf("Error opening config file: %v", err)
	}

	configJson, err := ioutil.ReadAll(configFile)
	if err != nil {
		log.Printf("Error reading config file: %v", err)
	}

	var config proxy.Config
	err = json.Unmarshal(configJson, &config)
	if err != nil {
		log.Printf("Error Unmarshaling config file: %v", err)
	}

	proxy.NewProxyServer(config).Listen()
}
