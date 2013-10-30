package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"
)

type (
	loggingHandler struct {
		h http.Handler
		l *log.Logger
	}
)

const (
	// https://github.com/git/git/blob/master/color.h
	_COLOR_NORMAL       = ""
	_COLOR_RESET        = "\033[m"
	_COLOR_BOLD         = "\033[1m"
	_COLOR_RED          = "\033[31m"
	_COLOR_GREEN        = "\033[32m"
	_COLOR_YELLOW       = "\033[33m"
	_COLOR_BLUE         = "\033[34m"
	_COLOR_MAGENTA      = "\033[35m"
	_COLOR_CYAN         = "\033[36m"
	_COLOR_BOLD_RED     = "\033[1;31m"
	_COLOR_BOLD_GREEN   = "\033[1;32m"
	_COLOR_BOLD_YELLOW  = "\033[1;33m"
	_COLOR_BOLD_BLUE    = "\033[1;34m"
	_COLOR_BOLD_MAGENTA = "\033[1;35m"
	_COLOR_BOLD_CYAN    = "\033[1;36m"
	_COLOR_BG_RED       = "\033[41m"
	_COLOR_BG_GREEN     = "\033[42m"
	_COLOR_BG_YELLOW    = "\033[43m"
	_COLOR_BG_BLUE      = "\033[44m"
	_COLOR_BG_MAGENTA   = "\033[45m"
	_COLOR_BG_CYAN      = "\033[46m"
)

func NewLoggingHandler(h http.Handler, l *log.Logger) *loggingHandler {
	if l == nil {
		l = log.New(os.Stdout, "HTTP ", 0)
	}

	return &loggingHandler{
		h: h,
		l: l,
	}
}

func (h *loggingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "" {
		// ignore "faked" requests
		h.h.ServeHTTP(w, r)
		return
	}

	start := time.Now()
	tw := &TransparentResponseWriter{w, http.StatusOK, 0}
	defer h.log(start, tw, r)

	h.h.ServeHTTP(tw, r)
}

func (h *loggingHandler) log(start time.Time, tw *TransparentResponseWriter, r *http.Request) {
	dt := time.Now().Sub(start)
	dtC := colored(fmt.Sprintf("%v", dt), _COLOR_CYAN)

	statusC := fmt.Sprintf("%v", tw.Status)
	if tw.Success() {
		statusC = colored(statusC, _COLOR_GREEN)
	} else if tw.Status >= 400 {
		statusC = colored(statusC, _COLOR_RED)
	}

	methodC := colored(r.Method, _COLOR_BLUE)
	requestURIC := colored(r.RequestURI, _COLOR_BOLD)

	sizeS := ""
	if tw.Size != 0 {
		sizeS = fmt.Sprintf(", %v bytes", tw.Size)
	}

	h.l.Printf("%v %v %v %v (%v%v)", statusC, r.RemoteAddr, methodC, requestURIC, dtC, sizeS)
}

func colored(s string, color string) string {
	if runtime.GOOS == "windows" {
		return s
	}
	return color + s + _COLOR_RESET
}
