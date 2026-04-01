package realm

import (
	"context"

	"github.com/progrium/rig/pkg/node"
	"github.com/progrium/rig/pkg/signal"
)

type R struct {
}

type Realm interface {
	ID() string
	Node(path string)
	// Resolve finds a node by ID across all connected realms
	Resolve(id string, skipRealmIDs ...string) node.Node
	Store(n node.Node) error
	Destroy(n node.Node) error

	Import(nodes ...node.Raw) error
	Export() ([]node.Raw, error)

	Bridge(r Realm, upstream bool) error

	Signals(ctx context.Context, ch chan signal.Signal[node.Node]) error
	Close() error
}
