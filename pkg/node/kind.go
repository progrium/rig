package node

const (
	// TypeObject identifies object nodes.
	TypeObject string = "obj"
	// TypeComponent identifies component nodes.
	TypeComponent string = "com"
)

// IsComponent reports whether n is a component node.
func IsComponent(n Node) bool {
	return Kind(n) == TypeComponent
}

// IsObject reports whether n is an object node.
func IsObject(n Node) bool {
	return Kind(n) == TypeObject
}

// KindNode is implemented by nodes that expose a kind string.
type KindNode interface {
	Node
	GetKind() string
}

// Kind returns the node kind.
// It panics if the underlying node does not implement KindNode.
func Kind(n Node) string {
	if kn, ok := Unwrap[KindNode](n); ok {
		return kn.GetKind()
	}
	panic("nodes must implement Kind")
}
