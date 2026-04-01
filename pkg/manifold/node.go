package manifold

import (
	"reflect"

	node "github.com/progrium/rig/pkg/node"
)

type N struct {
	n node.Node
}

func FromNode(n node.Node) *N {
	return &N{n: n}
}

func (n *N) Node() node.Node {
	return n.n
}

func (n *N) Realm() node.Realm {
	return node.GetRealm(n.n)
}

func (n *N) Signaled(s Signal) {
	node.Signaled(n.n, s)
}

func (n *N) NodeID() string {
	return n.n.NodeID()
}

// TODO: deprecated?
func (n *N) ID() string {
	return n.n.NodeID()
}

func (n *N) Name() string {
	return node.Name(n.n)
}

func (n *N) Kind() string {
	return node.Kind(n.n)
}

func (n *N) ComponentType() string {
	return node.ComponentType(n.n)
}

func (n *N) Value() any {
	return node.Value(n.n)
}

func (n *N) Attrs() []string {
	return node.Attrs(n.n)
}

func (n *N) HasAttr(key string) bool {
	return node.HasAttr(n.n, key)
}

func (n *N) Attr(key string) string {
	return node.Attr(n.n, key)
}

func (n *N) Parent() Node {
	return FromNode(node.Parent(n.n))
}

func (n *N) Components() List {
	return L{node: n, kind: node.TypeComponent}
}

func (n *N) AddComponent(v any) (Node, error) {
	var com *node.Raw
	if r, ok := v.(*node.Raw); ok {
		com = r
	} else {
		com = node.NewComponent(v)
	}
	if err := node.SetRealm(com, node.GetRealm(n)); err != nil {
		return nil, err
	}
	err := node.AppendSubnode(n, node.TypeComponent, com.ID)
	if err != nil {
		return nil, err
	}
	return FromNode(com), nil
}

func (n *N) Objects() List {
	return L{node: n, kind: node.TypeObject}
}

func (n *N) Duplicate() Node {
	nn := node.NewRaw(n.Name(), dupVal(n.Value()), "")
	nn.Kind = n.Kind()
	node.SetRealm(nn, node.GetRealm(n))

	for _, attr := range n.Attrs() {
		node.SetAttr(nn, attr, n.Attr(attr))
	}

	for _, c := range n.Components().Nodes() {
		dup := c.Duplicate()
		if err := node.AppendSubnode(nn, node.TypeComponent, dup.ID()); err != nil {
			panic(err)
		}
	}

	for _, c := range n.Objects().Nodes() {
		dup := c.Duplicate()
		if err := node.AppendSubnode(nn, node.TypeObject, dup.ID()); err != nil {
			panic(err)
		}
	}

	return FromNode(nn)
}

func (n *N) SetName(name string) error {
	return node.SetName(n, name)
}

func (n *N) SetValue(v any) error {
	return node.SetValue(n, v)
}

func (n *N) SetParent(p Node) error {
	return node.SetParent(n, p.ID())
}

func (n *N) SetAttr(key, val string) error {
	return node.SetAttr(n, key, val)
}

func (n *N) DelAttr(key string) error {
	return node.DelAttr(n, key)
}

func (n *N) Error() error {
	return node.Error(n)
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
