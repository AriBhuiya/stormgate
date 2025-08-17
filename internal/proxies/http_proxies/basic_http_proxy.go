package http_proxies

import (
	"io"
	"log"
	"net/http"
	"time"
)

type BasicProxy struct {
	client *http.Client
}

func NewBasicProxy() *BasicProxy {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 64,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false,
	}
	return &BasicProxy{&http.Client{Transport: transport}}
}

func (b BasicProxy) Forward(w http.ResponseWriter, req *http.Request, forwardingEndpoint *string) {
	//fmt.Printf("Forwarding %s -> %s\n", req.URL, *forwardingEndpoint)

	outReq, err := http.NewRequest(req.Method, *forwardingEndpoint+req.URL.RequestURI(), req.Body)
	if err != nil {
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}

	// Copy headers from incoming request
	for k, v := range req.Header {
		outReq.Header[k] = v
	}

	resp, err := b.client.Do(outReq)
	if err != nil {
		http.Error(w, "Backend unreachable", http.StatusBadGateway)
		return
	}
	defer func(Body io.ReadCloser) {
		Body.Close()
	}(resp.Body)

	// Copy response headers
	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)

	// Stream response body
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Println(err)
		return
	}
}
