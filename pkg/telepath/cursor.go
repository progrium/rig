package telepath

import (
	"path/filepath"
	"reflect"
	"strings"
)

type Cursor interface {
	Selector
	Path() string

	ValueTo(v any) error
	Value() (any, error)

	List() ([]string, error)
	Meta() (Metadata, error)

	Set(v any) error
	Delete() error
	Insert(idx int, v any) error
	Call(args []any, ret ...any) error
}

type C struct {
	Root
	Telepath
}

func NewCursor(root Root, path ...string) Cursor {
	return C{Root: root, Telepath: Telepath{strings.Join(path, "/")}}
}

func (c C) Select(path ...string) Cursor {
	return NewCursor(c.Root, filepath.Join(c.Path(), strings.Join(path, "/")))
}

func (c C) ValueTo(v any) error {
	vv, err := c.Root.Get(c.Path())
	if err != nil {
		return err
	}
	src := reflect.ValueOf(vv)
	dst := reflect.ValueOf(v)
	dst.Elem().Set(src)
	return nil
}

func (c C) Value() (any, error) {
	return c.Root.Get(c.Path())
}

func (c C) Set(v any) error {
	return c.Root.Set(c.Path(), v)
}

func (c C) Delete() error {
	return c.Root.Delete(c.Path())
}

func (c C) Insert(idx int, v any) error {
	return c.Root.Insert(c.Path(), idx, v)
}

func (c C) List() ([]string, error) {
	return c.Root.List(c.Path())
}

func (c C) Meta() (Metadata, error) {
	return c.Root.Meta(c.Path())
}

func min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

func (c C) Call(args []any, ret ...any) error {
	out, err := c.Root.Call(c.Path(), args)
	if err != nil {
		return err
	}
	if ret != nil && out != nil {
		if len(ret) == 1 && len(out) > 1 && reflect.TypeOf(ret[0]).Elem().Kind() == reflect.Slice {
			// if only 1 ret pointer and is a slice, and multiple out, set it to out
			reflect.ValueOf(ret).Elem().Set(reflect.ValueOf(out))
		} else {
			// otherwise set ret pointers to equivalent out value
			for i, r := range ret[:min(len(out), len(ret))] {
				reflect.ValueOf(r).Elem().Set(reflect.ValueOf(out[i]))
			}
		}
	}
	return nil
}

func (c C) Path() string {
	return c.Telepath.Telepath
}

func (c C) PathNames() (names []string) {
	names, _ = c.List()
	return
}

func (c C) PathValue(name string) reflect.Value {
	return reflect.ValueOf(c.Select(name))
}
