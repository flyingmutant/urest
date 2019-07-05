package urest

import (
	"bytes"
	"bufio"
	"compress/gzip"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"log"
 	"github.com/francoispqt/gojay"
)

type (
	DataResource interface {
		Data(string, *http.Request) (interface{}, error)
		LiveData(string, *http.Request) (interface{}, error)
	}

	RawReadResource interface {
		ReadRaw(string, *http.Request) ([]byte, error)
	}

	DefaultResourceImpl struct {
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
		isLive := GetRequestData(r,"livedata")
		data := interface{}(nil)
		err := error(nil)
		if b, ok := isLive.(bool); ok && b {
			data, err = del.LiveData(prefix, r)
		} else {
			data, err = del.Data(prefix, r)
		}
		if err != nil {
			return nil, err
		}
		if data == nil {
			return []byte{}, nil
		}
		var b bytes.Buffer
		iowriter := bufio.NewWriter(&b)
		enc := gojay.NewEncoder(iowriter)
		if err := enc.Encode(data); err != nil {
			log.Printf("Error in json encoding: %v", err)
			return []byte{},err

		}
		return b.Bytes(), nil
	}
}

func (d *DefaultResourceImpl) SetRawReadDelegate(del RawReadResource) {
	d.readRawFunc = func(prefix string, r *http.Request) ([]byte, error) {
		return del.ReadRaw(prefix, r)
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

func (d *DefaultResourceImpl) Child(name string, r *http.Request) Resource {
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

func (d *DefaultResourceImpl) ETag(r *http.Request) string {
	return ""
}

func (d *DefaultResourceImpl) Expires() time.Time {
	if d.CacheDuration == 0 {
		return time.Time{}
	}
	return time.Now().Add(d.CacheDuration)
}

func (d *DefaultResourceImpl) CacheControl() string {
	if d.CacheDuration == 0 {
		return ""
	}
	return fmt.Sprintf("max-age=%d", d.CacheDuration/time.Second)
}

func (d *DefaultResourceImpl) ContentType() string {
	return d.ContentType_
}

func (d *DefaultResourceImpl) Read(urlPrefix string, w http.ResponseWriter, r *http.Request) error {
	if d.readRawFunc == nil {
		panic("Not implemented")
	}

	data, err := d.readRawFunc(urlPrefix, r)
	if err != nil {
		return err
	}

	w.Header().Set("Vary", "Accept-Encoding")
	if d.Gzip && strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		b := bytes.Buffer{}
		gz := gzip.NewWriter(&b)
		gz.Write(data)
		gz.Close()

		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Content-Length", strconv.Itoa(b.Len()))
		w.Write(b.Bytes())
	} else {
		w.Header().Set("Content-Length", strconv.Itoa(len(data)))
		w.Write(data)
	}

	return nil
}

func (*DefaultResourceImpl) Update(*http.Request) error {
	panic("Not implemented")
}

func (*DefaultResourceImpl) Replace(*http.Request) error {
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

func (*DefaultResourceImpl) Create(*http.Request) (Resource, error) {
	panic("Not implemented")
}

func (*DefaultResourceImpl) Delete(string, *http.Request) error {
	panic("Not implemented")
}

func (*DefaultResourceImpl) LiveData(string, *http.Request) (interface{}, error) {
	return nil, nil
}
