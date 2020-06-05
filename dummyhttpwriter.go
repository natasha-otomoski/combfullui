package main

import "net/http"

type DummyHttpWriter struct{}

func (DummyHttpWriter) Header() http.Header {
	return make(http.Header)
}
func (DummyHttpWriter) Write(b []byte) (int, error) {
	return len(b), nil
}
func (DummyHttpWriter) WriteHeader(statusCode int) {
}
