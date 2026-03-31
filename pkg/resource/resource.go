package resource

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/go-test/deep"
	"github.com/progrium/rig/pkg/manifold"
	"github.com/progrium/rig/pkg/node"
	"github.com/progrium/rig/pkg/state"
)

const (
	Destroyed   state.Enum = 1
	Reconciling state.Enum = 100
	Destroying  state.Enum = 101
	Synced      state.Enum = 200
	OutOfSync   state.Enum = 300
)

var ErrNotExist = errors.New("resource does not exist")

type EventHandler interface {
	HandleEvent(name string, data any)
}

type Provider[T any] interface {
	// CurrentState() state.Enum
	Client() *T
	// Subscribe(EventHandler)
	// Unsubscribe(EventHandler)
}

type ID string

type Lister[R any] interface {
	List(context.Context) ([]Resource[R], error)
}

type Creater[C any, R any] interface {
	Create(context.Context, *C) (Resource[R], error)
}

type Reader[R any] interface {
	Read(context.Context, ID) (R, error)
}

type Deleter interface {
	Delete(context.Context, ID) error
}

type ReadDeleter[R any] interface {
	Reader[R]
	Deleter
}

type NewResource[C any] struct {
	Draft  *C
	Save   func()
	Cancel func()
}

type Resource[R any] struct {
	ID        ID
	Status    string
	res       Reader[*R]
	Latest    R
	LastRead  time.Time
	syncValue *R
}

func (r *Resource[R]) Delete(n *node.Raw) bool {
	if rd, ok := r.res.(Deleter); ok {
		if err := rd.Delete(context.TODO(), r.ID); err != nil {
			log.Println("delete:", err)
		}
	}
	return true // not sure what this means yet
}

func (r *Resource[R]) SetSync(v *R) {
	r.syncValue = v
}

func (r *Resource[R]) CheckSync() {
	diff := deep.Equal(r.Latest, *r.syncValue)
	if diff == nil {
		r.Status = "synced"
	} else {
		r.Status = "out-of-sync"
		for _, d := range diff {
			log.Println("sync diff:", d)
		}
	}
}

func New[R any](res Reader[*R], id string, v R) Resource[R] {
	return Resource[R]{
		ID:       ID(id),
		res:      res,
		Latest:   v,
		LastRead: time.Now(),
	}
}

func ReadFromList[R any](ctx context.Context, lister Lister[R], id ID) (*R, error) {
	lst, err := lister.List(ctx)
	if err != nil {
		return nil, err
	}
	for _, r := range lst {
		if r.ID == id {
			return &r.Latest, nil
		}
	}
	return nil, ErrNotExist
}

func NewNode[C, R, L any](name string, parent manifold.Node, oldView string, creator Creater[C, R]) manifold.Node {
	n := manifold.FromEntity(node.New(name, node.Attributes{
		"view": "fields",
	}))
	req := new(C)
	com := &NewResource[C]{
		Draft: req,
		Save: func() {
			res, err := creator.Create(context.TODO(), req)
			if err != nil {
				log.Println(err)
				return
			}

			r := &res
			v := r.Latest
			r.SetSync(&v)
			r.CheckSync()
			nn := node.New(
				node.Name(r.Latest),
				node.Attributes{
					"view":      "fields",
					"desc":      string(r.ID),
					"deletable": "",
				},
				r,
				&v,
			)
			node.SetStore(nn, node.GetStore(parent))

			lst, _ := node.ComponentNode[L](parent)
			if lst == nil {
				log.Panicf("no valid list on: %v", parent)
			}
			// add to list component node
			node.AppendEntity(lst, node.Object, nn.ID) // error
			// also add to object, setting its parent
			parent.Objects().Append(manifold.FromEntity(nn)) // todo: avoid FromEntity?

			parent.Objects().Remove(n)      // error
			parent.SetAttr("view", oldView) // error
		},
		Cancel: func() {
			node.Destroy(n)               // error
			parent.SetAttr("view", oldView) // error

		},
	}
	// todo: errors
	n.AddComponent(com)
	node.SetStore(n, node.GetStore(parent))
	parent.Objects().Append(n)
	parent.SetAttr("view", "objects")
	return n
}

// for use in components implementing Nodes()
func ListNodes[T any](com manifold.Node, lister Lister[T]) (nodes node.Nodes) {
	resources, err := lister.List(context.Background())
	if err != nil {
		log.Println(err)
		return node.Nodes{}
	}
	// iterate over existing resources
	for idx, newres := range resources {
		// find a child with a resource matching this id
		r := childResource[T](com, newres.ID)
		if r == nil {
			// if not found, make a new node as child to this component
			r = &resources[idx]
			v := r.Latest
			r.SetSync(&v)
			r.CheckSync()
			nn := manifold.FromEntity(node.New(
				node.Name(r.Latest),
				node.Attributes{
					"view":      "fields",
					"desc":      string(r.ID),
					"deletable": "",
				},
				r,
				&v,
			))
			node.SetStore(nn, node.GetStore(com)) // hmmm
			com.Objects().Append(nn)
			continue
		}
		// if found, update value and check sync
		r.LastRead = newres.LastRead
		r.Latest = newres.Latest
		r.CheckSync()
	}
	// iterate over children of this component
	for _, e := range com.Objects().Nodes() {
		// find resource matching latest list
		v := node.Get[*Resource[T]](e)
		found := false
		for _, r := range resources {
			if r.ID == v.ID {
				found = true
				break
			}
		}
		// if not found, remove the node
		if !found {
			log.Println("not in resources, removing")
			node.Destroy(e) //error
			continue
		}

		// otherwise append node to returned nodes
		nodes = append(nodes, e)
	}
	return
}

func childResource[T any](n manifold.Node, rid ID) *Resource[T] {
	for _, res := range node.GetAll[*Resource[T]](n, node.Include{Children: true}) {
		if res.ID == rid {
			return res
		}
	}
	return nil
}
