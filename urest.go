package urest

import (
	"log"
	"net/http"
)

type Id int
type AttrMap map[string]interface{}

type ResourceSpec struct {
	Name     string
	Parent   *ResourceSpec
	Children map[string]*ResourceSpec

	Create func(parent Resource) Resource
	Find   func(Id, parent Resource) Resource
	Remove func(Id, parent Resource)
}

func defaultCreate(parent Resource) Resource {
	log.Panic("Resource '%v' does not support creation of subresources", parent.Spec().Name)
	return nil
}

func defaultFind(Id, parent Resource) Resource {
	log.Panic("Resource '%v' does not support searching for subresources (id=%v)", parent.Spec().Name, Id)
	return nil
}
func defaultRemove(Id, parent Resource) {
	log.Panic("Resource '%v' does not support removal of subresources", parent.Spec().Name)
}

type Resource interface {
	Spec() *ResourceSpec
	Json(attrs AttrMap) []byte
	Patch(attrs AttrMap)
	Do(action string, attrs AttrMap)
}

func NewResourceSpec(name string, parent *ResourceSpec) *ResourceSpec {
	urlPrefix := name

	if parent == nil && urlPrefix != "" {
		log.Panic("Can not add resource spec with no parent and non-empty URL prefix")
	}

	r := &ResourceSpec{
		Name:     name,
		Parent:   parent,
		Children: make(map[string]*ResourceSpec),
		Create:   defaultCreate,
		Find:     defaultFind,
		Remove:   defaultRemove,
	}

	if parent != nil {
		parent.Children[urlPrefix] = r
	}

	return r
}

func (rs *ResourceSpec) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if rs.Parent == nil {
		log.Panic("Resource spec '%v' is not a root of the resource tree")
	}

	// TODO

}

// TODO default implementations of various methods
