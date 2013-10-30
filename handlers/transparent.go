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
		Size   int
	}
)

func (w *TransparentResponseWriter) WriteHeader(status int) {
	w.Status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *TransparentResponseWriter) Write(data []byte) (int, error) {
	w.Size += len(data)
	return w.ResponseWriter.Write(data)
}

func (w *TransparentResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

func (w *TransparentResponseWriter) Success() bool {
	return w.Status == 0 || (w.Status >= 200 && w.Status < 300)
}
