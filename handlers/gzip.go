package handlers

import (
	"compress/gzip"
	"net/http"
	"strings"
)

type (
	gzipHandler struct {
		checkFunc func(*http.Request) bool
		h         http.Handler
	}

	gzipResponseWriter struct {
		*gzip.Writer
		http.ResponseWriter
	}
)

func NewGzipHandler(checkFunc func(*http.Request) bool, h http.Handler) *gzipHandler {
	if checkFunc == nil {
		checkFunc = func(*http.Request) bool { return true }
	}

	return &gzipHandler{
		checkFunc: checkFunc,
		h:         h,
	}
}

func (h *gzipHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Vary", "Accept-Encoding")

	if !h.checkFunc(r) || !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		h.h.ServeHTTP(w, r)
	} else {
		w.Header().Set("Content-Encoding", "gzip")

		gz := gzip.NewWriter(w)
		gzrw := &gzipResponseWriter{gz, w}
		trrw := &TransparentResponseWriter{gzrw, 0, 0}

		defer func() {
			// for 304 case we should not send response body at all
			if trrw.Status != http.StatusNotModified {
				gz.Close()
			}
		}()

		h.h.ServeHTTP(trrw, r)
	}
}

func (w gzipResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}
