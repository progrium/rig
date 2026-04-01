package node

import (
	"github.com/progrium/rig/pkg/meta"
)

// Attributes is a facet for New/NewID: key/value pairs merged into the new node's attributes.
type Attributes map[string]string

// Children is a facet for New/NewID: raw child nodes attached under the new node.
type Children []*Raw

// Initializer is implemented by component values that run one-time setup when the component
// is created via New/NewID (non-factory facets).
type Initializer interface {
	Initialize()
}

// New builds a Raw node with an auto-generated id, the given name, and optional facets.
// See NewID for supported facet types and post-construction behavior.
func New(name string, facets ...any) *Raw {
	n := NewID("", name, facets...)
	return n
}

// NewID builds a Raw node with the given id and name, then applies facets in order:
//   - Attributes: merged into the node's Attrs
//   - Children or *Raw: attached as ordered children; their embedded map is folded into this
//     node, Embedded cleared on the child, realm cleared, root set to this node
//   - meta.Factory: builds a component via factory.New(), stores _factory on the component
//   - any other value: wrapped with NewComponent; if it implements Initializer, Initialize runs;
//     ComponentAttacher.ComponentAttached is invoked asynchronously; Deactivator forces
//     attribute "activated" to "false"
//
// After facets, EnableComponent is attempted for each direct component subnode that has no
// prior error attribute; the first enable error is written on this node as attribute "error".
func NewID(id, name string, facets ...any) *Raw {
	n := NewRaw(name, nil, id)
	n.Embedded = make(map[string]*Raw)
	addChild := func(child *Raw) {
		child.Parent = n.ID
		n.mu.Lock()

		// copy child embedded nodes into
		// our embedded nodes
		if len(child.Embedded) > 0 {
			for _, embed := range child.Embedded {
				n.Embedded[embed.ID] = embed
			}
		}
		// set child embedded nodes and store
		// to nil and set its root to us
		// before putting it in our embedded
		child.Embedded = nil
		child.realm = nil
		child.root = n
		n.Embedded[child.ID] = child

		n.Children = append(n.Children, child.ID)

		n.mu.Unlock()
	}
	for _, f := range facets {
		switch facet := f.(type) {
		case Attributes:
			for k, v := range facet {
				n.Attrs[k] = v
			}
		case Children:
			for _, child := range facet {
				addChild(child)
			}
		case *Raw:
			addChild(facet)
		default:
			var com *Raw
			factory, ok := facet.(meta.Factory)
			if ok {
				value, name := factory.New()
				com = NewComponent(value)
				com.Attrs["_factory"] = name
			} else {
				com = NewComponent(facet)
				if i, ok := com.Value.(Initializer); ok {
					i.Initialize()
				}
			}
			com.Parent = n.ID
			com.root = n
			n.Components = append(n.Components, com.ID)
			n.mu.Lock()
			n.Embedded[com.ID] = com
			n.mu.Unlock()
			// parent wont be set here...
			if ca, ok := com.Value.(ComponentAttacher); ok {
				go ca.ComponentAttached(com)
			}
			if _, ok := com.Value.(Deactivator); ok {
				n.Attrs["activated"] = "false"
			}
		}
	}
	// todo: wait until this node attached to root so enable can include parents?
	for _, com := range Subnodes(n, TypeComponent) {
		if Error(com) != nil {
			continue
		}
		if err := EnableComponent(com); err != nil {
			SetAttr(n, "error", err.Error())
			break
		}
	}
	return n
}
