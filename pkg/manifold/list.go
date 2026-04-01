package manifold

import "github.com/progrium/rig/pkg/node"

type L struct {
	node Node
	kind string
}

// func (l L) Parent() Node {
// 	return l.node
// }

func (l L) Nodes() (nodes Nodes) {
	for _, e := range node.Subnodes(l.node, l.kind) {
		nodes = append(nodes, FromNode(e))
	}
	return
}

//Iter() Iterator

func (l L) Count() int {
	return node.SubnodeCount(l.node, l.kind)
}

// func (l L) Index(int) Node {}

func (l L) IndexOf(n Node) (int, bool) {
	return node.SubnodeIndexOf(l.node, l.kind, n.ID())
}

func (l L) Append(n Node) error {
	return node.AppendSubnode(l.node, l.kind, n.ID())
}

func (l L) Remove(n Node) error {
	return node.RemoveSubnodeID(l.node, l.kind, n.ID())
}

func (l L) FindByName(name string) (Node, bool) {
	for _, e := range node.Subnodes(l.node, l.kind) {
		if node.Name(e) == name {
			return FromNode(e), true
		}
	}
	return nil, false
}

// InsertAt(int, Node)
// RemoveAt(int) Node
