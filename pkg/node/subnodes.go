package node

import "errors"

// SubnodesNode is implemented by nodes that expose an ordered list of child IDs per kind.
type SubnodesNode interface {
	GetSubnodes(kind string) []Node
	GetSubnodeIndexOf(kind, id string) (int, bool)
}

// Subnodes returns child nodes of the given kind under n.
// It returns nil when n does not implement SubnodesNode.
func Subnodes(n Node, kind string) []Node {
	if snn, ok := Unwrap[SubnodesNode](n); ok {
		return snn.GetSubnodes(kind)
	}
	return nil
}

// SubnodeIndexOf returns the index of the subnode with id under n for the given kind.
// The bool reports whether id was found; when n does not implement SubnodesNode, it returns (0, false).
func SubnodeIndexOf(n Node, kind, id string) (int, bool) {
	if snn, ok := Unwrap[SubnodesNode](n); ok {
		return snn.GetSubnodeIndexOf(kind, id)
	}
	return 0, false
}

// SubnodeCountNode is implemented by nodes that can report subnode counts without materializing the list.
type SubnodeCountNode interface {
	GetSubnodeCount(kind string) int
}

// SubnodeCount returns the number of subnodes of kind under n.
// When n does not implement SubnodeCountNode, it uses len(Subnodes(n, kind)).
func SubnodeCount(n Node, kind string) int {
	if snn, ok := Unwrap[SubnodeCountNode](n); ok {
		return snn.GetSubnodeCount(kind)
	}
	return len(Subnodes(n, kind))
}

// AppendNode is implemented by nodes that can append a subnode by kind and id.
type AppendNode interface {
	AppendSubnode(kind, id string) error
}

// AppendSubnode appends a subnode when supported.
// It emits a signal on success path and returns errors.ErrUnsupported when n does not implement AppendNode.
func AppendSubnode(n Node, kind, id string) error {
	if an, ok := Unwrap[AppendNode](n); ok {
		defer Send(n, "", kind, id)
		return an.AppendSubnode(kind, id)
	}
	return errors.ErrUnsupported
}

// InsertNode is implemented by nodes that can insert a subnode at a given index.
type InsertNode interface {
	InsertSubnode(kind string, idx int, id string) error
}

// InsertSubnode inserts a subnode at idx when supported.
// It emits a signal on success path and returns errors.ErrUnsupported when n does not implement InsertNode.
func InsertSubnode(n Node, kind string, idx int, id string) error {
	if in, ok := Unwrap[InsertNode](n); ok {
		defer Send(n, "", kind, idx, id)
		return in.InsertSubnode(kind, idx, id)
	}
	return errors.ErrUnsupported
}

// RemoveNode is implemented by nodes that can remove a subnode by id.
type RemoveNode interface {
	RemoveSubnodeID(kind, id string) error
}

// RemoveSubnodeID removes the subnode identified by kind and id when supported.
// It emits a signal on success path and returns errors.ErrUnsupported when n does not implement RemoveNode.
func RemoveSubnodeID(n Node, kind, id string) error {
	if rn, ok := Unwrap[RemoveNode](n); ok {
		defer Send(n, "", kind, id)
		return rn.RemoveSubnodeID(kind, id)
	}
	return errors.ErrUnsupported
}

// RemoveIndexNode is implemented by nodes that can remove a subnode by index.
type RemoveIndexNode interface {
	RemoveSubnodeIndex(kind string, idx int) error
}

// RemoveSubnodeIndex removes the subnode at idx for kind when supported.
// It emits a signal on success path and returns errors.ErrUnsupported when n does not implement RemoveIndexNode.
func RemoveSubnodeIndex(n Node, kind string, idx int) error {
	if rn, ok := Unwrap[RemoveIndexNode](n); ok {
		defer Send(n, "", kind, idx)
		return rn.RemoveSubnodeIndex(kind, idx)
	}
	return errors.ErrUnsupported
}

// MoveNode is implemented by nodes that can reorder subnodes by index.
type MoveNode interface {
	MoveSubnode(kind string, idx, to int) error
}

// MoveSubnode moves the subnode at idx to position to for kind when supported.
// It emits a signal on success path and returns errors.ErrUnsupported when n does not implement MoveNode.
func MoveSubnode(n Node, kind string, idx, to int) error {
	if mn, ok := Unwrap[MoveNode](n); ok {
		defer Send(n, "", kind, idx, to)
		return mn.MoveSubnode(kind, idx, to)
	}
	return errors.ErrUnsupported
}
