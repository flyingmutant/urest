package urest

import (
	"net/http"
	"strconv"
	"strings"
)

const (
	CONTENT_TYPE_JSON = "application/json; charset=utf-8"
)

func IsSafeRequest(r *http.Request) bool {
	return r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" || r.Method == "TRACE"
}

func ReportError(w http.ResponseWriter, err error) {
	errorCodes := []int{
		http.StatusContinue,
		http.StatusSwitchingProtocols,

		http.StatusOK,
		http.StatusCreated,
		http.StatusAccepted,
		http.StatusNonAuthoritativeInfo,
		http.StatusNoContent,
		http.StatusResetContent,
		http.StatusPartialContent,

		http.StatusMultipleChoices,
		http.StatusMovedPermanently,
		http.StatusFound,
		http.StatusSeeOther,
		http.StatusNotModified,
		http.StatusUseProxy,
		http.StatusTemporaryRedirect,

		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusPaymentRequired,
		http.StatusForbidden,
		http.StatusNotFound,
		http.StatusMethodNotAllowed,
		http.StatusNotAcceptable,
		http.StatusProxyAuthRequired,
		http.StatusRequestTimeout,
		http.StatusConflict,
		http.StatusGone,
		http.StatusLengthRequired,
		http.StatusPreconditionFailed,
		http.StatusRequestEntityTooLarge,
		http.StatusRequestURITooLong,
		http.StatusUnsupportedMediaType,
		http.StatusRequestedRangeNotSatisfiable,
		http.StatusExpectationFailed,
		http.StatusTeapot,

		http.StatusInternalServerError,
		http.StatusNotImplemented,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout,
		http.StatusHTTPVersionNotSupported,
	}

	e := err.Error()
	for _, code := range errorCodes {
		if strings.HasPrefix(e, strconv.Itoa(code)+" ") {
			http.Error(w, e, code)
			return
		}
	}
	http.Error(w, e, http.StatusBadRequest)
}

func FeatureFlagPresent(r *http.Request, featureFlagHeader string, featureFlag string) bool {
	return strings.Contains(r.Header.Get(featureFlagHeader), featureFlag)
}
