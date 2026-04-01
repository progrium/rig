package node

// Node is the minimal interface all node values implement.
// Node returns the underlying concrete node, and NodeID returns its stable ID.
type Node interface {
	Node() Node
	NodeID() string
}

// Nodes is a convenience slice type for collections of Node values.
type Nodes []Node

// Unwrap resolves n to its underlying node and attempts to cast it to T.
// It returns the typed value and whether the cast succeeded.
func Unwrap[T any](n Node) (T, bool) {
	node := n.Node()
	t, ok := node.(T)
	return t, ok
}
