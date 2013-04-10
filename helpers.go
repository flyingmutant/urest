package urest

import (
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	CONTENT_TYPE_JSON = "application/json; charset=utf-8"
)

func ReadBody(r *http.Request) ([]byte, error) {
	defer r.Body.Close()

	return ioutil.ReadAll(r.Body)
}

type DefaultResourceImpl struct {
	Parent_         Resource
	PathSegment_    string
	Children        map[string]Resource
	AllowedMethods_ []string
	Actions         map[string]func(*http.Request, time.Time) error
	ETagFunc        func() string
	ExpiresFunc     func() time.Time
	ContentType_    string
	GzipFunc        func() bool
	CacheControl_   string
	ReadFunc        func(string, *http.Request, time.Time) ([]byte, error)
	UpdateFunc      func(*http.Request, time.Time) error
	IsCollection_   bool
	CreateFunc      func(*http.Request, time.Time) (Resource, error)
	DeleteFunc      func(string, time.Time) error
}

func NewDefaultResourceImpl(parent Resource, pathSegment string, isCollection bool, contentType string) *DefaultResourceImpl {
	return &DefaultResourceImpl{
		Parent_:         parent,
		PathSegment_:    pathSegment,
		Children:        make(map[string]Resource),
		AllowedMethods_: []string{"HEAD"},
		Actions:         make(map[string]func(*http.Request, time.Time) error),
		ETagFunc:        func() string { return "" },
		ExpiresFunc:     func() time.Time { return time.Time{} },
		ContentType_:    contentType,
		GzipFunc:        func() bool { return false },
		ReadFunc:        func(string, *http.Request, time.Time) ([]byte, error) { panic("Not implemented") },
		UpdateFunc:      func(*http.Request, time.Time) error { panic("Not implemented") },
		IsCollection_:   isCollection,
		CreateFunc:      func(*http.Request, time.Time) (Resource, error) { panic("Not implemented") },
		DeleteFunc:      func(string, time.Time) error { panic("Not implemented") },
	}
}

func (d *DefaultResourceImpl) Parent() Resource {
	return d.Parent_
}

func (d *DefaultResourceImpl) PathSegment() string {
	return d.PathSegment_
}

func (d *DefaultResourceImpl) Child(name string) Resource {
	return d.Children[name]
}

func (d *DefaultResourceImpl) AllowedMethods() []string {
	return d.AllowedMethods_
}

func (d *DefaultResourceImpl) AllowedActions() []string {
	ret := make([]string, 0, len(d.Actions))

	for a, _ := range d.Actions {
		ret = append(ret, a)
	}

	return ret
}

func (d *DefaultResourceImpl) ETag() string {
	return d.ETagFunc()
}

func (d *DefaultResourceImpl) Expires() time.Time {
	return d.ExpiresFunc()
}

func (d *DefaultResourceImpl) CacheControl() string {
	return d.CacheControl_
}

func (d *DefaultResourceImpl) ContentType() string {
	return d.ContentType_
}

func (d *DefaultResourceImpl) Read(urlPrefix string, w http.ResponseWriter, r *http.Request, t time.Time) {
	if data, e := d.ReadFunc(urlPrefix, r, t); e != nil {
		http.Error(w, e.Error(), http.StatusBadRequest)
	} else {
		if d.GzipFunc() && strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			gz := gzip.NewWriter(w)
			defer gz.Close()
			w.Header().Set("Content-Encoding", "gzip")
			w.WriteHeader(http.StatusOK)
			gz.Write(data)
		} else {
			w.Header().Set("Content-Length", strconv.Itoa(len(data)))
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		}
	}
}

func (d *DefaultResourceImpl) Update(r *http.Request, t time.Time) error {
	return d.UpdateFunc(r, t)
}

func (d *DefaultResourceImpl) Do(action string, r *http.Request, t time.Time) error {
	if a := d.Actions[action]; a != nil {
		return a(r, t)
	}

	panic("Not implemented")
}

func (d *DefaultResourceImpl) IsCollection() bool {
	return d.IsCollection_
}

func (d *DefaultResourceImpl) Create(r *http.Request, t time.Time) (Resource, error) {
	return d.CreateFunc(r, t)
}

func (d *DefaultResourceImpl) Delete(s string, t time.Time) error {
	return d.DeleteFunc(s, t)
}
