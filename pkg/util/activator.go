package util

import (
	"context"
	"fmt"
	"log"
	"reflect"

	"github.com/progrium/rig/pkg/depgraph"
	"github.com/progrium/rig/pkg/manifold"
	"github.com/progrium/rig/pkg/node"
	"tractor.dev/toolkit-go/engine"
)

type Activator interface {
	Activate(ctx context.Context) error
}

type Deactivator interface {
	Activator
	Deactivate(ctx context.Context) error
}

type ObjectActivator struct {
	IncludeChildren bool
}

func (oa *ObjectActivator) ActivateObject(ctx context.Context) error {
	obj := manifold.FromContext(ctx)
	stateful := false

	if oa.IncludeChildren {
		for _, child := range obj.Objects().Nodes() {
			if err := node.Activate(ctx, child); err != nil {
				return err
			}
		}
	}

	// resolve depgraph of all object components
	res, missing, err := resolveComponentTypes(obj)
	if err != nil {
		return err
	}
	if len(missing) > 0 {
		log.Println("WARN: missing:", missing)
	}
	asm, _ := engine.New()
	// in resolved order: Assemble, Activate(if activator), add Provides if any
	for _, t := range res {
		com := componentFromType(obj, t)
		if com == nil {
			continue
		}
		v := node.Value(com)
		if err := asm.Add(v); err != nil {
			panic(err)
		}
		if err := asm.Assemble(v); err != nil {
			panic(err)
		}
		if activator, ok := v.(Activator); ok {
			if err := activator.Activate(node.Context(ctx, com)); err != nil {
				if e := node.SetAttr(obj, "error", fmt.Sprintf("%s: %s", node.Name(com), err.Error())); e != nil {
					panic(err)
				}
				return err
			}
		}
		if _, ok := v.(Deactivator); ok {
			stateful = true
		}

		provides := reflect.ValueOf(v).MethodByName("Provides")
		if provides.IsValid() && !provides.IsZero() {
			rets := provides.Call([]reflect.Value{})
			if len(rets) > 0 {
				for _, retval := range rets {
					if err := asm.Add(retval.Interface()); err != nil {
						panic(err)
					}
				}
			}
		}

	}

	if stateful {
		if err := node.SetAttr(obj, "activated", "true"); err != nil {
			return err
		}
	}

	return nil
}

func (oa *ObjectActivator) DeactivateObject(ctx context.Context) error {
	obj := manifold.FromContext(ctx)

	if oa.IncludeChildren {
		for _, child := range obj.Objects().Nodes() {
			if err := node.Deactivate(ctx, child); err != nil {
				return err
			}
		}
	}

	// resolve depgraph order of all object components, then deactivate in reverse
	res, _, err := resolveComponentTypes(obj)
	if err != nil {
		return err
	}
	for i := len(res) - 1; i >= 0; i-- {
		com := componentFromType(obj, res[i])
		if com == nil {
			continue
		}
		deactivator, ok := node.Value(com).(Deactivator)
		if ok {
			if err := deactivator.Deactivate(node.Context(ctx, com)); err != nil {
				if e := node.SetAttr(obj, "error", fmt.Sprintf("%s: %s", node.Name(com), err.Error())); e != nil {
					panic(err)
				}
				return err
			}
		}
	}
	return node.SetAttr(obj, "activated", "false")
}

func resolveComponentTypes(obj node.Node) (resolved []reflect.Type, missing []reflect.Type, err error) {
	var rv []reflect.Type
	for _, com := range node.Subnodes(obj, node.TypeComponent) {
		if !node.ComponentEnabled(com) {
			continue
		}
		rv = append(rv, reflect.TypeOf(node.Value(com)))
	}
	graph := depgraph.NewDependencyGraph(rv, "Assemble", "Provides")
	resolved, err = graph.Resolve()
	for t := range graph.Missing {
		missing = append(missing, t)
	}
	return
}

func componentFromType(obj node.Node, t reflect.Type) node.Node {
	for _, com := range node.Subnodes(obj, node.TypeComponent) {
		if !node.ComponentEnabled(com) {
			continue
		}
		v := node.Value(com)
		if reflect.TypeOf(v) == t {
			return com
		}
	}
	return nil
}
