package urest

import (
	"errors"
	"net/http"
	"time"
)

type DefaultResourceImpl struct {
	parent         Resource
	pathSegment    string
	children       map[string]Resource
	allowedMethods []string
	allowedActions []string
}

func NewRootDefaultResourceImpl() *DefaultResourceImpl {
	return &DefaultResourceImpl{
		children:       make(map[string]Resource),
		allowedMethods: []string{},
		allowedActions: []string{},
	}
}

func NewChildDefaultResourceImpl(parent *DefaultResourceImpl, name string) *DefaultResourceImpl {
	d := NewRootDefaultResourceImpl()
	d.parent = parent

	parent.children[name] = d

	return d
}

func (d *DefaultResourceImpl) Parent() Resource {
	return d.parent
}

func (d *DefaultResourceImpl) PathSegment() string {
	return d.pathSegment
}

func (d *DefaultResourceImpl) Child(name string) Resource {
	return d.children[name]
}

func (d *DefaultResourceImpl) AllowedMethods() []string {
	return d.allowedMethods
}

func (d *DefaultResourceImpl) AllowedActions() []string {
	return d.allowedActions
}

func (d *DefaultResourceImpl) ETag() string {
	return ""
}

func (d *DefaultResourceImpl) Expires() time.Time {
	return time.Time{}
}

func (d *DefaultResourceImpl) CacheControl() string {
	return ""
}

func (d *DefaultResourceImpl) ContentType() string {
	return "application/json; charset=utf-8"
}

func (d *DefaultResourceImpl) Get(string, *http.Request) ([]byte, error) {
	return nil, errors.New("Method not implemented")
}

func (d *DefaultResourceImpl) Patch(*http.Request) error {
	return errors.New("Method not implemented")
}

func (d *DefaultResourceImpl) Do(action string, r *http.Request) error {
	return errors.New("Method not implemented")
}

func (d *DefaultResourceImpl) Create(*http.Request) (Resource, error) {
	return nil, errors.New("Method not implemented")
}

func (d *DefaultResourceImpl) Remove(string) error {
	return errors.New("Method not implemented")
}
