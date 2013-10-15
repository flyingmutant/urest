package handlers

import (
	"net/http"
)

type (
	CollectingResponseWriter struct {
		Collect bool
		Data    []byte
	}
)

func (w *CollectingResponseWriter) Header() http.Header {
	return http.Header{}
}

func (w *CollectingResponseWriter) WriteHeader(int) {}

func (w *CollectingResponseWriter) Write(data []byte) (int, error) {
	if w.Collect {
		w.Data = append(w.Data, data...)
	}
	return len(data), nil
}
