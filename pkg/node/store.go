package node

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"sync"

	"github.com/mitchellh/mapstructure"
	"github.com/progrium/rig/pkg/meta"
	"github.com/progrium/rig/pkg/pointers"
	"github.com/progrium/rig/pkg/signal"
	"github.com/progrium/rig/pkg/telepath"
)

type Store interface {
	Resolve(id string, skip ...any) E
	Store(e E) error
	Destroy(e E) error
}

type StoreEntity interface {
	E
	SetStore(s Store) error
	GetStore() Store
}

func SetStore(v any, s Store) error {
	if e := ToEntity(v); e != nil {
		if se, ok := e.(StoreEntity); ok {
			return se.SetStore(s)
		}
	}
	return errors.ErrUnsupported
}

func GetStore(v any) Store {
	if e := ToEntity(v); e != nil {
		if se, ok := e.(StoreEntity); ok {
			return se.GetStore()
		}
	}
	return nil
}

type MemStore struct {
	nodes map[string]*Raw
	mu    sync.Mutex

	signal.Dispatcher[E]
}

func NewStore() *MemStore {
	return &MemStore{
		nodes: make(map[string]*Raw),
	}
}

func EmbeddedStore(n *Raw) *MemStore {
	return &MemStore{
		nodes: n.Embedded,
	}
}

// func (ss *embedStore) Signaled(s Signal) {
// 	log.Println("store signaled:", s)
// 	ss.Dispatcher.Signaled(s)
// }

// func (s *embedStore) Watch(n SignalReceiver) {
// 	s.Dispatcher.Watch(n)
// }

// func (s *embedStore) Unwatch(n SignalReceiver) {
// 	s.Dispatcher.Unwatch(n)
// }

func (s *MemStore) Destroy(e E) error {
	// TODO: walk and destroy linked
	s.mu.Lock()
	defer s.mu.Unlock()
	defer Send(e, "")
	delete(s.nodes, e.GetID())
	return nil
}

func (s *MemStore) Store(e E) error {
	if r, ok := e.(*Raw); ok {
		r.mu.Lock()
		s.mu.Lock()

		// put embedded nodes into our nodes
		if len(r.Embedded) > 0 {
			for _, embed := range r.Embedded {
				embed.root = nil // todo: is this necessary?
				embed.store = s
				s.nodes[embed.ID] = embed
				defer Send(embed, "")
			}
		}
		// unset embedded nodes and root
		// r.Embedded = nil
		r.root = nil // why?

		r.store = s
		s.nodes[r.ID] = r

		s.mu.Unlock()
		r.mu.Unlock()
		Send(e, "")
		return nil
	}
	return fmt.Errorf("unable to store entity: %v", e)
}

func (s *MemStore) Resolve(id string, skip ...any) E {
	// TODO: resolver
	s.mu.Lock()
	n, ok := s.nodes[id]
	s.mu.Unlock()
	if !ok {
		return nil
	}
	return n
}

func (s *MemStore) Export() (data []Raw, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, n := range s.nodes {
		if n.ID == "" || n.ID == RootID || strings.Contains(n.ID, "/") {
			continue
		}
		if n.ID != MainID && n.Parent == RootID {
			// only main is exportable node
			// with root as parent
			continue
		}
		snap := Snapshot(n)
		snap.Embedded = nil
		snap.Refs = make(map[string]string)
		for path, v := range pointers.From(snap.Value) {
			if other, ok := s.findValue(v); ok {
				snap.Refs[path] = other.ID

				// commenting this out since snapshot is not a deep copy this will impact the live tree.
				// luckily we ignore these values when unmarshaling if they are a pointer,
				// but it would be nice to not store unnecessary data...

				// if err := telepath.Select(snap.Value, path).Delete(); err != nil {
				// 	return nil, err
				// }
			}
		}
		data = append(data, snap)
	}
	return
}

type valuePtr struct {
	src  string
	path string
	dst  string
}

func (s *MemStore) Import(data []Raw) error {
	var ptrs []valuePtr
	var loaded []*Raw
	for _, d := range data {
		n, err := loadNode(d)
		if err != nil {
			return err
		}
		if Error(n) != nil {
			if IsComponent(n) {
				if err := DisableComponent(n); err != nil {
					panic(err)
				}
			}
		} else {
			for path, id := range n.Refs {
				ptrs = append(ptrs, valuePtr{
					src:  n.ID,
					path: path,
					dst:  id,
				})
			}
			n.Refs = nil
		}
		if err := s.Store(n); err != nil {
			return err
		}
		loaded = append(loaded, n)
	}
	for _, ptr := range ptrs {
		src := s.Resolve(ptr.src)
		dst := s.Resolve(ptr.dst)
		if err := telepath.Select(Value(src), ptr.path).Set(Value(dst)); err != nil {
			return err
		}
	}
	for _, n := range loaded {
		if p := Parent(n); p != nil && Value(n) != nil {
			if al, ok := Value(n).(ComponentAttacher); ok {
				go al.ComponentAttached(n)
			}
		}
	}
	return nil
}

func (s *MemStore) FindValue(v any) (found *Raw, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.findValue(v)
}

func (s *MemStore) findValue(v any) (found *Raw, ok bool) {
	for _, n := range s.nodes {
		if n.Value == nil || reflect.TypeOf(n.Value).Kind() == reflect.Map {
			// otherwise un-inflated map[str]any values will be
			// compared with uncomparable types
			continue
		}
		if n.Value == v {
			return n, true
		}
	}
	return
}

func loadNode(data any) (*Raw, error) {
	var raw Raw
	if err := mapstructure.Decode(data, &raw); err != nil {
		return nil, err
	}
	if factoryName, ok := raw.Attrs["_factory"]; ok {
		factory := meta.ComponentFactory(factoryName)
		raw.Value, _ = factory.New()
		return &raw, nil
	}
	raw.Refs = make(map[string]string)
	if com, ok := meta.Components[raw.Component]; raw.Component != "" && ok {
		// make new instance
		v := reflect.New(com).Interface()
		// initialize
		if i, ok := v.(Initializer); v != nil && ok {
			i.Initialize()
		}
		// decode marshaled value into new instance
		if err := decode(raw.Value, v); err != nil {
			log.Println(raw)
			return nil, err
		}
		// set new instance as value
		raw.Value = v
	} else {
		if raw.Component != "" {
			raw.Attrs["error"] = fmt.Sprintf("unable to load component: %s", raw.Component)
			raw.Value = nil
			if err := DisableComponent(raw); err != nil {
				panic(err)
			}
			return &raw, nil
		}
	}
	for k, v := range raw.Attrs {
		if strings.HasPrefix(k, "ref:") {
			raw.Refs[strings.TrimPrefix(k, "ref:")] = v
			delete(raw.Attrs, k)
		}
	}
	delete(raw.Attrs, "error") // clear any old errors
	return &raw, nil
}

func decode(input, output any) error {
	config := &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {

				// this whole hook was to take empty maps for ptr/iface values and set to nil instead.
				// but i started setting known ptrs to nil before serializing which avoids the need for this...
				// update: setting to nil wont work without a deepcopy that works for everything, back to this...
				if f.Kind() != t.Kind() && f.Kind() == reflect.Map && t.Kind() == reflect.Interface {
					// log.Println("import: skipping:", f, t)
					return nil, nil
				}

				// json serializes []byte to base64 strings but mapstructure doesn't know to reverse that
				if f.Kind() != t.Kind() && f.Kind() == reflect.String && t == reflect.TypeOf([]byte{}) {
					decoded, err := base64.StdEncoding.DecodeString(data.(string))
					if err != nil {
						return nil, err
					}
					return decoded, nil
				}

				return data, nil
			},
		),
		Result: output,
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(input)
}
