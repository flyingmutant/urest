package handlers

import (
	"net/http"
	"github.com/sporttech/urest"
)

type WithContextHandler struct {
	urest.WithContextHandler
}

func (h WithContextHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.WithContextHandler.ServeHTTP(w,r)
}

func SetRequestData(r *http.Request, name string, data interface{}) {
	urest.SetRequestData(r,name,data)
}

func GetRequestData(r *http.Request, name string) interface{} {
	return urest.GetRequestData(r, name)
}
