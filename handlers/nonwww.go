package handlers

import (
	"net/http"
	"strings"
)

type (
	RedirectToNonWWWHandler struct {
		http.Handler
	}
)

func (h RedirectToNonWWWHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	u := *r.URL
	if u.Host == "" {
		u.Host = r.Host
	}

	if !strings.HasPrefix(u.Host, "www.") {
		h.Handler.ServeHTTP(w, r)
	} else {
		if r.TLS == nil {
			u.Scheme = "http"
		} else {
			u.Scheme = "https"
		}

		u.Host = strings.Replace(u.Host, "www.", "", 1)
		w.Header().Set("Location", u.String())
		w.WriteHeader(http.StatusMovedPermanently)
	}
}
