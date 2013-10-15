package handlers

import (
	"compress/gzip"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

type (
	StaticHandler struct {
		basePath  string
		urlPrefix string
	}
)

func NewStaticHandler(basePath string, urlPrefix string) *StaticHandler {
	return &StaticHandler{
		basePath:  basePath,
		urlPrefix: urlPrefix,
	}
}

func (h StaticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p := path.Join(h.basePath, r.URL.Path[len(h.urlPrefix):])

	if isHiddenPath(p) {
		http.NotFound(w, r)
		return
	} else if fi, err := os.Stat(p); err != nil || fi.IsDir() {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("X-Frame-Options", "SAMEORIGIN")
	if isFromMSIE(r) {
		w.Header().Set("X-UA-Compatible", "IE=edge")
	}

	if strings.Contains(p, "/vendor/") || strings.Contains(p, "/assets/") {
		year := time.Hour * 24 * 365
		w.Header().Set("Expires", time.Now().Add(year).Format(http.TimeFormat))
		w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", year/time.Second))
	} else {
		w.Header().Set("Cache-Control", "no-cache")
	}

	toGzip := false
	for _, ext := range []string{".html", ".css", ".js", ".map", ".yml", ".xml", ".json", ".txt", ".md", ".csv", ".svg"} {
		if path.Ext(p) == ext {
			toGzip = true
			break
		}
	}
	w.Header().Set("Vary", "Accept-Encoding")

	if !toGzip || !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		http.ServeFile(w, r, p)
	} else {
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		http.ServeFile(gzipResponseWriter{gz, w}, r, p)
	}
}

func isHiddenPath(p string) bool {
	ix := strings.Index(p, "/.")

	return ix != -1 && len(p) > ix+2 && p[ix+2] != '.'
}
