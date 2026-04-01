package node

import "fmt"

func (rn *Raw) Node() Node {
	return rn
}

func (rn *Raw) SetRealm(r Realm) error {
	if r == nil {
		return nil
	}
	rn.mu.Lock()
	rn.realm = r
	rn.mu.Unlock()
	return r.Store(rn)
}

func (rn *Raw) GetRealm() Realm {
	rn.mu.RLock()
	if rn.realm != nil {
		defer rn.mu.RUnlock()
		return rn.realm
	}
	root := rn.root
	if root != nil {
		for root.root != nil {
			root = root.root
		}
		if root != nil {
			rn.mu.RUnlock()
			return GetRealm(root)
		}
	}
	// we're root with no store
	// so we make one on demand
	rn.realm = embeddedStore(rn)
	defer rn.mu.RUnlock()
	return rn.realm
}

func (r *Raw) NodeID() string {
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
	r.mu.RLock()
	defer r.mu.RUnlock()
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
	r.mu.RLock()
	defer r.mu.RUnlock()
	for k := range r.Attrs {
		keys = append(keys, k)
	}
	return
}

func (r *Raw) GetAttr(key string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()
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
	r.mu.RLock()
	defer r.mu.RUnlock()
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
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Parent
}

func (r *Raw) GetParent() Node {
	realm := GetRealm(r)
	if realm == nil {
		panic("realm not set on node")
	}
	id := r.GetParentID()
	if id == "" {
		return nil
	}
	return realm.Resolve(id)
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
	case TypeObject:
		return &r.Children
	case TypeComponent:
		return &r.Components
	default:
		panic("unknown kind")
	}
}

func (r *Raw) GetSubnodes(kind string) (subnodes []Node) {
	realm := GetRealm(r)
	if realm == nil {
		panic("realm not set on node")
	}
	r.mu.RLock()
	nodes := r.nodes(kind)
	for _, id := range *nodes {
		if hh := realm.Resolve(id); hh != nil {
			subnodes = append(subnodes, hh)
		}
	}
	r.mu.RUnlock()
	return
}

func (r *Raw) GetSubnodeIndexOf(kind, id string) (int, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	nodes := r.nodes(kind)
	for idx, nid := range *nodes {
		if nid == id {
			return idx, true
		}
	}
	return 0, false
}

func (r *Raw) AppendSubnode(kind, id string) error {
	realm := GetRealm(r)
	r.mu.Lock()
	nodes := r.nodes(kind)
	*nodes = append(*nodes, id)
	r.N++
	r.mu.Unlock()
	if realm != nil {
		if node := realm.Resolve(id); node != nil {
			return SetParent(node, r.ID)
		}
	}
	return nil
}

func (r *Raw) InsertSubnode(kind string, idx int, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	nodes := r.nodes(kind)
	if idx < 0 || idx > len(*nodes) {
		return fmt.Errorf("index out of bounds: %d", idx)
	}

	// Insert id at idx in the slice
	*nodes = append(*nodes, "")            // make room
	copy((*nodes)[idx+1:], (*nodes)[idx:]) // shift
	(*nodes)[idx] = id

	r.N++

	realm := GetRealm(r)
	if realm != nil {
		if node := realm.Resolve(id); node != nil {
			return SetParent(node, r.ID)
		}
	}

	return nil
}

func (r *Raw) RemoveSubnodeID(kind, id string) error {
	idx, ok := r.GetSubnodeIndexOf(kind, id)
	if !ok {
		// removing node not in there is no-op?
		return nil
	}
	return r.RemoveSubnodeIndex(kind, idx)
}

func (r *Raw) RemoveSubnodeIndex(kind string, idx int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	nodes := r.nodes(kind)
	length := len(*nodes)
	if idx < 0 || idx >= length {
		return fmt.Errorf("index out of bounds: %d", idx)
	}
	*nodes = append((*nodes)[:idx], (*nodes)[idx+1:]...)
	r.N++
	return nil
}

func (r *Raw) MoveEntity(kind string, idx, to int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	nodes := r.nodes(kind)
	length := len(*nodes)
	if idx < 0 || idx >= length || to < 0 || to >= length {
		return fmt.Errorf("index out of bounds: idx=%d, to=%d, len=%d", idx, to, length)
	}
	if idx == to {
		return nil
	}

	val := (*nodes)[idx]
	if idx < to {
		copy((*nodes)[idx:], (*nodes)[idx+1:to+1])
	} else {
		copy((*nodes)[to+1:], (*nodes)[to:idx])
	}
	(*nodes)[to] = val

	r.N++
	return nil
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
