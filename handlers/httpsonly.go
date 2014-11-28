package handlers

import "net/http"

type (
	HTTPSRedirectHandler struct {
		http.Handler
	}
)

func (h HTTPSRedirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.TLS != nil {
		h.Handler.ServeHTTP(w, r)
	} else {
		u := *r.URL
		if u.Host == "" {
			u.Host = r.Host
		}
		u.Scheme = "https"

		w.Header().Set("Location", u.String())
		w.WriteHeader(http.StatusMovedPermanently)
	}
}
