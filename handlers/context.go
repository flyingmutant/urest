package handlers

import (
	"net/http"
)

type (
	WithContextHandler struct {
		http.Handler
	}
)

var (
	requestData = map[*http.Request]map[string]interface{}{}
)

func (h WithContextHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer delete(requestData, r)

	h.Handler.ServeHTTP(w, r)
}

func SetRequestData(r *http.Request, name string, data interface{}) {
	if _, ok := requestData[r]; !ok {
		requestData[r] = map[string]interface{}{}
	}
	requestData[r][name] = data
}

func GetRequestData(r *http.Request, name string) interface{} {
	return requestData[r][name]
}
