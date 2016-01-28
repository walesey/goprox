package util

import (
	"io"
	"log"
	"net/http"
	"net/url"
)

func CopyHeaders(src, dest http.Header) {
	for key, values := range src {
		for _, value := range values {
			dest.Add(key, value)
		}
	}
}

func CopyRequest(r *http.Request) (*http.Request, error) {
	req, err := http.NewRequest(r.Method, "", r.Body)
	if err != nil {
		log.Printf("Error Creating proxy request: %v", err)
		return req, err
	}

	CopyHeaders(r.Header, req.Header)
	return req, nil
}

func ProxyHttpRequest(proxyurl string, req *http.Request, w http.ResponseWriter) error {
	client := http.Client{}
	parsedURL, err := url.Parse(proxyurl)
	if err != nil {
		log.Printf("Error Parsing URL: %v Error: %v", proxyurl, err)
		return err
	}
	req.URL = parsedURL

	response, err := client.Do(req)
	if err != nil {
		log.Printf("Error requesting url: %v Error: %v", proxyurl, err)
		return err
	}

	CopyHeaders(response.Header, w.Header())
	w.WriteHeader(response.StatusCode)
	nb, err := io.Copy(w, response.Body)
	if err != nil {
		log.Printf("Error writing the response (length: %v): %v", nb, err)
	}

	return nil
}
