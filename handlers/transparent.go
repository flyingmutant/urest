package handlers

import (
	"bufio"
	"net"
	"net/http"
)

type (
	TransparentResponseWriter struct {
		http.ResponseWriter
		Status int
	}
)

func (w *TransparentResponseWriter) WriteHeader(status int) {
	w.Status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *TransparentResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

func (w *TransparentResponseWriter) Success() bool {
	return w.Status >= 200 && w.Status < 300
}
