package urest

import (
	"errors"
	"net/http"
	"time"
)

type DefaultResourceImpl struct{}

func (DefaultResourceImpl) Parent() Resource {
	return nil
}

func (DefaultResourceImpl) PathSegment() string {
	return ""
}

func (DefaultResourceImpl) Child(string) Resource {
	return nil
}

func (DefaultResourceImpl) AllowedMethods() []string {
	return []string{}
}

func (DefaultResourceImpl) AllowedActions() []string {
	return []string{}
}

func (DefaultResourceImpl) ETag() string {
	return ""
}

func (DefaultResourceImpl) Expires() time.Time {
	return time.Time{}
}

func (DefaultResourceImpl) CacheControl() string {
	return ""
}

func (DefaultResourceImpl) ContentType() string {
	return "application/json; charset=utf-8"
}

func (DefaultResourceImpl) Get(string, *http.Request) ([]byte, error) {
	return nil, errors.New("Method not implemented")
}

func (DefaultResourceImpl) Patch(*http.Request) error {
	return errors.New("Method not implemented")
}

func (DefaultResourceImpl) Do(action string, r *http.Request) error {
	return errors.New("Method not implemented")
}

func (DefaultResourceImpl) Create(*http.Request) (Resource, error) {
	return nil, errors.New("Method not implemented")
}

func (DefaultResourceImpl) Remove(string) error {
	return errors.New("Method not implemented")
}
