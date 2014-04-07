package urest

import (
	"net/http"
	"strings"
)

const (
	CONTENT_TYPE_JSON = "application/json; charset=utf-8"
)

func IsSafeRequest(r *http.Request) bool {
	return r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" || r.Method == "TRACE"
}

func FeatureFlagPresent(r *http.Request, featureFlagHeader string, featureFlag string) bool {
	return strings.Contains(r.Header.Get(featureFlagHeader), featureFlag)
}
