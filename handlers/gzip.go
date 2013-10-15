package handlers

import (
	"compress/gzip"
	"net/http"
	"path"
	"strings"
)

type (
	GzipHandler struct {
		checkExt bool
		h        http.Handler
	}

	gzipResponseWriter struct {
		*gzip.Writer
		http.ResponseWriter
	}
)

func NewGzipHandler(checkExt bool, h http.Handler) *GzipHandler {
	return &GzipHandler{
		checkExt: checkExt,
		h:        h,
	}
}

func (h *GzipHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	toGzip := true
	if h.checkExt {
		toGzip = false
		for _, ext := range []string{".html", ".css", ".js", ".map", ".yml", ".xml", ".json", ".txt", ".md", ".csv", ".svg"} {
			if path.Ext(r.URL.Path) == ext {
				toGzip = true
				break
			}
		}
	}

	w.Header().Set("Vary", "Accept-Encoding")

	if !toGzip || !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		h.h.ServeHTTP(w, r)
	} else {
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		h.h.ServeHTTP(gzipResponseWriter{gz, w}, r)
	}
}

func (w gzipResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}
