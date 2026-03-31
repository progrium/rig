package runtime

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/progrium/rig/pkg/manifold"
	"github.com/progrium/rig/pkg/node"
	"github.com/progrium/rig/pkg/pubsub"
	"github.com/progrium/rig/pkg/signal"
	"github.com/progrium/rig/pkg/telepath"
	"golang.org/x/net/websocket"
	"tractor.dev/toolkit-go/duplex/codec"
	"tractor.dev/toolkit-go/duplex/mux"
	"tractor.dev/toolkit-go/duplex/rpc"
	"tractor.dev/toolkit-go/duplex/talk"
)

type Action struct {
	Selector string
	Type     string
	Value    any
}

type Inspector struct {
	srv *http.Server

	updates *pubsub.Topic[node.Raw]
	ticker  *pubsub.Topic[bool]

	manifold.Component
}

func (m *Inspector) Initialize() {
	m.updates = pubsub.New[node.Raw]()
	m.ticker = pubsub.New[bool]()
}

func (m *Inspector) Activate(ctx context.Context) (err error) {
	if m.srv != nil {
		return nil
	}
	log.Println("starting inspector...")

	m.srv = &http.Server{
		Addr: ":11010",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := http.NewServeMux()
			h.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				websocket.Handler(func(ws *websocket.Conn) {
					ws.PayloadType = websocket.BinaryFrame
					sess := mux.New(ws)
					peer := talk.NewPeer(sess, codec.CBORCodec{})
					peer.Handle("setValue", rpc.HandlerFunc(func(r rpc.Responder, c *rpc.Call) {
						var action Action
						c.Receive(&action)
						parts := strings.SplitN(action.Selector, "/", 2)
						n := m.Object().Store().Resolve(parts[0])
						if n == nil {
							r.Return(fmt.Errorf("not found: %s", parts[0]))
							return
						}
						if err := telepath.Select(node.Value(n.Entity()), parts[1]).Set(action.Value); err != nil {
							r.Return(err)
							return
						}
						node.Send(n, "UpdateValue", action.Selector, action.Value)
					}))
					// TODO: besides finishing these, maybe entity should be responsible for
					// using telepath to update values to keep the signaling in entity
					peer.Handle("appendValue", rpc.HandlerFunc(func(r rpc.Responder, c *rpc.Call) {
						var action Action
						c.Receive(&action)
						log.Println("TODO: append", action)
					}))
					peer.Handle("unsetValue", rpc.HandlerFunc(func(r rpc.Responder, c *rpc.Call) {
						var action Action
						c.Receive(&action)
						log.Println("TODO: unset", action)
					}))
					defer peer.Close()
					go m.handlePeer(peer)
					peer.Respond()
				}).ServeHTTP(w, r)
			})
			h.ServeHTTP(w, r)
		}),
	}
	go m.srv.ListenAndServe()
	go m.updates.Dispatch()
	go m.ticker.Dispatch()
	go func() {
		dur := time.Second / 30
		for {
			<-time.After(dur)
			m.ticker.Publish(true)
		}
	}()

	// todo: shouldn't there be a better way to watch the store?
	// todo: race? occasionally get:
	// panic: interface conversion: entity.Store is nil, not *node.Store
	store := m.Object().Store().(*node.MemStore)
	store.Watch(m)

	return nil
}

func (m *Inspector) Signaled(s signal.Signal[node.E]) {
	n := node.Snapshot(s.Receiver.(node.E))
	m.updates.Publish(n)
}

func (m *Inspector) handlePeer(peer *talk.Peer) {
	stateBuffer := make(map[string]any)
	var stateMu sync.Mutex

	// write initial state to state buffer
	stateMu.Lock()
	manifold.Walk(m.Object(), func(n manifold.Node) error {
		snap := node.Snapshot(n)
		_, err := cbor.Marshal(snap)
		if err != nil {
			log.Println("skip:", n.Name(), n.ID(), err)
			return manifold.SkipNode
		}
		stateBuffer[n.ID()] = snap
		return nil
	})
	stateMu.Unlock()

	// write updates to state buffer
	updates := make(chan node.Raw)
	m.updates.Subscribe(updates)
	go func() {
		for n := range updates {
			nn := m.Object().Store().Resolve(n.ID)
			if nn == nil {
				// deleted
				stateMu.Lock()
				stateBuffer[n.ID] = nil
				stateMu.Unlock()
				continue
			}
			_, err := cbor.Marshal(n)
			if err != nil {
				log.Println("skip:", n.Name, n.ID, err)
				continue
			}
			stateMu.Lock()
			stateBuffer[n.ID] = n
			stateMu.Unlock()
		}
	}()

	// subscribe to update flusher ticks
	// and send and clear stateBuffer if
	// there is anything in it
	ticks := make(chan bool)
	m.ticker.Subscribe(ticks)
	for range ticks {
		stateMu.Lock()
		if len(stateBuffer) == 0 {
			stateMu.Unlock()
			continue
		}
		state := stateBuffer
		stateBuffer = map[string]any{}
		stateMu.Unlock()
		if _, err := peer.Call(context.Background(), "update", state, nil); err != nil {
			m.ticker.Unsubscribe(ticks)
			if !strings.Contains(err.Error(), "broken pipe") && !strings.Contains(err.Error(), "use of closed") {
				log.Println(err)
			}
		}
	}
	m.updates.Unsubscribe(updates)
}

func (m *Inspector) Deactivate(ctx context.Context) error {
	if m.srv != nil {
		srv := m.srv
		m.srv = nil
		return srv.Shutdown(ctx)
	}
	m.updates.Close()
	m.ticker.Close()
	return nil
}
