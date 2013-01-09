package urest

import (
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path"
	"time"
)

type roFSResource struct {
	path_  string
	parent *roFSResource
}

func NewReadonlyFSResource(path_ string) *roFSResource {
	return &roFSResource{
		path_:  path.Clean(path_),
		parent: nil,
	}
}

func (res *roFSResource) Parent() Resource {
	if res.parent != nil {
		return res.parent
	}
	return nil
}

func (res *roFSResource) PathSegment() string {
	return path.Base(res.path_)
}

func (res *roFSResource) Child(p string) Resource {
	return &roFSResource{path.Join(res.path_, p), res}
}

func (res *roFSResource) AllowedMethods() []string {
	return []string{"HEAD", "GET"}
}

func (res *roFSResource) AllowedActions() []string {
	return []string{}
}

func (res *roFSResource) ETag() string {
	fi, e := os.Stat(res.path_)
	if e != nil {
		return ""
	}
	return fmt.Sprintf("W/\"%v\"", fi.ModTime().UnixNano())
}

func (res *roFSResource) Expires() time.Time {
	return time.Time{}
}

func (res *roFSResource) CacheControl() string {
	return "private, no-transform, max-age=0, must-revalidate"
}

func (res *roFSResource) ContentType() string {
	return mime.TypeByExtension(path.Ext(res.path_))
}

func (res *roFSResource) Gzip() bool {
	switch path.Ext(res.path_) {
	case ".txt":
		return true
	case ".html":
		return true
	case ".htm":
		return true
	case ".css":
		return true
	case ".js":
		return true
	}

	return false
}

func (res *roFSResource) Get(urlPrefix string, r *http.Request) ([]byte, error) {
	f, e := os.Open(res.path_)
	if e != nil {
		return nil, e
	}
	fi, fe := f.Stat()
	if fe != nil {
		return nil, e
	}

	if fi.IsDir() {
		f, e = os.Open(path.Join(res.path_, "index.html"))
		if e != nil {
			return nil, e
		}
	}

	return ioutil.ReadAll(f)
}

func (res *roFSResource) Patch(*http.Request) error {
	panic("Not implemented")
}

func (res *roFSResource) Do(action string, r *http.Request) error {
	panic("Not implemented")
}

func (res *roFSResource) IsCollection() bool {
	return false
}
