package urest

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"
	"time"
)

const (
	SERVER = "uREST/0.1"

	t_RESET     = "\x1b[0m"
	t_FG_RED    = "\x1b[31m"
	t_FG_GREEN  = "\x1b[32m"
	t_FG_YELLOW = "\x1b[33m"
	t_FG_CYAN   = "\x1b[36m"
)

func tColor(s string, color string) string {
	return color + s + t_RESET
}

type Resource interface {
	Parent() Resource
	PathSegment() string
	Child(string) Resource

	AllowedMethods() []string
	AllowedActions() []string

	ETag() string
	Expires() time.Time
	CacheControl() string
	ContentType() string

	Get(urlPrefix string, r *http.Request) ([]byte, error)
	Patch(*http.Request) error
	Do(action string, r *http.Request) error
}

type Collection interface {
	Resource

	Create(*http.Request) (Resource, error)
	Remove(string) error
}

type loggingResponseWriter struct {
	http.ResponseWriter
	r      *http.Request
	status int
	start  time.Time
}

func (lrw *loggingResponseWriter) WriteHeader(status int) {
	lrw.ResponseWriter.WriteHeader(status)
	lrw.status = status
}

func (lrw *loggingResponseWriter) log() {
	d := time.Now().Sub(lrw.start)
	dC := tColor(fmt.Sprintf("%v", d), t_FG_CYAN)

	statusC := tColor(fmt.Sprintf("%v", lrw.status), t_FG_GREEN)
	if lrw.status >= 400 {
		statusC = tColor(fmt.Sprintf("%v", lrw.status), t_FG_RED)
	}

	methodC := tColor(lrw.r.Method, t_FG_YELLOW)

	log.Printf("[%v] %v %v (%v)", statusC, methodC, lrw.r.RequestURI, dC)
}

func HandlerWithPrefix(res Resource, prefix string) func(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(prefix, "/") || !strings.HasSuffix(prefix, "/") {
		panic(fmt.Sprintf("Invalid prefix '%v'", prefix))
	}

	return func(w http.ResponseWriter, r *http.Request) {
		lrw := &loggingResponseWriter{w, r, http.StatusInternalServerError, time.Now()}

		w.Header().Set("Server", SERVER)

		defer func() {
			lrw.log()
			if rec := recover(); rec != nil {
				log.Printf("Panic: %v", rec)
				debug.PrintStack()
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		handleWithPrefix(res, prefix, lrw, r)
	}
}

func handleWithPrefix(res Resource, prefix string, w http.ResponseWriter, r *http.Request) {
	if res.Parent() != nil {
		panic(fmt.Sprintf("Resource '%v' is not a root of the resource tree", relativeURL(res)))
	}

	if r.URL.Path[0:len(prefix)] != prefix {
		panic(fmt.Sprintf("Prefix '%v' does not match request URL path '%v'", prefix, r.URL.Path))
	}

	steps := strings.Split(r.URL.Path[len(prefix):len(r.URL.Path)], "/")
	if steps[len(steps)-1] == "" {
		steps = steps[0 : len(steps)-1]
	}

	ch, rest := navigate(res, steps)
	postAction := (*string)(nil)
	if len(rest) > 0 {
		postAction = &rest[0]
	}
	if ch == nil || (r.Method != "POST" && postAction != nil) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	handle(ch, postAction, prefix, w, r)
}

func navigate(res Resource, steps []string) (Resource, []string) {
	if len(steps) == 0 {
		return res, []string{}
	}

	head := steps[0]
	rest := steps[1:len(steps)]

	if head == "" {
		panic("Empty non-last step during navigation")
	}

	if ch := res.Child(head); ch != nil {
		if ch.PathSegment() != head {
			panic(fmt.Sprintf("Resource '%v' has wrong path segment ('%v' / '%v')", relativeURL(ch), ch.PathSegment(), head))
		}
		return navigate(ch, rest)
	} else {
		if len(rest) != 0 {
			return nil, steps
		}

		// custom POST action
		return res, steps
	}

	// to shut up the compiler
	return nil, nil
}

func handle(res Resource, postAction *string, prefix string, w http.ResponseWriter, r *http.Request) {
	if index(res.AllowedMethods(), r.Method) == -1 {
		w.Header().Set("Allow", strings.Join(res.AllowedMethods(), ", "))
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	switch r.Method {
	case "HEAD":
		setHeaders(res, w)
		w.WriteHeader(http.StatusOK)
	case "GET":
		if etag := res.ETag(); etag != "" {
			if r.Header.Get("If-None-Match") == etag {
				setHeaders(res, w)
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}

		if data, e := res.Get(prefix, r); e != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(e.Error()))
		} else {
			setHeaders(res, w)
			w.Header().Set("Content-Type", res.ContentType())
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		}
	case "POST":
		if postAction != nil {
			if index(res.AllowedActions(), *postAction) == -1 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if e := res.Do(*postAction, r); e != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(e.Error()))
			} else {
				w.WriteHeader(http.StatusNoContent)
			}
		} else {
			coll := res.(Collection)
			if coll == nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if ch, e := coll.Create(r); e != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(e.Error()))
			} else {
				w.Header().Set("Location", AbsoluteURL(r, RelativeURL(prefix, ch)).String())
				w.WriteHeader(http.StatusCreated)
			}
		}
		return
	case "PATCH":
		if e := res.Patch(r); e != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(e.Error()))
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	case "DELETE":
		if res.Parent() == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		coll := res.Parent().(Collection)
		if coll == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if e := coll.Remove(res.PathSegment()); e != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(e.Error()))
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

func index(arr []string, s string) int {
	for i, e := range arr {
		if e == s {
			return i
		}
	}
	return -1
}

func setHeaders(res Resource, w http.ResponseWriter) {
	if etag := res.ETag(); etag != "" {
		w.Header().Set("ETag", etag)
	}
	if t := res.Expires(); !t.IsZero() {
		w.Header().Set("Expires", t.Format(time.RFC1123))
	}
	if cc := res.CacheControl(); cc != "" {
		w.Header().Set("Cache-Control", cc)
	}
}

func AbsoluteURL(r *http.Request, u *url.URL) *url.URL {
	au := *r.URL
	au.Host = r.Host

	return au.ResolveReference(u)
}

func RelativeURL(prefix string, res Resource) *url.URL {
	u := relativeURL(res)
	u.Path = prefix[0:len(prefix)-1] + u.Path
	return u
}

func relativeURL(res Resource) *url.URL {
	parts := make([]string, 0)

	for res != nil {
		parts = append(parts, res.PathSegment())
		res = res.Parent()
	}

	parts = append(parts, "")

	// I love Go
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}

	return &url.URL{
		Path: strings.Join(parts, "/"),
	}
}
