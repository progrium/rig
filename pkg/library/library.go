// library provides an API for defining component catalogs that can be
// registered with modules.
package library

import (
	"fmt"
	"path"
	"reflect"
	"runtime"
)

type Item any

var DefaultCatalog Catalog

type Catalog []Item

func (b *Catalog) Register(items ...Item) error {
	*b = append(*b, items...)
	return nil
}

type Component struct {
	Value  any
	Prefix string
	Source string
	Icon   string
	Desc   string
}

func (c Component) New() any {
	nv := reflect.New(reflect.TypeOf(c.Value))
	nv.Elem().Set(reflect.ValueOf(c.Value))
	return nv.Interface()
}

// when adding components to raw nodes, there is no
// catalog just consistent component IDs, so we make
// this available
func ComponentID(v any) string {
	t := reflect.TypeOf(v)
	return fmt.Sprintf("%s.%s", t.PkgPath(), t.Name())
}

func (c Component) ID() string {
	return ComponentID(c.Value)
}

func (c Component) Name() string {
	if c.Prefix == "" {
		return path.Base(c.ID())
	}
	return fmt.Sprintf("%s.%s", c.Prefix, reflect.TypeOf(c.Value).Name())
}

// Filepath returns the file path of the calling
// source file with an optional subpath
func Filepath(subpaths ...string) string {
	_, filename, _, _ := runtime.Caller(1)
	return path.Join(path.Dir(filename), path.Join(subpaths...))
}
