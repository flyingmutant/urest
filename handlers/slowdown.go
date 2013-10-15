package handlers

import (
	"math/rand"
	"net/http"
	"time"
)

type (
	SlowdownHandler struct {
		MaxDelayMS int
		http.Handler
	}
)

func (h SlowdownHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.MaxDelayMS > 0 {
		time.Sleep(time.Duration(rand.Intn(h.MaxDelayMS)) * time.Millisecond)
	}

	h.Handler.ServeHTTP(w, r)
}
