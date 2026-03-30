package module

import "github.com/progrium/rig/pkg/node"

type Provider interface {
	Exists() bool
	LoadAll() ([]node.Raw, error)
	SaveAll(nodes []node.Raw) error
	Save(n node.Raw) error
	Close() error
}
