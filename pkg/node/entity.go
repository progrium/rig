package node

import (
	"github.com/progrium/rig/pkg/entity"
)

func (r *Raw) Entity() entity.E {
	return r
}

func (r *Raw) SetStore(s entity.Store) error {
	if s == nil {
		return nil
	}
	r.store = s
	return s.Store(r)
}

func (r *Raw) GetStore() entity.Store {
	if r.store != nil {
		return r.store
	}
	root := r.root
	if root != nil {
		for root.root != nil {
			root = root.root
		}
		if root != nil {
			return entity.GetStore(root)
		}
	}
	// we're root with no store
	// so we make one on demand
	r.store = EmbeddedStore(r)
	return r.store
}

func (r *Raw) GetID() string {
	return r.ID
}

func (r *Raw) GetKind() string {
	return r.Kind
}

func (r *Raw) GetName() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.Name
}

func (r *Raw) GetComponentType() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.Component
}

func (r *Raw) SetName(name string) error {
	r.mu.Lock()
	r.Name = name
	r.N++
	r.mu.Unlock()
	return nil
}

func (r *Raw) GetAttrs() (keys []string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for k := range r.Attrs {
		keys = append(keys, k)
	}
	return
}

func (r *Raw) GetAttr(key string) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.Attrs[key]
}

func (r *Raw) SetAttr(key, value string) error {
	r.mu.Lock()
	r.Attrs[key] = value
	r.N++
	r.mu.Unlock()
	return nil
}

func (r *Raw) DelAttr(key string) error {
	r.mu.Lock()
	delete(r.Attrs, key)
	r.N++
	r.mu.Unlock()
	return nil
}

func (r *Raw) GetValue() any {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.Value
}

func (r *Raw) SetValue(v any) error {
	r.mu.Lock()
	r.Value = v
	r.N++
	r.mu.Unlock()
	return nil
}

func (r *Raw) GetParentID() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.Parent
}

func (r *Raw) GetParent() entity.E {
	store := entity.GetStore(r)
	if store == nil {
		panic("store not set on entity")
	}
	parent := r.GetParentID()
	if parent == "" {
		return nil
	}
	return store.Resolve(parent)
}

func (r *Raw) SetParent(id string) error {
	r.mu.Lock()
	r.Parent = id
	r.N++
	r.mu.Unlock()
	if ca, ok := r.Value.(ComponentAttacher); ok {
		ca.ComponentAttached(r)
	}
	return nil
}

func (r *Raw) nodes(kind string) *[]string {
	switch kind {
	case Object:
		return &r.Children
	case Component:
		return &r.Components
	default:
		panic("unknown kind")
	}
}

func (r *Raw) GetEntities(kind string) (ents []entity.E) {
	store := entity.GetStore(r)
	if store == nil {
		panic("store not set on entity")
	}
	r.mu.Lock()
	nodes := r.nodes(kind)
	for _, id := range *nodes {
		if hh := store.Resolve(id); hh != nil {
			ents = append(ents, hh)
		}
	}
	r.mu.Unlock()
	return
}

func (r *Raw) GetEntityIndexOf(kind, id string) (int, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	nodes := r.nodes(kind)
	for idx, nid := range *nodes {
		if nid == id {
			return idx, true
		}
	}
	return 0, false
}

func (r *Raw) AppendEntity(kind, id string) error {
	store := entity.GetStore(r)
	r.mu.Lock()
	nodes := r.nodes(kind)
	*nodes = append(*nodes, id)
	r.N++
	r.mu.Unlock()
	if store != nil {
		ent := store.Resolve(id)
		return entity.SetParent(ent, r.ID)
	}
	return nil
}

func (r *Raw) InsertEntityAt(kind string, idx int, id string) error {
	return nil // todo
}

func (r *Raw) RemoveEntity(kind, id string) error {
	idx, ok := r.GetEntityIndexOf(kind, id)
	if !ok {
		// removing entity not in there is no-op?
		return nil
	}
	return r.RemoveEntityAt(kind, idx)
}

func (r *Raw) RemoveEntityAt(kind string, idx int) error {
	r.mu.Lock()
	nodes := r.nodes(kind)
	*nodes = append((*nodes)[:idx], (*nodes)[idx+1:]...)
	r.N++
	r.mu.Unlock()
	return nil
}

func (r *Raw) MoveEntity(kind string, idx, to int) error {
	return nil // todo
}

// func (r *Raw) GetError() error {
// 	r.mu.Lock()
// 	defer r.mu.Unlock()
// 	return r.err
// }

// func (r *Raw) SetError(err error) error {
// 	r.mu.Lock()
// 	defer r.mu.Unlock()
// 	r.err = err
// 	return nil
// }
