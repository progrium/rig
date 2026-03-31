package node

import (
	"strings"
	"sync"

	"github.com/progrium/rig/pkg/signal"
	"github.com/rs/xid"
)

const (
	RootID = "@root"
	MainID = "@main"
)

func NextID() string {
	return xid.New().String()
}

type Raw struct {
	ID   string
	Kind string // obj, com
	Bus  string `json:",omitempty"`

	Name  string
	Value any

	Component string `json:",omitempty"` // component type
	Parent    string
	Attrs     map[string]string

	Children   []string
	Components []string

	// only used by root node of raw tree
	Embedded map[string]*Raw `json:",omitempty"`

	// used when marshaling out of a module
	// where value can point to other nodes
	Refs map[string]string `json:",omitempty"`

	N uint

	store Store
	root  *Raw
	mu    sync.Mutex
	// err   error

	signal.Dispatcher[E]
}

func NewRaw(name string, value any, id string) *Raw {
	if id == "" {
		id = NextID()
		if strings.HasPrefix(name, "@") {
			id = name
		}
	}
	return &Raw{
		Kind:  Object,
		ID:    id,
		Name:  name,
		Value: value,
		Attrs: make(map[string]string),
	}
}
