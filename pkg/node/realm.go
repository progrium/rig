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

// Realm is a node registry: resolve by id, persist nodes into the realm, and tear them down.
// Implementations are expected to keep ids consistent with Node.NodeID.
type Realm interface {
	// Resolve returns the node for id, or nil if unknown. skip is reserved for future use.
	Resolve(id string, skip ...any) Node
	// Store registers n in the realm (typically a *Raw) and wires it for lookup and lifecycle.
	Store(n Node) error
	// Destroy removes n from the realm and releases associated resources.
	Destroy(n Node) error
}

// RealmNode is implemented by nodes that belong to a Realm (lookup, store, destroy).
type RealmNode interface {
	Node
	SetRealm(s Realm) error
	GetRealm() Realm
}

// SetRealm attaches realm s to n when n implements RealmNode; otherwise returns errors.ErrUnsupported.
func SetRealm(n Node, s Realm) error {
	if rn, ok := Unwrap[RealmNode](n); ok {
		return rn.SetRealm(s)
	}
	return errors.ErrUnsupported
}

// GetRealm returns n's realm when n implements RealmNode; otherwise nil.
func GetRealm(n Node) Realm {
	if rn, ok := Unwrap[RealmNode](n); ok {
		return rn.GetRealm()
	}
	return nil
}

// BasicRealm is an in-memory Realm backed by a map of node id to *Raw.
// It embeds signal.Dispatcher[Node] so stored nodes can participate in signaling.
// The mutex protects the nodes map; callers should not mutate the map directly.
type BasicRealm struct {
	nodes  map[string]*Raw
	strict bool // if true, import will fail if component type is not found
	mu     sync.Mutex

	signal.Dispatcher[Node]
}

// NewRealm returns an empty BasicRealm ready for Store/Resolve/Destroy.
func NewRealm(strict bool) *BasicRealm {
	return &BasicRealm{
		nodes:  make(map[string]*Raw),
		strict: strict,
	}
}

func embeddedStore(n *Raw) *BasicRealm {
	return &BasicRealm{
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

func (s *BasicRealm) Destroy(n Node) error {
	// TODO: walk and destroy linked
	s.mu.Lock()
	defer s.mu.Unlock()
	defer Send(n, "")
	delete(s.nodes, n.NodeID())
	return nil
}

func (s *BasicRealm) Store(n Node) error {
	if r, ok := Unwrap[*Raw](n); ok {
		r.mu.Lock()
		s.mu.Lock()

		// put embedded nodes into our nodes
		if len(r.Embedded) > 0 {
			for _, embed := range r.Embedded {
				embed.root = nil // todo: is this necessary?
				embed.realm = s
				s.nodes[embed.ID] = embed
				defer Send(embed, "")
			}
		}
		// unset embedded nodes and root
		// r.Embedded = nil
		r.root = nil // why?

		r.realm = s
		s.nodes[r.ID] = r

		s.mu.Unlock()
		r.mu.Unlock()
		Send(n, "")
		return nil
	}
	return fmt.Errorf("unable to store node: %v", n.NodeID())
}

func (s *BasicRealm) Resolve(id string, skip ...any) Node {
	// TODO: resolver
	s.mu.Lock()
	n, ok := s.nodes[id]
	s.mu.Unlock()
	if !ok {
		return nil
	}
	return n
}

func (s *BasicRealm) Export() (data []Raw, err error) {
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

func (s *BasicRealm) Import(data []Raw) error {
	var ptrs []valuePtr
	var loaded []*Raw
	for _, d := range data {
		n, err := loadNode(d, s.strict)
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

func (s *BasicRealm) FindValue(v any) (found *Raw, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.findValue(v)
}

func (s *BasicRealm) findValue(v any) (found *Raw, ok bool) {
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

func loadNode(data any, strict bool) (*Raw, error) {
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
		if raw.Component != "" && strict {
			raw.Attrs["error"] = fmt.Sprintf("unable to load component: %s", raw.Component)
			raw.Value = nil
			if err := DisableComponent(&raw); err != nil {
				panic(err)
			}
			return &raw, fmt.Errorf("unable to load component: %s", raw.Component)
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
