package proxy

import (
	"crypto/tls"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func init() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
}

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

func transformUrls(target, dest *url.URL) {
	targetQuery := target.RawQuery
	dest.Scheme = target.Scheme
	dest.Host = target.Host
	dest.Path = singleJoiningSlash(target.Path, dest.Path)
	if targetQuery == "" || dest.RawQuery == "" {
		dest.RawQuery = targetQuery + dest.RawQuery
	} else {
		dest.RawQuery = targetQuery + "&" + dest.RawQuery
	}
}

func newLoadBallancerReverseProxy(route string, targets ...*url.URL) *httputil.ReverseProxy {
	index := 0
	director := func(req *http.Request) {
		if index >= len(targets) {
			index = 0
		}
		target := targets[index]
		index = index + 1
		req.URL.Path = req.URL.Path[len(route):]
		transformUrls(target, req.URL)
	}
	return &httputil.ReverseProxy{
		Director: director,
	}
}

func newSingleProxy(route string, target *url.URL) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		req.URL.Path = req.URL.Path[len(route):]
		transformUrls(target, req.URL)
	}
	return &httputil.ReverseProxy{
		Director: director,
	}
}
