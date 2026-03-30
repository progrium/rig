package util

import (
	"context"
	"reflect"

	"github.com/progrium/rig/pkg/manifold"
	"github.com/progrium/rig/pkg/node"
)

// deprecated
type NodeEnabler interface {
	OnEnabled()
	OnDisabled()
}

// deprecated
type ObjectToggler struct {
	IncludeChildren bool
}

func assignableTo(in map[reflect.Value]bool, t reflect.Type) (out []reflect.Value) {
	if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		t = t.Elem()
	}
	for v := range in {
		if v.Type().AssignableTo(t) {
			out = append(out, v)
		}
	}
	return
}

func callWithVals(meth reflect.Value, vals map[reflect.Value]bool) ([]reflect.Value, bool) {
	args := []reflect.Value{}
	for i := 0; i < meth.Type().NumIn(); i++ {
		argType := meth.Type().In(i)
		// if argType.Kind() == reflect.Interface && argType.Name() == "" {
		// 	// not sure what this case is
		// 	args = append(args, reflect.Zero(argType))
		// 	continue
		// }
		assignable := assignableTo(vals, argType)
		if len(assignable) == 0 {
			return []reflect.Value{}, false
		}
		switch argType.Kind() {
		case reflect.Ptr, reflect.Interface:
			args = append(args, assignable[0])
		default:
			args = append(args, assignable[0].Elem())
		}
	}
	return meth.Call(args), true
}

func (t *ObjectToggler) Activate(ctx context.Context) error {
	obj := manifold.FromContext(ctx).Parent()
	if obj == nil || obj.Kind() != node.Object {
		return nil
	}
	if t.IncludeChildren {
		for _, child := range obj.Objects().Nodes() {
			if err := node.Activate(ctx, child); err != nil {
				return err
			}
		}
	}
	t.enable(obj)
	return nil
}

func (t *ObjectToggler) enable(obj manifold.Node) {
	values := make(map[reflect.Value]bool)
	enablers := make(map[reflect.Value]bool)
	exporters := make(map[reflect.Value]bool)
	for _, com := range obj.Components().Nodes() {
		rv := reflect.ValueOf(com.Value())
		values[rv] = true
		enableMeth := rv.MethodByName("ComponentEnable")
		if enableMeth.IsValid() && !enableMeth.IsZero() {
			enablers[rv] = false
		}
		exportMeth := rv.MethodByName("ComponentExport")
		if exportMeth.IsValid() && !exportMeth.IsZero() {
			exporters[rv] = false
		}
	}

	// this doesn't cover all cases of
	// dependency trees but this is a start:
	// call possible enables, call possible exports, then again, then enables again
	for rv, called := range enablers {
		if called {
			continue
		}
		_, called = callWithVals(rv.MethodByName("ComponentEnable"), values)
		enablers[rv] = called
	}
	for rv, called := range exporters {
		if called {
			continue
		}
		var e []reflect.Value
		e, called = callWithVals(rv.MethodByName("ComponentExport"), values)
		if called {
			for _, v := range e {
				values[v] = true
			}
			exporters[rv] = called
		}
	}
	for rv, called := range exporters {
		if called {
			continue
		}
		var e []reflect.Value
		e, called = callWithVals(rv.MethodByName("ComponentExport"), values)
		if called {
			for _, v := range e {
				values[v] = true
			}
			exporters[rv] = called
		}
	}
	for rv, called := range enablers {
		if called {
			continue
		}
		_, called = callWithVals(rv.MethodByName("ComponentEnable"), values)
		enablers[rv] = called
	}

	for _, enabler := range node.GetAll[NodeEnabler](obj) {
		enabler.OnEnabled()
	}

}

func (t *ObjectToggler) Deactivate(ctx context.Context) error {
	obj := manifold.FromContext(ctx).Parent()
	if obj == nil || obj.Kind() != node.Object {
		return nil
	}
	t.disable(obj)
	if t.IncludeChildren {
		for _, child := range obj.Objects().Nodes() {
			if err := node.Deactivate(ctx, child); err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *ObjectToggler) disable(obj manifold.Node) {
	for _, com := range obj.Components().Nodes() {
		rv := reflect.ValueOf(com.Value())
		disableMeth := rv.MethodByName("ComponentDisable")
		if disableMeth.IsValid() && !disableMeth.IsZero() {
			disableMeth.Call([]reflect.Value{})
		}

		enabler, ok := com.Value().(NodeEnabler)
		if ok {
			enabler.OnDisabled()
		}
	}
}
