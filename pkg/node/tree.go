package node

import "errors"

// ParentNode is implemented by nodes that track a parent relationship.
type ParentNode interface {
	Node
	GetParentID() string
	GetParent() Node
}

// Parent returns n's parent when n implements ParentNode.
// It returns nil when parent information is unavailable.
func Parent(n Node) Node {
	if pnn, ok := Unwrap[ParentNode](n); ok {
		return pnn.GetParent()
	}
	return nil
}

// ParentID returns n's parent node ID when available.
// It returns an empty string when n does not implement ParentNode.
func ParentID(n Node) string {
	if pnn, ok := Unwrap[ParentNode](n); ok {
		return pnn.GetParentID()
	}
	return ""
}

// HasParent reports whether n has a non-empty parent ID.
func HasParent(n Node) bool {
	return ParentID(n) != ""
}

// Ancestors returns n's parent chain from closest parent to root.
// If n does not implement ParentNode, it returns an empty slice.
func Ancestors(n Node) (parents []Node) {
	if pnn, ok := Unwrap[ParentNode](n); ok {
		cur := pnn.GetParent()
		for cur != nil {
			parents = append(parents, cur)
			cur = Parent(cur)
		}
	}
	return
}

// SetParentNode is implemented by nodes whose parent ID can be changed.
type SetParentNode interface {
	Node
	SetParent(id string) error
}

// SetParent updates n's parent ID when supported.
// It emits a signal on success path and returns errors.ErrUnsupported
// when n does not implement SetParentNode.
func SetParent(n Node, id string) error {
	if snn, ok := Unwrap[SetParentNode](n); ok {
		defer Send(n, "", id)
		return snn.SetParent(id)
	}
	return errors.ErrUnsupported
}

// Root returns the top-most ancestor of n.
// It returns nil when n has no parent API or no parent chain.
func Root(n Node) Node {
	if _, ok := Unwrap[ParentNode](n); !ok {
		return nil
	}
	cur := n
	var root Node
	for cur != nil {
		cur = Parent(cur)
		if cur != nil {
			root = cur
		}
	}
	return root
}

// Siblings returns nodes that share n's parent and kind, excluding n itself.
// It returns nil when n has no parent API, and an empty slice when parent exists
// but there are no matching siblings.
func Siblings(n Node) (siblings []Node) {
	if _, ok := Unwrap[ParentNode](n); !ok {
		return nil
	}
	p := Parent(n)
	if p == nil {
		return
	}
	for _, child := range Subnodes(p, Kind(n)) {
		if child.NodeID() == n.NodeID() {
			continue
		}
		siblings = append(siblings, child)
	}
	return
}
