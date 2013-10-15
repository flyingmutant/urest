package handlers

import (
	"compress/gzip"
	"net/http"
)

type (
	GzipResponseWriter struct {
		*gzip.Writer
		http.ResponseWriter
	}
)

func (w GzipResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w GzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}
