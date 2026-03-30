package node

import (
	"reflect"

	"github.com/progrium/rig/pkg/entity"
	"github.com/progrium/rig/pkg/meta"
)

type Attrs map[string]string

type Children []*Raw

type Initializer interface {
	Initialize()
}

func New(name string, facets ...any) *Raw {
	n := NewID("", name, facets...)
	return n
}

func NewID(id, name string, facets ...any) *Raw {
	n := NewRaw(name, nil, id)
	n.Embedded = make(map[string]*Raw)
	addChild := func(child *Raw) {
		child.Parent = n.ID
		n.mu.Lock()

		// copy child embedded nodes into
		// our embedded nodes
		if len(child.Embedded) > 0 {
			for _, embed := range child.Embedded {
				n.Embedded[embed.ID] = embed
			}
		}
		// set child embedded nodes and store
		// to nil and set its root to us
		// before putting it in our embedded
		child.Embedded = nil
		child.store = nil
		child.root = n
		n.Embedded[child.ID] = child

		n.Children = append(n.Children, child.ID)

		n.mu.Unlock()
	}
	for _, f := range facets {
		switch facet := f.(type) {
		case Attrs:
			for k, v := range facet {
				n.Attrs[k] = v
			}
		case Children:
			for _, child := range facet {
				addChild(child)
			}
		case *Raw:
			addChild(facet)
		default:
			var com *Raw
			factory, ok := facet.(meta.Factory)
			if ok {
				value, name := factory.New()
				com = NewComponent(value)
				com.Attrs["_factory"] = name
			} else {
				com = NewComponent(facet)
				if i, ok := com.Value.(Initializer); ok {
					i.Initialize()
				}
			}
			com.Parent = n.ID
			com.root = n
			n.Components = append(n.Components, com.ID)
			n.mu.Lock()
			n.Embedded[com.ID] = com
			n.mu.Unlock()
			// parent wont be set here...
			if ca, ok := com.Value.(ComponentAttacher); ok {
				go ca.ComponentAttached(com)
			}
			if _, ok := com.Value.(Deactivator); ok {
				n.Attrs["activated"] = "false"
			}
		}
	}
	// todo: wait until this node attached to root so enable can include parents?
	for _, com := range entity.Entities(n, Component) {
		if entity.Error(com) != nil {
			continue
		}
		if err := EnableComponent(com); err != nil {
			entity.SetAttr(n, "error", err.Error())
			break
		}
	}
	return n
}

func Snapshot(e entity.Node) Raw {
	if r, ok := e.Entity().(*Raw); ok {
		v := *r
		v.Value = dupVal(v.Value)
		return v
	}
	panic("todo: build raw from non-raw based entity")
}

// DupVal uses reflection to duplicate a value. Not a deep copy!
func dupVal(v any) any {
	if v == nil {
		return nil
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface {
		el := rv.Elem()
		nv := reflect.New(el.Type())
		nv.Elem().Set(el)
		return nv.Interface()
	}
	return rv.Interface()
}
