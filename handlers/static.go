package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

type (
	staticHandler struct {
		basePath  string
		urlPrefix string
	}
)

func NewStaticHandler(basePath string, urlPrefix string) http.Handler {
	return NewGzipHandler(true, &staticHandler{
		basePath:  basePath,
		urlPrefix: urlPrefix,
	})
}

func (h *staticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	http.ServeFile(w, r, p)
}

func isHiddenPath(p string) bool {
	ix := strings.Index(p, "/.")

	return ix != -1 && len(p) > ix+2 && p[ix+2] != '.'
}
