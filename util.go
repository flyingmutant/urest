package urest

import (
	"net/http"
)

const (
	CONTENT_TYPE_JSON = "application/json; charset=utf-8"
)

func IsSafeRequest(r *http.Request) bool {
	return r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" || r.Method == "TRACE"
}
