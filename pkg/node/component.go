package node

import (
	"path"
	"reflect"
	"strings"

	"github.com/progrium/rig/pkg/library"
	"tractor.dev/toolkit-go/duplex/fn"
)

// ComponentAttacher is implemented by values that want notification when a component is attached.
type ComponentAttacher interface {
	ComponentAttached(com Node)
}

// ComponentEnabled reports whether n's "enabled" attribute is the string "true".
func ComponentEnabled(n Node) bool {
	return Attr(n, "enabled") == "true"
}

// ComponentEntity is implemented by nodes that expose a component type identifier (e.g. library path).
type ComponentEntity interface {
	GetComponentType() string
}

// ComponentType returns the component type string when n implements ComponentEntity; otherwise "".
func ComponentType(n Node) string {
	if ce, ok := Unwrap[ComponentEntity](n); ok {
		return ce.GetComponentType()
	}
	return ""
}

// NewComponent builds a component Raw node from value: normalizes to a pointer, derives the
// component id from library.ComponentID (stripping generic type parameters), sets kind TypeComponent,
// and uses the path base of that id as the node name.
func NewComponent(value any) *Raw {
	// normalize value to a ptr
	rv := reflect.ValueOf(value)
	rvv := reflect.Indirect(rv)
	ptr := reflect.New(rvv.Type())
	if rv.Kind() == reflect.Ptr {
		ptr = rv
	} else {
		ptr.Elem().Set(rv)
	}
	com := library.ComponentID(rvv.Interface())
	// remove type param for now
	if idx := strings.LastIndex(com, "["); idx != -1 {
		com = com[:idx]
	}
	n := NewRaw(path.Base(com), ptr.Interface(), "")
	n.Kind = TypeComponent
	n.Component = com
	return n
}

// EnableComponent turns the component on: no-op if already enabled. If the underlying value
// has an EnableComponent method, it is invoked first; on error the error is stored on the node
// and returned. On success (or no such method), sets attribute "enabled" to "true".
func EnableComponent(n Node) error {
	if ComponentEnabled(n) {
		return nil
	}
	rv := reflect.ValueOf(Value(n))
	enableFn := rv.MethodByName("EnableComponent")
	if enableFn.IsValid() {
		_, err := fn.Call(enableFn, []any{})
		if err != nil {
			if e := SetAttr(n, "error", err.Error()); e != nil {
				panic(e)
			}
			return err
		}
	}
	return SetAttr(n, "enabled", "true")
}

// DisableComponent turns the component off: no-op if already disabled. If the underlying value
// has a DisableComponent method, it is invoked first; on error the error is stored on the node
// and returned. On success (or no such method), sets attribute "enabled" to "false".
func DisableComponent(n Node) error {
	if !ComponentEnabled(n) {
		return nil
	}
	rv := reflect.ValueOf(Value(n))
	if rv.IsValid() {
		enableFn := rv.MethodByName("DisableComponent")
		if enableFn.IsValid() {
			_, err := fn.Call(enableFn, []any{})
			if err != nil {
				if e := SetAttr(n, "error", err.Error()); e != nil {
					panic(e)
				}
				return err
			}
		}
	}
	return SetAttr(n, "enabled", "false")
}

// TODO: clean all this up...

func GetComponent[T any](n Node, includes ...Include) (e Node, c T) {
	ee, cc := GetAllComponents[T](n, includes...)
	if len(ee) > 0 {
		return ee[0], cc[0]
	}
	return
}

func Get[T any](n Node, includes ...Include) (c T) {
	all := GetAll[T](n, includes...)
	if len(all) > 0 {
		return all[0]
	}
	return
}

func getAll[T any](n Node, includeDisabled bool) (e []Node, c []T) {
	for _, com := range Subnodes(n, TypeComponent) {
		if !ComponentEnabled(com) && !includeDisabled {
			continue
		}
		if v, ok := Value(com).(T); ok {
			c = append(c, v)
			e = append(e, com)
		}
	}
	return
}

func GetAll[T any](n Node, includes ...Include) (c []T) {
	_, c = GetAllComponents[T](n, includes...)
	return
}

func GetAllComponents[T any](n Node, includes ...Include) (e []Node, c []T) {
	include := mergeIncludes(includes)
	if !include.NotSelf {
		ee, cc := getAll[T](n, include.Disabled)
		c = append(c, cc...)
		e = append(e, ee...)
	}
	if include.Siblings {
		if p := Parent(n); IsComponent(n) && p != nil {
			ee, cc := getAll[T](p, include.Disabled)
			c = append(c, cc...)
			e = append(e, ee...)
		} else {
			for _, sibling := range Siblings(n) {
				ee, cc := getAll[T](sibling, include.Disabled)
				c = append(c, cc...)
				e = append(e, ee...)
			}
		}
	}
	if include.Parents {
		for _, parent := range Ancestors(n) {
			ee, cc := getAll[T](parent, include.Disabled)
			c = append(c, cc...)
			e = append(e, ee...)
		}
	}
	if include.Children && !include.Descendants {
		for _, child := range Subnodes(n, TypeObject) {
			ee, cc := getAll[T](child, include.Disabled)
			c = append(c, cc...)
			e = append(e, ee...)
		}
	}
	if include.Descendants && !include.Children {
		getFromDescendants[T](n, &e, &c, include.Disabled)
	}
	return
}

func getFromDescendants[T any](n Node, e *[]Node, c *[]T, includeDisabled bool) {
	for _, child := range Subnodes(n, TypeObject) {
		ee, cc := getAll[T](child, includeDisabled)
		*c = append(*c, cc...)
		*e = append(*e, ee...)
		getFromDescendants[T](child, e, c, includeDisabled)
	}
}

func mergeIncludes(includes []Include) (inc Include) {
	for _, i := range includes {
		inc.Children = inc.Children || i.Children
		inc.Descendants = inc.Descendants || i.Descendants
		inc.Parents = inc.Parents || i.Parents
		inc.Disabled = inc.Disabled || i.Disabled
		inc.Siblings = inc.Siblings || i.Siblings
		inc.NotSelf = inc.NotSelf || i.NotSelf
	}
	return
}

type Include struct {
	Children    bool
	Siblings    bool
	Parents     bool
	Descendants bool
	Disabled    bool
	NotSelf     bool
}

func zero[T any]() T {
	var zero T
	return zero
}
