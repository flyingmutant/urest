package handlers

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"time"
)

type (
	LoggingHandler struct {
		http.Handler
	}
)

const (
	color_RESET     = "\x1b[0m"
	color_FG_RED    = "\x1b[31m"
	color_FG_GREEN  = "\x1b[32m"
	color_FG_YELLOW = "\x1b[33m"
	color_FG_CYAN   = "\x1b[36m"
)

func (h LoggingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "" {
		// ignore "faked" requests
		h.Handler.ServeHTTP(w, r)
		return
	}

	start := time.Now()
	tw := &TransparentResponseWriter{w, http.StatusOK, 0}

	h.Handler.ServeHTTP(tw, r)

	h.log(start, tw, r)
}

func (h LoggingHandler) log(start time.Time, tw *TransparentResponseWriter, r *http.Request) {
	dt := time.Now().Sub(start)
	dtC := colored(fmt.Sprintf("%v", dt), color_FG_CYAN)

	statusC := fmt.Sprintf("%v", tw.Status)
	if tw.Success() {
		statusC = colored(statusC, color_FG_GREEN)
	} else if tw.Status >= 400 {
		statusC = colored(statusC, color_FG_RED)
	}

	methodC := colored(r.Method, color_FG_YELLOW)

	sizeS := ""
	if tw.Size != 0 {
		sizeS = fmt.Sprintf(", %v bytes", tw.Size)
	}

	log.Printf("[%v] %v %v (%v%v)", statusC, methodC, r.RequestURI, dtC, sizeS)
}

func colored(s string, color string) string {
	if runtime.GOOS == "windows" {
		return s
	}
	return color + s + color_RESET
}
