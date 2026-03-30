package manifold

import "github.com/progrium/rig/pkg/entity"

type L struct {
	node Node
	kind string
}

// func (l L) Parent() Node {
// 	return l.node
// }

func (l L) Nodes() (nodes Nodes) {
	for _, e := range entity.Entities(l.node, l.kind) {
		nodes = append(nodes, FromEntity(e))
	}
	return
}

//Iter() Iterator

func (l L) Count() int {
	return entity.EntityCount(l.node, l.kind)
}

// func (l L) Index(int) Node {}

func (l L) IndexOf(n Node) (int, bool) {
	return entity.EntityIndexOf(l.node, l.kind, n.ID())
}

func (l L) Append(n Node) error {
	return entity.AppendEntity(l.node, l.kind, n.ID())
}

func (l L) Remove(n Node) error {
	return entity.RemoveEntity(l.node, l.kind, n.ID())
}

func (l L) FindByName(name string) (Node, bool) {
	for _, e := range entity.Entities(l.node, l.kind) {
		if entity.Name(e) == name {
			return FromEntity(e), true
		}
	}
	return nil, false
}

// InsertAt(int, Node)
// RemoveAt(int) Node
