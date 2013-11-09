package handlers

import (
	"fmt"
	"github.com/sporttech/termcolor"
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
			stack := make([]byte, 1024*32)
			runtime.Stack(stack, false)
			log.Printf("%v: %v\n%v", termcolor.Colorized("PANIC", termcolor.RED), rec, string(stack))
			http.Error(w, fmt.Sprintf("Server panic: %v", rec), http.StatusInternalServerError)
		}
	}()

	h.Handler.ServeHTTP(w, r)
}
