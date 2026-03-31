package manifold

import (
	"reflect"

	rig "github.com/progrium/rig/pkg/node"
)

type N struct {
	e rig.E
}

func FromEntity(v any) *N {
	return &N{e: rig.ToEntity(v)}
}

func (n *N) Entity() rig.E {
	return n.e
}

func (n *N) Store() rig.Store {
	return rig.GetStore(n.e)
}

func (n *N) Signaled(s Signal) {
	rig.Signaled(n.e, s)
}

func (n *N) ID() string {
	return n.e.GetID()
}

func (n *N) Name() string {
	return n.e.GetName()
}

func (n *N) Kind() string {
	return rig.Kind(n)
}

func (n *N) ComponentType() string {
	return rig.ComponentType(n)
}

func (n *N) Value() any {
	return rig.Value(n)
}

func (n *N) Attrs() []string {
	return rig.Attrs(n)
}

func (n *N) HasAttr(key string) bool {
	return rig.HasAttr(n, key)
}

func (n *N) Attr(key string) string {
	return rig.Attr(n, key)
}

func (n *N) Parent() Node {
	return FromEntity(rig.Parent(n))
}

func (n *N) Components() List {
	return L{node: n, kind: rig.Component}
}

func (n *N) AddComponent(v any) (Node, error) {
	var com *rig.Raw
	if r, ok := v.(*rig.Raw); ok {
		com = r
	} else {
		com = rig.NewComponent(v)
	}
	if err := rig.SetStore(com, rig.GetStore(n)); err != nil {
		return nil, err
	}
	err := rig.AppendEntity(n, rig.Component, com.ID)
	if err != nil {
		return nil, err
	}
	return FromEntity(com), nil
}

func (n *N) Objects() List {
	return L{node: n, kind: rig.Object}
}

func (n *N) Duplicate() Node {
	nn := rig.NewRaw(n.Name(), dupVal(n.Value()), "")
	nn.Kind = n.Kind()
	rig.SetStore(nn, rig.GetStore(n))

	for _, attr := range n.Attrs() {
		rig.SetAttr(nn, attr, n.Attr(attr))
	}

	for _, c := range n.Components().Nodes() {
		dup := c.Duplicate()
		if err := rig.AppendEntity(nn, rig.Component, dup.ID()); err != nil {
			panic(err)
		}
	}

	for _, c := range n.Objects().Nodes() {
		dup := c.Duplicate()
		if err := rig.AppendEntity(nn, rig.Object, dup.ID()); err != nil {
			panic(err)
		}
	}

	return FromEntity(nn)
}

func (n *N) SetName(name string) error {
	return rig.SetName(n, name)
}

func (n *N) SetValue(v any) error {
	return rig.SetValue(n, v)
}

func (n *N) SetParent(p Node) error {
	return rig.SetParent(n, p.ID())
}

func (n *N) SetAttr(key, val string) error {
	return rig.SetAttr(n, key, val)
}

func (n *N) DelAttr(key string) error {
	return rig.DelAttr(n, key)
}

func (n *N) Error() error {
	return rig.Error(n)
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
