package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var PORT = ":9001"

type EchoResponse struct {
	Port       string              `json:"port"`
	Method     string              `json:"method"`
	Path       string              `json:"path"`
	Query      map[string][]string `json:"query"`
	Headers    map[string][]string `json:"headers"`
	Cookies    map[string]string   `json:"cookies"`
	Body       string              `json:"body"`
	Proto      string              `json:"proto"`
	Host       string              `json:"host"`
	RemoteAddr string              `json:"remote_addr"`
	RequestURI string              `json:"request_uri"`
	ContentLen int64               `json:"content_length"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)

	cookies := make(map[string]string)
	for _, c := range r.Cookies() {
		cookies[c.Name] = c.Value
	}

	response := EchoResponse{
		Port:       PORT,
		Method:     r.Method,
		Path:       r.URL.Path,
		Query:      r.URL.Query(),
		Headers:    r.Header,
		Cookies:    cookies,
		Body:       string(body),
		Proto:      r.Proto,
		Host:       r.Host,
		RemoteAddr: r.RemoteAddr,
		RequestURI: r.RequestURI,
		ContentLen: r.ContentLength,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

func main() {
	port := PORT
	http.HandleFunc("/", handler)
	fmt.Printf("🚀 Echo backend listening on %s}", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println("<UNK> Echo backend listening on :9001 failed:", err)
		return
	}
}
