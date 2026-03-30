package pointers

import (
	"reflect"
	"strings"

	"github.com/progrium/rig/pkg/telepath"
)

// PointersFrom walks a value's structure and returns a map of paths to any pointers found.
func From(v any) (ptrs map[string]any) {
	ptrs = make(map[string]any)
	if v == nil {
		return
	}
	walk(reflect.ValueOf(v), []string{}, func(v reflect.Value, parent reflect.Value, path []string) error {
		if v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
			if v.IsNil() {
				return nil
			}
			ptrs[strings.Join(path, "/")] = v.Interface()
		}
		return nil
	})
	return
}

// walk uses telepath to walk a data structure calling visitor for each node.
// TODO: should this be part of telepath?
func walk(v reflect.Value, path []string, visitor func(v reflect.Value, parent reflect.Value, path []string) error) error {
	for _, k := range telepath.Names(v.Interface()) {
		subpath := append(path, k)
		vv := telepath.Value(v, k)
		if !vv.IsValid() || vv.IsZero() {
			continue
		}
		if err := visitor(vv, v, subpath); err != nil {
			return err
		}
		if err := walk(vv, subpath, visitor); err != nil {
			return err
		}
	}
	return nil
}
