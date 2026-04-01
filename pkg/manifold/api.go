package manifold

import (
	"context"

	"github.com/progrium/rig/pkg/node"
	"github.com/progrium/rig/pkg/signal"
)

type Signal = signal.Signal[node.Node]

func Receiver(s Signal) Node {
	n, ok := s.Receiver.(Node)
	if !ok {
		panic("signal receiver is not a node")
	}
	return FromNode(n)
}

type Bus interface {

	// BusExtern API

	Name() string

	Make(name string, value any) (Node, error) // should delegate if supports path instead of name
	Find(path string) (Node, error)            // from resolve: looks at local nodes then registered buses, caches external (later? cache clearing on destroy)
	// Destroy(id string) error                   // localFind, then remoteLookup mod.Destroy, delegates

	// Export() ([]node.Raw, error)
	// Import([]node.Raw) error // triggers OnAttached

	Close() error

	Signals(ctx context.Context, ch chan Signal) error

	// BusExtern API end

	//Select(path) Cursor
	// PathNames() []string
	// PathValue(string) reflect.Value

	Resolve(id string, skip ...any) Node // TODO: fix support for []string in qtalk

	//Register(...) // will register self as upstream on bus
	Bridge(b Bus)
	BridgeUp(b Bus)
}

type Node interface {
	Node() node.Node
	NodeID() string
	Signaled(s Signal)
	Realm() node.Realm
	// Bus() Bus

	//RawRef() *node.Raw // copy, unless single threaded. always if remotenode
	// isDestroyed?

	ID() string
	Kind() string

	Name() string
	SetName(string) error

	Value() any
	SetValue(any) error

	Parent() Node           // uses Resolve
	SetParent(p Node) error // triggers OnAttached

	Attrs() []string
	Attr(key string) string
	HasAttr(key string) bool
	SetAttr(key, val string) error
	DelAttr(key string) error

	// Nodes(kind node.Kind) []Node          // uses Resolve to build slice
	// AppendNode(kind node.Kind, id string) // uses Resolve
	// NodesIndexOf(kind, id string) int
	// NodesInsert(kind string, idx int, id string) // uses Resolve to SetParent (if not ref)
	// NodesRemove(kind string, idx int)            // uses Resolve to SetParent (if not ref)
	// NodesMove(kind string, idx, to int)

	Duplicate() Node

	//Select(path ...string) telepath.Cursor
	//PathNames() (names []string)
	//PathValue(name string) reflect.Value

	// NodeDelegate API end

	AddComponent(v any) (Node, error)

	Error() error

	// PathDelete
	// PathNames
	// PathValue
	// PathCall...

	// Select(path ...string) telepath.Cursor

	// helpers

	// Destroy()
	// Path() string

	// List() List       // Siblings?
	Objects() List    // ChildNodes?
	Components() List // ComponentNodes?

	ComponentType() string
	// RemoveComponent(any)
	// Component(any)
}

type Nodes = []Node

type Iterator interface {
	Next() bool
	Node() Node
}

type List interface {
	// Parent() Node // Owner? Node? Super?
	Nodes() Nodes
	// Iter() Iterator

	Count() int
	// Index(int) Node
	IndexOf(Node) (int, bool)

	Append(Node) error // triggers OnAttached
	Remove(Node) error
	// InsertAt(int, Node) // triggers OnAttached
	// RemoveAt(int) Node

	FindByName(name string) (Node, bool)
}

type Component struct {
	com Node
}

func (embedder *Component) ComponentAttached(com node.Node) {
	embedder.com = FromNode(com)
}

func (embedder *Component) Node() Node {
	return embedder.com
}

func (embedder *Component) Object() Node {
	return FromNode(node.Parent(embedder.com))
}

func Equal(a, b Node) bool {
	return a.ID() == b.ID()
}
