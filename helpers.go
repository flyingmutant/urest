package urest

import (
	"errors"
	"io/ioutil"
	"net/http"
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
	Actions         map[string]func(*http.Request) error
	ETagFunc        func() string
	ExpiresFunc     func() time.Time
	ContentType_    string
	CacheControl_   string
	GetFunc         func(string, *http.Request) ([]byte, error)
	PatchFunc       func(*http.Request) error
	IsCollection_   bool
	CreateFunc      func(*http.Request) (Resource, error)
	RemoveFunc      func(string) error
}

func NewDefaultResourceImpl(parent Resource, pathSegment string, isCollection bool) *DefaultResourceImpl {
	return &DefaultResourceImpl{
		Parent_:         parent,
		PathSegment_:    pathSegment,
		Children:        make(map[string]Resource),
		AllowedMethods_: []string{},
		Actions:         make(map[string]func(*http.Request) error),
		ETagFunc: func() string {
			return ""
		},
		ExpiresFunc: func() time.Time {
			return time.Time{}
		},
		GetFunc: func(string, *http.Request) ([]byte, error) {
			return nil, errors.New("Not implemented")
		},
		PatchFunc: func(*http.Request) error {
			return errors.New("Not implemented")
		},
		IsCollection_: isCollection,
		CreateFunc: func(*http.Request) (Resource, error) {
			return nil, errors.New("Not implemented")
		},
		RemoveFunc: func(string) error {
			return errors.New("Not implemented")
		},
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

func (d *DefaultResourceImpl) Get(urlPrefix string, r *http.Request) ([]byte, error) {
	return d.GetFunc(urlPrefix, r)
}

func (d *DefaultResourceImpl) Patch(r *http.Request) error {
	return d.PatchFunc(r)
}

func (d *DefaultResourceImpl) Do(action string, r *http.Request) error {
	if a := d.Actions[action]; a != nil {
		return a(r)
	}

	return errors.New("Action not supported")
}

func (d *DefaultResourceImpl) IsCollection() bool {
	return d.IsCollection_
}

func (d *DefaultResourceImpl) Create(r *http.Request) (Resource, error) {
	return d.CreateFunc(r)
}

func (d *DefaultResourceImpl) Remove(s string) error {
	return d.RemoveFunc(s)
}
