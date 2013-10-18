package handlers

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
)

type (
	PanicHandler struct {
		http.Handler
	}
)

func (h PanicHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			var stack []byte
			runtime.Stack(stack, false)
			log.Printf("%v: %v\n%v", colored("PANIC", _COLOR_RED), rec, string(stack))
			http.Error(w, fmt.Sprintf("Server panic: %v", rec), http.StatusInternalServerError)
		}
	}()

	h.Handler.ServeHTTP(w, r)
}
