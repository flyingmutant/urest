package urest

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

const (
	CONTENT_TYPE_JSON = "application/json; charset=utf-8"

	SERVER = "uREST/0.2"

	t_RESET     = "\x1b[0m"
	t_FG_RED    = "\x1b[31m"
	t_FG_GREEN  = "\x1b[32m"
	t_FG_YELLOW = "\x1b[33m"
	t_FG_CYAN   = "\x1b[36m"
)

var (
	requestData = map[*http.Request]map[string]interface{}{}
)

type (
	Resource interface {
		Parent() Resource
		PathSegment() string
		Child(string, *http.Request) Resource

		AllowedMethods() []string
		AllowedActions() []string

		ETag(*http.Request) string
		Expires() time.Time
		CacheControl() string
		ContentType() string

		Read(urlPrefix string, w http.ResponseWriter, r *http.Request)
		Update(*http.Request) error
		Do(action string, r *http.Request) error

		IsCollection() bool
	}

	Collection interface {
		Resource

		Create(*http.Request) (Resource, error)
		Delete(string, *http.Request) error
	}

	Handler struct {
		res    Resource
		prefix string
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		r      *http.Request
		status int
		start  time.Time
		size   int
	}
)

func SetRequestData(r *http.Request, name string, data interface{}) {
	if _, ok := requestData[r]; !ok {
		requestData[r] = map[string]interface{}{}
	}
	requestData[r][name] = data
}

func GetRequestData(r *http.Request, name string) interface{} {
	return requestData[r][name]
}

func (lrw *loggingResponseWriter) WriteHeader(status int) {
	lrw.ResponseWriter.WriteHeader(status)
	lrw.status = status
}

func (lrw *loggingResponseWriter) Write(data []byte) (int, error) {
	lrw.size += len(data)
	return lrw.ResponseWriter.Write(data)
}

func (lrw *loggingResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return lrw.ResponseWriter.(http.Hijacker).Hijack()
}

func (lrw *loggingResponseWriter) success() bool {
	return lrw.status >= 200 && lrw.status < 300
}

func (lrw *loggingResponseWriter) log() {
	d := time.Now().Sub(lrw.start)
	dC := tColor(fmt.Sprintf("%v", d), t_FG_CYAN)

	statusC := fmt.Sprintf("%v", lrw.status)
	if lrw.success() {
		statusC = tColor(statusC, t_FG_GREEN)
	} else if lrw.status >= 400 {
		statusC = tColor(fmt.Sprintf("%v", lrw.status), t_FG_RED)
	}

	methodC := tColor(lrw.r.Method, t_FG_YELLOW)

	sizeS := ""
	if lrw.size != 0 {
		sizeS = fmt.Sprintf(", %v bytes", lrw.size)
	}

	uri := lrw.r.RequestURI
	if uri == "" {
		uri = "@" + lrw.r.URL.String()
	}

	log.Printf("[%v] %v %v (%v%v)", statusC, methodC, uri, dC, sizeS)
}

func tColor(s string, color string) string {
	return color + s + t_RESET
}

func NewHandler(res Resource, prefix string) *Handler {
	if !strings.HasPrefix(prefix, "/") || !strings.HasSuffix(prefix, "/") {
		panic(fmt.Sprintf("Invalid prefix '%v'", prefix))
	}

	return &Handler{res, prefix}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("%v: %v", tColor("PANIC", t_FG_RED), rec)
			debug.PrintStack()
			http.Error(w, "Server panic", http.StatusInternalServerError)
		}
	}()

	lrw := &loggingResponseWriter{w, r, http.StatusOK, time.Now(), 0}
	defer lrw.log()

	w.Header().Set("Server", fmt.Sprintf("%v (%v %v)", SERVER, runtime.GOOS, runtime.GOARCH))

	defer delete(requestData, r)

	h.handle(lrw, r)
}

func (h *Handler) handle(w http.ResponseWriter, r *http.Request) {
	if h.res.Parent() != nil {
		panic(fmt.Sprintf("Resource '%v' is not a root of the resource tree", relativeURL(h.res)))
	}

	if r.URL.Path[:len(h.prefix)] != h.prefix {
		panic(fmt.Sprintf("Prefix '%v' does not match request URL path '%v'", h.prefix, r.URL.Path))
	}

	steps := strings.Split(r.URL.Path[len(h.prefix):], "/")
	if len(steps) == 1 && steps[0] == "" {
		steps = []string{}
	}

	ch, rest := navigate(h.res, steps, r)
	if ch == nil {
		if rest == nil {
			u := *r.URL
			u.Path += "/"
			w.Header().Set("Location", u.String())
			w.WriteHeader(http.StatusMovedPermanently)
			return
		}

		if len(rest) == 1 && rest[0] == "" {
			u := *r.URL
			u.Path = u.Path[:len(u.Path)-1]
			w.Header().Set("Location", u.String())
			w.WriteHeader(http.StatusMovedPermanently)
			return
		}

		http.NotFound(w, r)
		return
	}

	postAction := (*string)(nil)
	if len(rest) > 0 {
		postAction = &rest[0]
	}

	if r.Method != "POST" && postAction != nil {
		http.NotFound(w, r)
		return
	}

	handle(ch, postAction, h.prefix, w, r)
}

func navigate(res Resource, steps []string, r *http.Request) (Resource, []string) {
	if len(steps) == 0 {
		if res.IsCollection() {
			// collection URL without a trailing '/'
			return nil, nil
		} else {
			return res, []string{}
		}
	}

	head := steps[0]
	rest := steps[1:]

	if head == "" {
		if len(rest) != 0 {
			panic("Empty non-last step during navigation")
		}

		if res.IsCollection() {
			// collection URL with a trailing '/'
			return res, []string{}
		} else {
			return nil, []string{""}
		}
	}

	if ch := res.Child(head, r); ch != nil {
		if ch.PathSegment() != head {
			panic(fmt.Sprintf("Resource '%v' has wrong path segment ('%v' / '%v')", relativeURL(ch), ch.PathSegment(), head))
		}
		return navigate(ch, rest, r)
	}

	if len(rest) != 0 {
		return nil, steps
	}

	// custom POST action
	return res, []string{head}
}

func handle(res Resource, postAction *string, prefix string, w http.ResponseWriter, r *http.Request) {
	if index(res.AllowedMethods(), r.Method) == -1 {
		w.Header().Set("Allow", strings.Join(res.AllowedMethods(), ", "))
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	switch r.Method {
	case "HEAD":
		setHeaders(res, w, r)
		w.Header().Set("Allow", strings.Join(res.AllowedMethods(), ", "))
		w.WriteHeader(http.StatusOK)
	case "GET":
		setHeaders(res, w, r)
		if et := etag(res, r); et != "" {
			if r.Header.Get("If-None-Match") == et {
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}

		res.Read(prefix, w, r)
	case "POST":
		if postAction != nil {
			if index(res.AllowedActions(), *postAction) == -1 {
				http.Error(w, "Unknown action", http.StatusBadRequest)
				return
			}

			if e := res.Do(*postAction, r); e != nil {
				reportError(w, e)
			} else {
				w.WriteHeader(http.StatusNoContent)
			}
		} else {
			if !res.IsCollection() {
				http.Error(w, "Not a collection", http.StatusBadRequest)
				return
			}

			if ch, e := res.(Collection).Create(r); e != nil {
				reportError(w, e)
			} else {
				w.Header().Set("Location", RelativeURL(prefix, ch).String())
				w.WriteHeader(http.StatusCreated)
			}
		}
	case "PATCH":
		if e := res.Update(r); e != nil {
			reportError(w, e)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	case "DELETE":
		if res.Parent() == nil {
			http.Error(w, "No parent", http.StatusBadRequest)
			return
		}

		if !res.Parent().IsCollection() {
			http.Error(w, "Parent is not a collection", http.StatusBadRequest)
			return
		}

		if e := res.Parent().(Collection).Delete(res.PathSegment(), r); e != nil {
			reportError(w, e)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

func setHeaders(res Resource, w http.ResponseWriter, r *http.Request) {
	if ct := res.ContentType(); ct != "" {
		w.Header().Set("Content-Type", ct)
	}
	if t := res.Expires(); !t.IsZero() {
		w.Header().Set("Expires", t.Format(time.RFC1123))
	}
	cc := res.CacheControl()
	if cc != "" {
		w.Header().Set("Cache-Control", cc)
	}
	if et := etag(res, r); et != "" {
		w.Header().Set("ETag", et)
		if cc == "" {
			w.Header().Set("Cache-Control", CacheControl(0))
		}
	} else {
		if cc == "" {
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		}
	}
}

func etag(res Resource, r *http.Request) string {
	for res != nil {
		if et := res.ETag(r); et != "" {
			return et
		}
		res = res.Parent()
	}
	return ""
}

func reportError(w http.ResponseWriter, err error) {
	errorCodes := []int{
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
		if http.StatusText(code) == e {
			http.Error(w, fmt.Sprintf("%d %s", code, e), code)
			return
		}
	}
	http.Error(w, e, http.StatusBadRequest)
}

func AbsoluteURL(r *http.Request, prefix string, res Resource) *url.URL {
	au := *r.URL
	if au.Host == "" {
		au.Host = r.Host
	}

	if r.TLS == nil {
		au.Scheme = "http"
	} else {
		au.Scheme = "https"
	}

	u := RelativeURL(prefix, res)

	return au.ResolveReference(u)
}

func RelativeURL(prefix string, res Resource) *url.URL {
	u := relativeURL(res)
	u.Path = prefix[:len(prefix)-1] + u.Path
	return u
}

func relativeURL(res Resource) *url.URL {
	parts := make([]string, 0)
	isColl := res.IsCollection()

	for res != nil {
		parts = append(parts, res.PathSegment())
		res = res.Parent()
	}

	// I love Go
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}

	path := strings.Join(parts, "/")
	if isColl {
		path = path + "/"
	}

	return &url.URL{Path: path}
}

func index(arr []string, s string) int {
	for i, e := range arr {
		if e == s {
			return i
		}
	}
	return -1
}
