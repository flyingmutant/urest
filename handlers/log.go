package handlers

import (
	"fmt"
	"github.com/sporttech/termcolor"
	"log"
	"net/http"
	"os"
	"time"
)

type (
	loggingHandler struct {
		h http.Handler
		l *log.Logger
	}
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
	dtC := termcolor.Colorized(fmt.Sprintf("%v", dt), termcolor.CYAN)

	statusC := fmt.Sprintf("%v", tw.Status)
	if tw.Success() {
		statusC = termcolor.Colorized(statusC, termcolor.GREEN)
	} else if tw.Status >= 400 {
		statusC = termcolor.Colorized(statusC, termcolor.RED)
	}

	methodC := termcolor.Colorized(r.Method, termcolor.BLUE)
	requestURIC := termcolor.Colorized(r.RequestURI, termcolor.BOLD)

	sizeS := ""
	if tw.Size != 0 {
		sizeS = fmt.Sprintf(", %v bytes", tw.Size)
	}

	h.l.Printf("%v %v %v %v (%v%v)", statusC, r.RemoteAddr, methodC, requestURIC, dtC, sizeS)
}
