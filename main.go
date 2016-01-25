package main

import "github.com/walesey/goprox/proxy"

func main() {
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
