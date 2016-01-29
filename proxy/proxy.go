package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func newLoadBallancerReverseProxy(route string, targets ...*url.URL) *httputil.ReverseProxy {
	index := 0
	director := func(req *http.Request) {
		if index >= len(targets) {
			index = 0
		}
		target := targets[index]
		index = index + 1
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path[len(route):])
	}
	return &httputil.ReverseProxy{
		Director: director,
	}
}

func newSingleProxy(route string, target *url.URL) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path[len(route):])
	}
	return &httputil.ReverseProxy{
		Director: director,
	}
}
