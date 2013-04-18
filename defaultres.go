package urest

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type (
	DataResource interface {
		Data(string, *http.Request) (interface{}, error)
	}

	RawReadResource interface {
		ReadRaw(string, *http.Request) ([]byte, error)
	}

	Digester interface {
		Digest() []byte
	}

	DefaultResourceImpl struct {
		etagFunc        func() string
		readRawFunc     func(string, *http.Request) ([]byte, error)
		Parent_         Resource
		PathSegment_    string
		IsCollection_   bool
		Children        map[string]Resource
		AllowedMethods_ []string
		Actions         map[string]func(*http.Request) error
		ContentType_    string
		Gzip            bool
		CacheDuration   time.Duration
	}
)

func CacheControl(maxAge time.Duration) string {
	return fmt.Sprintf("max-age=%d, must-revalidate", maxAge/time.Second)
}

func NewDefaultResourceImpl(parent Resource, pathSegment string) *DefaultResourceImpl {
	return &DefaultResourceImpl{
		Parent_:         parent,
		PathSegment_:    pathSegment,
		Children:        map[string]Resource{},
		AllowedMethods_: []string{"HEAD"},
		Actions:         map[string]func(*http.Request) error{},
		ContentType_:    CONTENT_TYPE_JSON,
		Gzip:            true,
	}
}

func (d *DefaultResourceImpl) SetDataDelegate(del DataResource) {
	if ct := d.ContentType(); ct != CONTENT_TYPE_JSON {
		panic("Resource has Data function but non-JSON Content-Type")
	}

	d.readRawFunc = func(prefix string, r *http.Request) ([]byte, error) {
		data, err := del.Data(prefix, r)
		if err != nil {
			return nil, err
		}
		return json.Marshal(data)
	}
}

func (d *DefaultResourceImpl) SetRawReadDelegate(del RawReadResource) {
	d.readRawFunc = func(prefix string, r *http.Request) ([]byte, error) {
		return del.ReadRaw(prefix, r)
	}
}

func (d *DefaultResourceImpl) SetDigester(dg Digester) {
	d.etagFunc = func() string {
		return fmt.Sprintf("%x", dg.Digest())
	}
}

func (d *DefaultResourceImpl) AddAction(action string, f func(*http.Request) error) {
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
	if d.etagFunc == nil {
		return ""
	}
	return d.etagFunc()
}

func (d *DefaultResourceImpl) Expires() time.Time {
	return time.Time{}
}

func (d *DefaultResourceImpl) CacheControl() string {
	if d.CacheDuration == 0 {
		return ""
	}
	return CacheControl(d.CacheDuration)
}

func (d *DefaultResourceImpl) ContentType() string {
	return d.ContentType_
}

func (d *DefaultResourceImpl) Read(urlPrefix string, w http.ResponseWriter, r *http.Request) {
	data, err := d.readRawFunc(urlPrefix, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if d.Gzip && strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		b := bytes.Buffer{}
		gz := gzip.NewWriter(&b)
		gz.Write(data)
		gz.Close()

		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Content-Length", strconv.Itoa(b.Len()))
		w.WriteHeader(http.StatusOK)
		w.Write(b.Bytes())
	} else {
		w.Header().Set("Content-Length", strconv.Itoa(len(data)))
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

func (d *DefaultResourceImpl) Update(r *http.Request) error {
	panic("Not implemented")
}

func (d *DefaultResourceImpl) Do(action string, r *http.Request) error {
	if a := d.Actions[action]; a != nil {
		return a(r)
	}

	panic("Not implemented")
}

func (d *DefaultResourceImpl) IsCollection() bool {
	return d.IsCollection_
}

func (d *DefaultResourceImpl) Create(r *http.Request) (Resource, error) {
	panic("Not implemented")
}

func (d *DefaultResourceImpl) Delete(s string) error {
	panic("Not implemented")
}
