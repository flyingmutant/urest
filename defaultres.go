package urest

import (
	"compress/gzip"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	CONTENT_TYPE_JSON = "application/json; charset=utf-8"
)

type (
	DataResource interface {
		Data(string, *http.Request, time.Time) (interface{}, error)
	}

	RawReadResource interface {
		ReadRaw(string, *http.Request, time.Time) ([]byte, error)
	}

	DefaultResourceImpl struct {
		readRawFunc     func(string, *http.Request, time.Time) ([]byte, error)
		Parent_         Resource
		PathSegment_    string
		Children        map[string]Resource
		AllowedMethods_ []string
		Actions         map[string]func(*http.Request, map[string]interface{}, time.Time) error
		ContentType_    string
		Gzip            bool
		CacheControl_   string
		IsCollection_   bool
	}
)

func NewDefaultResourceImpl(parent Resource, pathSegment string) *DefaultResourceImpl {
	return &DefaultResourceImpl{
		Parent_:         parent,
		PathSegment_:    pathSegment,
		Children:        map[string]Resource{},
		AllowedMethods_: []string{"HEAD"},
		Actions:         map[string]func(*http.Request, map[string]interface{}, time.Time) error{},
		ContentType_:    CONTENT_TYPE_JSON,
	}
}

func (d *DefaultResourceImpl) SetDataDelegate(del DataResource) {
	if ct := d.ContentType(); ct != CONTENT_TYPE_JSON {
		panic("Resource has Data function but non-JSON Content-Type")
	}

	d.readRawFunc = func(prefix string, r *http.Request, t time.Time) ([]byte, error) {
		data, err := del.Data(prefix, r, t)
		if err != nil {
			return nil, err
		}
		return json.Marshal(data)
	}
}

func (d *DefaultResourceImpl) SetRawReadDelegate(del RawReadResource) {
	d.readRawFunc = func(prefix string, r *http.Request, t time.Time) ([]byte, error) {
		return del.ReadRaw(prefix, r, t)
	}
}

func (d *DefaultResourceImpl) AddAction(action string, f func(*http.Request, map[string]interface{}, time.Time) error) {
	d.Actions[action] = f
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
	r := make([]string, 0, len(d.Actions))
	for a, _ := range d.Actions {
		r = append(r, a)
	}
	return r
}

func (d *DefaultResourceImpl) ETag() string {
	return ""
}

func (d *DefaultResourceImpl) Expires() time.Time {
	return time.Time{}
}

func (d *DefaultResourceImpl) CacheControl() string {
	return d.CacheControl_
}

func (d *DefaultResourceImpl) ContentType() string {
	return d.ContentType_
}

func (d *DefaultResourceImpl) Read(urlPrefix string, w http.ResponseWriter, r *http.Request, t time.Time) {
	data, err := d.readRawFunc(urlPrefix, r, t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if d.Gzip && strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
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

func (d *DefaultResourceImpl) Update(r *http.Request, body map[string]interface{}, t time.Time) error {
	panic("Not implemented")
}

func (d *DefaultResourceImpl) Do(action string, r *http.Request, body map[string]interface{}, t time.Time) error {
	if a := d.Actions[action]; a != nil {
		return a(r, body, t)
	}

	panic("Not implemented")
}

func (d *DefaultResourceImpl) IsCollection() bool {
	return d.IsCollection_
}

func (d *DefaultResourceImpl) Create(r *http.Request, body map[string]interface{}, t time.Time) (Resource, error) {
	panic("Not implemented")
}

func (d *DefaultResourceImpl) Delete(s string, body map[string]interface{}, t time.Time) error {
	panic("Not implemented")
}
