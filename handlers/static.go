package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type (
	staticHandler struct {
		basePath  string
		urlPrefix string
	}
)

const (
	_SVG_SIG          = "<SVG"
	_SVG_DETECT_BLOCK = 512
)

func NewStaticHandler(basePath string, urlPrefix string) http.Handler {
	checkFunc := func(r *http.Request) bool {
		exts := []string{".html", ".css", ".js", ".map", ".yml", ".xml", ".json", ".txt", ".md", ".csv", ".svg"}
		for _, ext := range exts {
			if filepath.Ext(r.URL.Path) == ext {
				return true
			}
		}
		return false
	}

	return NewGzipHandler(checkFunc, &staticHandler{
		basePath:  basePath,
		urlPrefix: urlPrefix,
	})
}

func (h *staticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p := filepath.Join(h.basePath, r.URL.Path[len(h.urlPrefix):])

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

	// ServeFile() sends 'text/xml; charset=utf-8' for SVG by default
	if filepath.Ext(p) == "" && detectSVG(p) {
		w.Header().Set("Content-Type", "image/svg+xml")
	}

	http.ServeFile(w, r, p)
}

func detectSVG(p string) bool {
	f, err := os.Open(p)
	if err != nil {
		return false
	}
	defer f.Close()

	buf := make([]byte, _SVG_DETECT_BLOCK)
	f.Read(buf)

	return strings.Contains(strings.ToUpper(string(buf)), _SVG_SIG)
}

func isHiddenPath(p string) bool {
	ix := strings.Index(p, "/.")

	return ix != -1 && len(p) > ix+2 && p[ix+2] != '.'
}
