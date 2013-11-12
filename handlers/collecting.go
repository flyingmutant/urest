package handlers

import (
	"net/http"
)

type (
	CollectingResponseWriter struct {
		Collect bool
		Header_ http.Header
		Data    []byte
	}
)

func (w *CollectingResponseWriter) Header() http.Header {
	if w.Header_ != nil {
		return w.Header_
	} else {
		return http.Header{}
	}
}

func (w *CollectingResponseWriter) WriteHeader(int) {}

func (w *CollectingResponseWriter) Write(data []byte) (int, error) {
	if w.Collect {
		w.Data = append(w.Data, data...)
	}
	return len(data), nil
}
