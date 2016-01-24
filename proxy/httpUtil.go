package proxy

import (
	"io"
	"log"
	"net/http"
)

func makeHttpRequest(proxyurl string, w http.ResponseWriter, r *http.Request) error {
	client := http.Client{}
	req, err := http.NewRequest(r.Method, proxyurl, r.Body)
	if err != nil {
		log.Printf("Error Creating proxy request: %v", err)
		return err
	}

	req.Form = r.Form
	req.Header = r.Header
	req.Host = r.Host
	req.MultipartForm = r.MultipartForm
	req.PostForm = r.PostForm
	req.Proto = r.Proto
	req.ProtoMajor = r.ProtoMajor
	req.ProtoMinor = r.ProtoMinor
	req.Trailer = r.Trailer
	response, err := client.Do(req)
	if err != nil {
		log.Printf("Error requesting url:%v Error: %v", proxyurl, err)
		return err
	}

	w.WriteHeader(response.StatusCode)
	for key := range w.Header() {
		w.Header().Del(key)
	}
	for key, values := range response.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	nb, err := io.Copy(w, response.Body)
	if err != nil {
		log.Printf("Error reading body (length: %v) from response: %v", nb, err)
		return err
	}

	return nil
}
