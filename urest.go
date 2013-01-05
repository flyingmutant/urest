package urest

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	t_RESET     = "\x1b[0m"
	t_FG_RED    = "\x1b[31m"
	t_FG_GREEN  = "\x1b[32m"
	t_FG_BLUE   = "\x1b[34m"
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

	ETag() string
	Expires() time.Time
	CacheControl() string

	JSON(url.Values) ([]byte, error)
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

	log.Printf("[%v %v] %v %v (%v)", statusC, http.StatusText(lrw.status), methodC, lrw.r.RequestURI, dC)
}

func HandlerWithPrefix(res Resource, prefix string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		lrw := &loggingResponseWriter{w, r, 0, time.Now()}

		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("Error while serving %v %v: %v", r.Method, r.RequestURI, rec)
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				lrw.log()
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
	if len(steps) == 0 || (len(steps) == 1 && steps[0] == "") {
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
		if len(rest) != 1 {
			return nil, rest
		}

		// custom POST action
		return res, rest
	}

	// to shut up the compiler
	return nil, nil
}

func handle(res Resource, postAction *string, prefix string, w http.ResponseWriter, r *http.Request) {
	methodAllowed := false
	for _, m := range res.AllowedMethods() {
		if m == r.Method {
			methodAllowed = true
			break
		}
	}

	if !methodAllowed {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Header().Set("Allow", strings.Join(res.AllowedMethods(), ", "))
		return
	}

	switch r.Method {
	case "HEAD":
		w.WriteHeader(http.StatusOK)
		writeHeaders(res, w)
	case "GET":
		if etag := res.ETag(); etag != "" {
			if r.Header.Get("If-None-Match") == etag {
				w.WriteHeader(http.StatusNotModified)
				writeHeaders(res, w)
				return
			}
		}

		if data, e := res.JSON(r.URL.Query()); e != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(e.Error()))
		} else {
			w.WriteHeader(http.StatusOK)
			writeHeaders(res, w)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Write(data)
		}
	case "POST":
		if postAction != nil {
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
				w.WriteHeader(http.StatusCreated)
				relURL := relativeURL(ch)
				relURL.Path = prefix + relURL.Path
				w.Header().Set("Location", r.URL.ResolveReference(relURL).String())
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

func writeHeaders(res Resource, w http.ResponseWriter) {
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
