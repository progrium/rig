package manifold

import rig "github.com/progrium/rig/pkg/node"

type L struct {
	node Node
	kind string
}

// func (l L) Parent() Node {
// 	return l.node
// }

func (l L) Nodes() (nodes Nodes) {
	for _, e := range rig.Entities(l.node, l.kind) {
		nodes = append(nodes, FromEntity(e))
	}
	return
}

//Iter() Iterator

func (l L) Count() int {
	return rig.EntityCount(l.node, l.kind)
}

// func (l L) Index(int) Node {}

func (l L) IndexOf(n Node) (int, bool) {
	return rig.EntityIndexOf(l.node, l.kind, n.ID())
}

func (l L) Append(n Node) error {
	return rig.AppendEntity(l.node, l.kind, n.ID())
}

func (l L) Remove(n Node) error {
	return rig.RemoveEntity(l.node, l.kind, n.ID())
}

func (l L) FindByName(name string) (Node, bool) {
	for _, e := range rig.Entities(l.node, l.kind) {
		if rig.Name(e) == name {
			return FromEntity(e), true
		}
	}
	return nil, false
}

// InsertAt(int, Node)
// RemoveAt(int) Node
