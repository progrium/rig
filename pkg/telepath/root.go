package telepath

import (
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/progrium/jsonschema"
	"tractor.dev/toolkit-go/duplex/fn"
)

var ErrInvalidPath = "invalid path for %s: %s"

type Root interface {
	Selector

	Get(path string) (any, error)
	Set(path string, v any) error
	List(path string) ([]string, error)
	Meta(path string) (Metadata, error)
	Delete(path string) error
	Insert(path string, idx int, v any) error
	Call(path string, args []any) ([]any, error)
}

type R struct {
	Value any
}

func New(v any) R {
	return R{Value: v}
}

func (r R) Select(path ...string) Cursor {
	return C{Root: r, Telepath: Telepath{strings.TrimLeft(strings.Join(path, "/"), "/")}}
}

func (r R) get(path string) reflect.Value {
	if path == "" {
		return reflect.ValueOf(r.Value)
	}

	v := WalkTo(r.Value, path)

	// TODO: check if this is necessary, Value already does this?
	// if its a callable with no arguments (and single return value), return its output
	if v.Kind() == reflect.Func && v.Type().NumIn() == 0 && v.Type().NumOut() == 1 {
		v = v.Call([]reflect.Value{})[0]
	}

	return v
}

func (r R) Meta(path string) (Metadata, error) {
	v := r.get(path)
	if !v.IsValid() {
		return Metadata{}, fmt.Errorf(ErrInvalidPath, "meta", path)
	}
	if c, ok := v.Interface().(Cursor); ok {
		return c.Meta()
	}
	var fields []string
	var methods []string
	if v.Type().Kind() == reflect.Struct {
		for i := 0; i < v.Type().NumField(); i++ {
			fields = append(fields, v.Type().Field(i).Name)
		}
		for i := 0; i < v.Type().NumMethod(); i++ {
			methods = append(methods, v.Type().Method(i).Name)
		}
	}
	if v.Type().Kind() == reflect.Pointer && v.Type().Elem().Kind() == reflect.Struct {
		for i := 0; i < v.Type().Elem().NumField(); i++ {
			fields = append(fields, v.Type().Elem().Field(i).Name)
		}
		for i := 0; i < v.Type().NumMethod(); i++ {
			methods = append(methods, v.Type().Method(i).Name)
		}
		for i := 0; i < v.Type().Elem().NumMethod(); i++ {
			methods = append(methods, v.Type().Elem().Method(i).Name)
		}
	}
	if v.Type().Kind() == reflect.Interface {
		t := v.Elem().Type()

		for i := 0; i < t.Elem().NumField(); i++ {
			fields = append(fields, t.Elem().Field(i).Name)
		}
		for i := 0; i < t.NumMethod(); i++ {
			methods = append(methods, t.Method(i).Name)
		}
		for i := 0; i < t.Elem().NumMethod(); i++ {
			methods = append(methods, t.Elem().Method(i).Name)
		}
	}

	jsr := &jsonschema.Reflector{
		AnnotatePointers:          true,
		AnnotatePackages:          true,
		AnnotateMethods:           true,
		AnnotateNames:             true,
		AllowAdditionalProperties: true,
	}
	m := Metadata{
		Type: Type{
			Name:    v.Type().Name(),
			Kind:    v.Type().Kind().String(),
			Fields:  fields,
			Methods: methods,
			// TODO: Elem
		},
		Schema: jsr.ReflectFromType(v.Type()),
	}
	switch v.Type().Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		m.Len = v.Len()
	}
	return m, nil
}

func (r R) Set(path string, v any) error {
	// todo: catch panics

	parts := strings.Split(strings.TrimLeft(path, "/"), "/")
	if len(parts) == 0 {
		return fmt.Errorf(ErrInvalidPath, "set", path) // unable to set root
	}
	var dst reflect.Value
	var base string
	if len(parts) == 1 {
		dst = reflect.ValueOf(r.Value)
		base = parts[0]
	} else {
		dir := filepath.Dir(path)
		dst = WalkTo(r.Value, dir)
		base = filepath.Base(path)
	}

	var rv reflect.Value
	if c, ok := FromTelepath(r, v); ok {
		v = c
		rv = reflect.ValueOf(r.get(c.Path()).Interface()) // re-ValueOf to get actual type
	} else {
		rv = reflect.ValueOf(v)
	}

	if dst.Kind() == reflect.Map {
		key := reflect.ValueOf(base).Convert(dst.Type().Key())
		if v == nil {
			dst.SetMapIndex(key, rv)
		} else {
			dst.SetMapIndex(key, ensureType(rv, dst.Type().Elem()))
		}
		return nil
	}

	dst = Value(dst, base)
	if c, ok := dst.Interface().(Cursor); ok {
		return c.Set(v)
	}
	if v == nil {
		dst.Set(reflect.Zero(dst.Type()))
	} else {
		dst.Set(ensureType(rv, dst.Type()))
	}
	return nil
}

func (r R) List(path string) ([]string, error) {
	v := r.get(path)
	if !v.IsValid() {
		return nil, fmt.Errorf(ErrInvalidPath, "list", path)
	}
	if c, ok := v.Interface().(Cursor); ok {
		return c.List()
	}
	return Names(v.Interface()), nil
}

func (r R) Delete(path string) error {
	parts := strings.Split(strings.TrimLeft(path, "/"), "/")
	if len(parts) == 0 {
		return fmt.Errorf(ErrInvalidPath, "delete", path) // unable to delete root
	}
	var dst reflect.Value
	var base string
	if len(parts) == 1 {
		dst = reflect.ValueOf(r.Value)
		base = parts[0]
	} else {
		dir := filepath.Dir(path)
		dst = WalkTo(r.Value, dir)
		base = filepath.Base(path)
	}

	if d, ok := Value(dst, base).Interface().(Deleter); ok {
		return d.PathDelete()
	}

	if dst.Kind() == reflect.Slice {
		idx, err := strconv.Atoi(base)
		if err != nil {
			return err
		}
		dst.Set(reflect.AppendSlice(dst.Slice(0, idx), dst.Slice(idx+1, dst.Len())))
		return nil
	}

	return r.Set(path, nil)
}

func (r R) Insert(path string, idx int, v any) error {
	dst := WalkTo(r.Value, path)
	if dst.Kind() != reflect.Slice {
		return errors.New("cannot insert on non-slice value")
	}
	if idx < -1 || idx >= dst.Len() {
		return errors.New("index out of range for insert")
	}

	var rv reflect.Value
	if c, ok := FromTelepath(r, v); ok {
		rv = reflect.ValueOf(r.get(c.Path()).Interface()) // re-ValueOf to get actual type
	} else {
		rv = reflect.ValueOf(v)
	}

	if idx == -1 {
		dst.Set(reflect.Append(dst, rv))
		return nil
	}

	// TODO: use cursor.Set if v is cursor?

	nv := reflect.Append(dst, reflect.Zero(dst.Type().Elem()))
	reflect.Copy(nv.Slice(idx+1, nv.Len()), nv.Slice(idx, nv.Len()))
	nv.Index(idx).Set(rv)
	dst.Set(nv)
	return nil
}

func (r R) Get(path string) (any, error) {
	v := r.get(path)
	if !v.IsValid() {
		return nil, fmt.Errorf(ErrInvalidPath, "get", path)
	}
	if c, ok := v.Interface().(Cursor); ok {
		return c.Value()
	}
	return v.Interface(), nil
}

func (r R) Call(path string, args []any) (ret []any, err error) {
	method := WalkTo(r.Value, path)
	if !method.IsValid() {
		return nil, fmt.Errorf(ErrInvalidPath, "call", path)
	}
	if c, ok := method.Interface().(Cursor); ok {
		err = c.Call(args, &ret)
		return ret, err
	}
	return fn.Call(method.Interface(), args)
}
