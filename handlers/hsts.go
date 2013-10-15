package handlers

import (
	"fmt"
	"net/http"
	"time"
)

type (
	HSTSMux struct{}
)

func (HSTSMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.TLS == nil {
		r.URL.Scheme = "https"
		if r.URL.Host == "" {
			r.URL.Host = r.Host
		}

		w.Header().Set("Location", r.URL.String())
		w.WriteHeader(http.StatusMovedPermanently)
		return
	} else if r.TLS != nil {
		year := time.Hour * 24 * 365
		w.Header().Set("Strict-Transport-Security", fmt.Sprintf("max-age=%d", year/time.Second))
	}

	http.DefaultServeMux.ServeHTTP(w, r)
}
