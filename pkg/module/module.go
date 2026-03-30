package module

import (
	"fmt"
	"time"

	"github.com/progrium/rig/pkg/debouncer"
	"github.com/progrium/rig/pkg/entity"
	"github.com/progrium/rig/pkg/manifold"
	"github.com/progrium/rig/pkg/node"
	"github.com/progrium/rig/pkg/signal"
)

var SaveDebounceDuration = 500 * time.Millisecond

type M struct {
	name     string
	store    *node.Store
	provider Provider
	debounce func(func())
}

func LoadFrom(name string, provider Provider, defaultNodes ...*node.Raw) (*M, error) {
	m := New(name)
	m.provider = provider

	// Set up autosave if provider
	// @Incomplete: make this configurable?
	if m.provider != nil {
		defer m.store.Watch(m)
	}

	if m.provider != nil && m.provider.Exists() {
		d, err := m.provider.LoadAll()
		if err != nil {
			return nil, err
		}
		if err := m.store.Import(d); err != nil {
			return nil, err
		}
		return m, nil
	}

	for _, n := range defaultNodes {
		if err := m.store.Store(n); err != nil {
			return nil, err
		}
	}

	return m, nil
}

func New(name string) *M {
	m := &M{
		name:     name,
		store:    node.NewStore(),
		debounce: debouncer.New(SaveDebounceDuration),
	}
	return m
}

func (m *M) Name() string {
	return m.name
}

func (m *M) Main() manifold.Node {
	return manifold.FromEntity(m.store.Resolve("@main"))
}

func (m *M) Save() error {
	if m.provider == nil {
		return fmt.Errorf("no provider")
	}
	nodes, err := m.store.Export()
	if err != nil {
		return err
	}
	return m.provider.SaveAll(nodes)
}

func (m *M) Signaled(s signal.Signal[entity.E]) {
	// e := s.Receiver.(entity.E)
	// log.Println("event:", s.Name, e.GetName(), s.Args)
	m.debounce(func() {
		// if err := m.Save(); err != nil {
		// 	log.Println(err)
		// 	return
		// }
		// log.Println("autosaved")
	})
}
