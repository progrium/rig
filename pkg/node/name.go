package node

import "errors"

// NameNode is implemented by nodes that expose a name.
type NameNode interface {
	Node
	GetName() string
}

// Name returns the node name when n implements NameNode.
// It returns an empty string when no name is available.
func Name(n Node) string {
	if nn, ok := Unwrap[NameNode](n); ok {
		return nn.GetName()
	}
	return ""
}

// SetNameNode is implemented by nodes whose name can be updated.
type SetNameNode interface {
	Node
	SetName(name string) error
}

// SetName updates a node's name when supported.
// It is a no-op if the name is already set, emits a signal on success path,
// and returns errors.ErrUnsupported when n does not implement SetNameNode.
func SetName(n Node, name string) error {
	if Name(n) == name {
		return nil
	}
	if snn, ok := Unwrap[SetNameNode](n); ok {
		defer Send(n, "", name)
		return snn.SetName(name)
	}
	return errors.ErrUnsupported
}
