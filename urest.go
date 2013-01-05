package urest

// TODO
// - predefined Cache-Control strings

// TODO struct with default impl of resource methods

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	// "strconv"
	"strings"
	"time"
)

type Matcher func(Resource, string) Resource

type Resource interface {
	Parent() Resource
	PathSegment() string
	Matchers() []Matcher

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

func HandlerWithPrefix(res Resource, prefix string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("Error while serving %v %v: %v", r.Method, r.RequestURI, rec)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		handleWithPrefix(res, prefix, w, r)
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
	if ch == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	handle(ch, rest, prefix, w, r)
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

	if ch := child(res, head); ch != nil {
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

func child(res Resource, pathSegment string) Resource {
	for _, m := range res.Matchers() {
		if ret := m(res, pathSegment); ret != nil {
			if ret.PathSegment() != pathSegment {
				panic(fmt.Sprintf("Resource '%v' has wrong path segment ('%v' / '%v')", relativeURL(ret), ret.PathSegment(), pathSegment))
			}
			return ret
		}
	}

	return nil
}

func handle(res Resource, rest []string, prefix string, w http.ResponseWriter, r *http.Request) {
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
			writeHeaders(res, w)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Write(data)
		}
	case "POST":
		if len(rest) == 1 {
			e := res.Do(rest[0], r)
			if e != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(e.Error()))
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
	case "PATCH":
		// TODO
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
