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

type DefaultResourceImpl struct {
	Parent_         Resource
	PathSegment_    string
	Children        map[string]Resource
	AllowedMethods_ []string
	Actions         map[string]func(*http.Request, map[string]interface{}, time.Time) error
	ContentType_    string
	Gzip            bool
	CacheControl_   string
	IsCollection_   bool

	etagFunc    func() string
	expiresFunc func() time.Time
	readFunc    func(string, *http.Request, time.Time) ([]byte, error)
	updateFunc  func(*http.Request, map[string]interface{}, time.Time) error
	createFunc  func(*http.Request, map[string]interface{}, time.Time) (Resource, error)
	deleteFunc  func(string, map[string]interface{}, time.Time) error
}

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

func (d *DefaultResourceImpl) SetDelegate(del interface{}) {
	d.setETagDelegate(del)
	d.setExpiresDelegate(del)
	d.setReadDelegate(del)
	d.setUpdateDelegate(del)
	d.setCreateDelegate(del)
	d.setDeleteDelegate(del)
}

func (d *DefaultResourceImpl) setETagDelegate(del interface{}) {
	if i, ok := del.(interface {
		ETag() string
	}); ok {
		d.etagFunc = func() string {
			return i.ETag()
		}
	} else {
		d.etagFunc = func() string {
			return ""
		}
	}
}

func (d *DefaultResourceImpl) setExpiresDelegate(del interface{}) {
	if i, ok := del.(interface {
		Expires() time.Time
	}); ok {
		d.expiresFunc = func() time.Time {
			return i.Expires()
		}
	} else {
		d.expiresFunc = func() time.Time {
			return time.Time{}
		}
	}
}

func (d *DefaultResourceImpl) setReadDelegate(del interface{}) {
	if i, ok := del.(interface {
		Data(string, *http.Request, time.Time) (interface{}, error)
	}); ok {
		if ct := d.ContentType(); ct != CONTENT_TYPE_JSON {
			panic("Resource has Data function but non-JSON Content-Type")
		}

		d.readFunc = func(prefix string, r *http.Request, t time.Time) ([]byte, error) {
			data, err := i.Data(prefix, r, t)
			if err != nil {
				return nil, err
			}
			return json.Marshal(data)
		}
	} else {
		ri, rok := del.(interface {
			Read(string, *http.Request, time.Time) ([]byte, error)
		})
		if rok {
			d.readFunc = func(prefix string, r *http.Request, t time.Time) ([]byte, error) {
				return ri.Read(prefix, r, t)
			}
		}
	}
}

func (d *DefaultResourceImpl) setUpdateDelegate(del interface{}) {
	i, ok := del.(interface {
		Update(*http.Request, map[string]interface{}, time.Time) error
	})
	if ok {
		d.updateFunc = func(r *http.Request, body map[string]interface{}, t time.Time) error {
			return i.Update(r, body, t)
		}
	}
}

func (d *DefaultResourceImpl) setCreateDelegate(del interface{}) {
	i, ok := del.(interface {
		Create(*http.Request, map[string]interface{}, time.Time) (Resource, error)
	})
	if ok {
		d.createFunc = func(r *http.Request, body map[string]interface{}, t time.Time) (Resource, error) {
			return i.Create(r, body, t)
		}
	}
}

func (d *DefaultResourceImpl) setDeleteDelegate(del interface{}) {
	i, ok := del.(interface {
		Delete(string, map[string]interface{}, time.Time) error
	})
	if ok {
		d.deleteFunc = func(s string, body map[string]interface{}, t time.Time) error {
			return i.Delete(s, body, t)
		}
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
	return d.etagFunc()
}

func (d *DefaultResourceImpl) Expires() time.Time {
	return d.expiresFunc()
}

func (d *DefaultResourceImpl) CacheControl() string {
	return d.CacheControl_
}

func (d *DefaultResourceImpl) ContentType() string {
	return d.ContentType_
}

func (d *DefaultResourceImpl) Read(urlPrefix string, w http.ResponseWriter, r *http.Request, t time.Time) {
	data, err := d.readFunc(urlPrefix, r, t)
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
	return d.updateFunc(r, body, t)
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
	return d.createFunc(r, body, t)
}

func (d *DefaultResourceImpl) Delete(s string, body map[string]interface{}, t time.Time) error {
	return d.deleteFunc(s, body, t)
}
