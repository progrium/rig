package manifold

import (
	"reflect"

	"github.com/progrium/rig/pkg/entity"
	"github.com/progrium/rig/pkg/node"
)

type N struct {
	e entity.E
}

func FromEntity(v any) *N {
	return &N{e: entity.ToEntity(v)}
}

func (n *N) Entity() entity.E {
	return n.e
}

func (n *N) Store() entity.Store {
	return entity.GetStore(n.e)
}

func (n *N) Signaled(s Signal) {
	entity.Signaled(n.e, s)
}

func (n *N) ID() string {
	return n.e.GetID()
}

func (n *N) Name() string {
	return n.e.GetName()
}

func (n *N) Kind() string {
	return entity.Kind(n)
}

func (n *N) ComponentType() string {
	return entity.ComponentType(n)
}

func (n *N) Value() any {
	return entity.Value(n)
}

func (n *N) Attrs() []string {
	return entity.Attrs(n)
}

func (n *N) HasAttr(key string) bool {
	return entity.HasAttr(n, key)
}

func (n *N) Attr(key string) string {
	return entity.Attr(n, key)
}

func (n *N) Parent() Node {
	return FromEntity(entity.Parent(n))
}

func (n *N) Components() List {
	return L{node: n, kind: node.Component}
}

func (n *N) AddComponent(v any) (Node, error) {
	var com *node.Raw
	if r, ok := v.(*node.Raw); ok {
		com = r
	} else {
		com = node.NewComponent(v)
	}
	if err := entity.SetStore(com, entity.GetStore(n)); err != nil {
		return nil, err
	}
	err := entity.AppendEntity(n, node.Component, com.ID)
	if err != nil {
		return nil, err
	}
	return FromEntity(com), nil
}

func (n *N) Objects() List {
	return L{node: n, kind: node.Object}
}

func (n *N) Duplicate() Node {
	nn := node.NewRaw(n.Name(), dupVal(n.Value()), "")
	nn.Kind = n.Kind()
	entity.SetStore(nn, entity.GetStore(n))

	for _, attr := range n.Attrs() {
		entity.SetAttr(nn, attr, n.Attr(attr))
	}

	for _, c := range n.Components().Nodes() {
		dup := c.Duplicate()
		if err := entity.AppendEntity(nn, node.Component, dup.ID()); err != nil {
			panic(err)
		}
	}

	for _, c := range n.Objects().Nodes() {
		dup := c.Duplicate()
		if err := entity.AppendEntity(nn, node.Object, dup.ID()); err != nil {
			panic(err)
		}
	}

	return FromEntity(nn)
}

func (n *N) SetName(name string) error {
	return entity.SetName(n, name)
}

func (n *N) SetValue(v any) error {
	return entity.SetValue(n, v)
}

func (n *N) SetParent(p Node) error {
	return entity.SetParent(n, p.ID())
}

func (n *N) SetAttr(key, val string) error {
	return entity.SetAttr(n, key, val)
}

func (n *N) DelAttr(key string) error {
	return entity.DelAttr(n, key)
}

func (n *N) Error() error {
	return entity.Error(n)
}

// DupVal uses reflection to duplicate a value
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
