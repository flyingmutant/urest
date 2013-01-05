package urest

import (
	"errors"
	"net/http"
	"time"
)

type ActionFunc func(r *http.Request) error

type DefaultResourceImpl struct {
	Parent_         Resource
	PathSegment_    string
	Children        map[string]Resource
	AllowedMethods_ []string
	Actions         map[string]ActionFunc
	ContentType_    string
	CacheControl_   string
}

func NewRootDefaultResourceImpl() *DefaultResourceImpl {
	return &DefaultResourceImpl{
		Children:        make(map[string]Resource),
		AllowedMethods_: []string{},
		Actions:         make(map[string]ActionFunc),
		ContentType_:    "application/json; charset=utf-8",
	}
}

func NewChildDefaultResourceImpl(parent *DefaultResourceImpl, pathSegment string) *DefaultResourceImpl {
	d := NewRootDefaultResourceImpl()
	d.Parent_ = parent
	d.PathSegment_ = pathSegment

	parent.Children[pathSegment] = d

	return d
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

func (d *DefaultResourceImpl) Get(string, *http.Request) ([]byte, error) {
	return nil, errors.New("Not implemented")
}

func (d *DefaultResourceImpl) Patch(*http.Request) error {
	return errors.New("Not implemented")
}

func (d *DefaultResourceImpl) Do(action string, r *http.Request) error {
	if a := d.Actions[action]; a != nil {
		return a(r)
	}

	return errors.New("Action not supported")
}
