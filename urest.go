package urest

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
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

		Read(urlPrefix string, w http.ResponseWriter, r *http.Request) error
		Update(*http.Request) error
		Replace(*http.Request) error
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
)

func NewHandler(res Resource, prefix string) *Handler {
	if !strings.HasPrefix(prefix, "/") {
		log.Panicf("Invalid prefix '%v'", prefix)
	}

	return &Handler{res, prefix}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.res.Parent() != nil {
		log.Panicf("Resource '%v' is not a root of the resource tree", relativeURL(h.res))
	}

	if r.URL.Path[:len(h.prefix)] != h.prefix {
		log.Panicf("Prefix '%v' does not match request URL path '%v'", h.prefix, r.URL.Path)
	}

	steps := strings.Split(r.URL.Path[len(h.prefix):], "/")
	ch, rest, err := navigate(h.res, steps, r)

	if err != nil {
		log.Printf("Navigation failed: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	postAction := (*string)(nil)
	if len(rest) > 0 {
		postAction = &rest[0]
	}

	if postAction != nil && r.Method != "POST" {
		http.NotFound(w, r)
		return
	}
	if postAction == nil && RelativeURL(h.prefix, ch).Path != r.URL.Path {
		w.Header().Set("Location", RelativeURL(h.prefix, ch).Path)
		w.WriteHeader(http.StatusMovedPermanently)
		return
	}

	handle(ch, postAction, h.prefix, w, r)
}

func navigate(res Resource, steps []string, r *http.Request) (Resource, []string, error) {
	if len(steps) == 0 {
		return res, []string{}, nil
	}

	head := steps[0]
	rest := steps[1:]

	if head == "" {
		return navigate(res, rest, r)
	}

	if ch := res.Child(head, r); ch != nil {
		if ch.PathSegment() != head {
			return nil, nil, fmt.Errorf("Resource '%v' has wrong path segment ('%v' / '%v')", relativeURL(ch), ch.PathSegment(), head)
		}
		return navigate(ch, rest, r)
	} else {
		if len(rest) != 0 {
			return nil, nil, fmt.Errorf("Non-empty descendants of non-child node ('%v' / '%v')", head, rest)
		}
		return res, []string{head}, nil
	}
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

		if e := res.Read(prefix, w, r); e != nil {
			ReportError(w, e)
		}
	case "POST":
		if postAction != nil {
			if index(res.AllowedActions(), *postAction) == -1 {
				http.Error(w, "Unknown action", http.StatusBadRequest)
				return
			}

			if e := res.Do(*postAction, r); e != nil {
				ReportError(w, e)
			} else {
				w.WriteHeader(http.StatusNoContent)
			}
		} else {
			if res.IsCollection() {
				if ch, e := res.(Collection).Create(r); e != nil {
					ReportError(w, e)
				} else {
					w.Header().Set("Location", RelativeURL(prefix, ch).String())
					w.WriteHeader(http.StatusCreated)
				}
			} else {
				if e := res.Replace(r); e != nil {
					ReportError(w, e)
				} else {
					w.WriteHeader(http.StatusNoContent)
				}
			}
		}
	case "PATCH":
		if e := res.Update(r); e != nil {
			ReportError(w, e)
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
			ReportError(w, e)
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
		w.Header().Set("Expires", t.Format(http.TimeFormat))
	}
	if cc := res.CacheControl(); cc != "" {
		w.Header().Set("Cache-Control", cc)
	} else {
		w.Header().Set("Cache-Control", "private, must-revalidate, max-age=0")
	}
	if et := etag(res, r); et != "" {
		w.Header().Set("ETag", et)
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
	if strings.HasSuffix(prefix, "/") && strings.HasPrefix(u.Path, "/") {
		prefix = prefix[:len(prefix)-1]
	}
	u.Path = prefix + u.Path
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
