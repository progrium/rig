package node

import (
	"path"
	"reflect"
	"strings"

	"github.com/progrium/rig/pkg/library"
	"tractor.dev/toolkit-go/duplex/fn"
)

type ComponentAttacher interface {
	ComponentAttached(com Node)
}

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
	n.Kind = Component
	n.Component = com
	return n
}

// func GetComponent[T any](n any) *T {
// 	// check value first in case this is a component node
// 	if v, ok := Value(n).(*T); ok {
// 		return v
// 	}
// 	for _, c := range Entities(n, Component) {
// 		if v, ok := Value(c).(*T); ok {
// 			return v
// 		}
// 	}
// 	return nil
// }

// func GetComponentInChildren[T any](n any) (c *T) {
// 	for _, child := range Entities(n, Object) {
// 		for _, com := range Entities(child, Component) {
// 			if v, ok := Value(com).(*T); ok {
// 				return v
// 			}
// 		}
// 	}
// 	return
// }

// func GetComponents[T any](n any) (c []*T) {
// 	for _, com := range Entities(n, Component) {
// 		if v, ok := Value(com).(*T); ok {
// 			c = append(c, v)
// 		}
// 	}
// 	return
// }

// func GetComponentsInChildren[T any](n any) (c []*T) {
// 	for _, child := range Entities(n, Object) {
// 		for _, com := range Entities(child, Component) {
// 			if v, ok := Value(com).(*T); ok {
// 				c = append(c, v)
// 			}
// 		}
// 	}
// 	return
// }

func GetComponent[T any](n any, includes ...Include) (e E, c T) {
	ee, cc := GetAllComponents[T](n, includes...)
	if len(ee) > 0 {
		return ee[0], cc[0]
	}
	return
}

func Get[T any](n any, includes ...Include) (c T) {
	all := GetAll[T](n, includes...)
	if len(all) > 0 {
		return all[0]
	}
	return
}

func getAll[T any](n any, includeDisabled bool) (e []E, c []T) {
	for _, com := range Entities(n, Component) {
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

func GetAll[T any](n any, includes ...Include) (c []T) {
	_, c = GetAllComponents[T](n, includes...)
	return
}

func GetAllComponents[T any](n any, includes ...Include) (e []E, c []T) {
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
		for _, parent := range Parents(n) {
			ee, cc := getAll[T](parent, include.Disabled)
			c = append(c, cc...)
			e = append(e, ee...)
		}
	}
	if include.Children && !include.Descendants {
		for _, child := range Entities(n, Object) {
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

func getFromDescendants[T any](n any, e *[]E, c *[]T, includeDisabled bool) {
	for _, child := range Entities(n, Object) {
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

// deprecated, use GetComponent
func ComponentNode[T any](n any) (E, T) {
	for _, c := range Entities(n, Component) {
		if v, ok := Value(c).(T); ok {
			return c, v
		}
	}
	return nil, zero[T]()
}

func EnableComponent(n any) error {
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

func DisableComponent(n any) error {
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
