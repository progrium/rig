package runtime

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/progrium/rig/pkg/field"
	"github.com/progrium/rig/pkg/manifold"
	"github.com/progrium/rig/pkg/meta"
	"github.com/progrium/rig/pkg/node"
	"github.com/progrium/rig/pkg/pubsub"
	"github.com/progrium/rig/pkg/signal"
	"github.com/progrium/rig/pkg/telepath"
	"github.com/progrium/rig/web"
	"golang.org/x/net/websocket"
	"tractor.dev/toolkit-go/duplex/codec"
	"tractor.dev/toolkit-go/duplex/fn"
	"tractor.dev/toolkit-go/duplex/mux"
	"tractor.dev/toolkit-go/duplex/rpc"
	"tractor.dev/toolkit-go/duplex/talk"
	"tractor.dev/toolkit-go/engine/fs/watchfs"
)

type Action struct {
	Selector string
	Type     string
	Value    any
	Args     []any
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
					peer.Handle("watchFile", rpc.HandlerFunc(func(r rpc.Responder, c *rpc.Call) {
						var path string
						c.Receive(&path)
						w, err := watchfs.New(os.DirFS("/").(fs.StatFS)).Watch(strings.TrimPrefix(path, "/"), &watchfs.Config{
							Recursive: true,
						})
						if err != nil {
							r.Return("watch error: " + err.Error())
							return
						}
						defer w.Close()
						ch, err := r.Continue(nil)
						if err != nil {
							log.Println(err)
							return
						}
						defer ch.Close()

						for event := range w.Iter() {
							if event.Type == watchfs.EventCreate || event.Type == watchfs.EventWrite {
								// log.Println("event:", event.Path)
								if err := r.Send(event.Path); err != nil {
									log.Println("send error:", err)
									return
								}
							}
						}
					}))
					peer.Handle("fields", rpc.HandlerFunc(func(r rpc.Responder, c *rpc.Call) {
						var id string
						c.Receive(&id)

						n := m.Object().Realm().Resolve(id)
						if n == nil {
							r.Return(fmt.Errorf("not found: %s", id))
							return
						}

						ch, err := r.Continue()
						if err != nil {
							log.Println(err)
							return
						}
						defer ch.Close()

						if node.IsComponent(n) {
							if node.Value(n) == nil {
								return
							}
							com := field.FromValue(node.Value(n), field.WithFieldInfo(node.Name(n), n.NodeID()))
							for _, f := range field.ToData(com).Fields {
								if err := r.Send(f); err != nil {
									log.Println(err)
									return
								}
							}
						} else {
							mn := manifold.FromNode(n)
							for _, n := range mn.Components().Nodes() {
								if n.Value() == nil {
									continue
								}
								f := field.FromValue(n.Value(), field.WithFieldInfo(n.Name(), n.ID()))
								if err := r.Send(field.ToData(f)); err != nil {
									log.Println(err)
									return
								}
							}
						}
					}))

					peer.Handle("GetAddComponents", rpc.HandlerFunc(func(r rpc.Responder, c *rpc.Call) {
						var args []string
						c.Receive(&args)

						var others []string
						var main []string
						for pkgpath := range meta.Components {
							if strings.HasPrefix(pkgpath, "main.") {
								main = append(main, pkgpath)
								continue
							}
							others = append(others, pkgpath)
						}
						var items []string
						items = append(items, "--Main")
						items = append(items, main...)
						items = append(items, "--Library")
						items = append(items, others...)
						r.Return(items)
					}))
					peer.Handle("AddComponent", rpc.HandlerFunc(func(r rpc.Responder, c *rpc.Call) {
						var args []string
						c.Receive(&args)
						id := args[0]
						typ := args[1]

						n := manifold.FromNode(m.Object().Realm().Resolve(id))
						if n == nil {
							r.Return(fmt.Errorf("not found: %s", id))
							return
						}

						t, ok := meta.Components[typ]
						if !ok {
							r.Return(fmt.Errorf("%s type not found", typ))
							return
						}
						v := reflect.New(t).Interface()
						if i, ok := v.(node.Initializer); ok {
							i.Initialize()
						}

						com, err := n.AddComponent(v)
						if err != nil {
							r.Return(err)
							return
						}
						r.Return(com.SetAttr("enabled", "true"))
					}))

					peer.Handle("listEditors", rpc.HandlerFunc(func(r rpc.Responder, c *rpc.Call) {
						c.Receive(nil)
						entries, err := fs.ReadDir(web.FS, "editors")
						if err != nil {
							r.Return(err)
							return
						}
						var editors []string
						for _, entry := range entries {
							if entry.IsDir() {
								editors = append(editors, entry.Name())
							}
						}
						r.Return(editors)
					}))
					peer.Handle("listCatalog", rpc.HandlerFunc(func(r rpc.Responder, c *rpc.Call) {
						c.Receive(nil)
						var symbols []string
						for k := range meta.Components {
							symbols = append(symbols, k)
						}
						r.Return(symbols)
					}))
					peer.Handle("setValue", rpc.HandlerFunc(func(r rpc.Responder, c *rpc.Call) {
						var action Action
						c.Receive(&action)
						log.Println("setValue", action)
						parts := strings.SplitN(action.Selector, "/", 2)
						n := m.Object().Realm().Resolve(parts[0])
						if n == nil {
							r.Return(fmt.Errorf("not found: %s", parts[0]))
							return
						}
						if err := telepath.Select(node.Value(n), parts[1]).Set(action.Value); err != nil {
							r.Return(err)
							return
						}
						node.Send(n, "UpdateValue", action.Selector, action.Value)
					}))
					peer.Handle("addComponent", rpc.HandlerFunc(func(r rpc.Responder, c *rpc.Call) {
						var action Action
						c.Receive(&action)
						log.Println("addComponent", action)
						n := m.Object().Realm().Resolve(action.Selector)
						if n == nil {
							r.Return(fmt.Errorf("not found: %s", action.Selector))
							return
						}

						if action.Value == nil || action.Value == "" {
							r.Return(fmt.Errorf("type is required"))
							return
						}

						typ := action.Value.(string)
						t, ok := meta.Components[typ]
						if !ok {
							r.Return(fmt.Errorf("%s type not found", typ))
							return
						}
						v := reflect.New(t).Interface()
						if i, ok := v.(node.Initializer); ok {
							i.Initialize()
						}

						mn := manifold.FromNode(n)
						com, err := mn.AddComponent(v)
						if err != nil {
							r.Return(err)
							return
						}
						r.Return(com.ID())
					}))
					peer.Handle("addObject", rpc.HandlerFunc(func(r rpc.Responder, c *rpc.Call) {
						var action Action
						c.Receive(&action)
						log.Println("addObject", action)
						if action.Value == nil {
							r.Return(fmt.Errorf("name is required"))
							return
						}
						newNode := node.New(action.Value.(string))
						if err := m.Object().Realm().Store(newNode); err != nil {
							r.Return(err)
							return
						}
						if action.Selector != "" {
							n := m.Object().Realm().Resolve(action.Selector)
							if n == nil {
								r.Return(fmt.Errorf("not found: %s", action.Selector))
								return
							}
							mn := manifold.FromNode(n)
							if err := mn.Children().Append(manifold.FromNode(newNode)); err != nil {
								r.Return(err)
								return
							}
						}
						r.Return(newNode.NodeID())
					}))
					peer.Handle("callMethod", rpc.HandlerFunc(func(r rpc.Responder, c *rpc.Call) {
						var action Action
						c.Receive(&action)
						log.Println("callMethod", action)
						parts := strings.SplitN(action.Selector, "/", 2)
						n := m.Object().Realm().Resolve(parts[0])
						if n == nil {
							r.Return(fmt.Errorf("not found: %s", parts[0]))
							return
						}
						mn := manifold.FromNode(n)
						if err := telepath.Select(mn, parts[1]).Call(fn.Args(action.Args)); err != nil {
							r.Return(err)
							return
						}
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
						log.Println("unsetValue", action)
						parts := strings.SplitN(action.Selector, "/", 2)
						n := m.Object().Realm().Resolve(parts[0])
						if n == nil {
							r.Return(fmt.Errorf("not found: %s", parts[0]))
							return
						}
						if err := telepath.Select(node.Value(n), parts[1]).Delete(); err != nil {
							r.Return(err)
							return
						}
						node.Send(n, "UnsetValue", action.Selector)
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
	store := m.Object().Realm().(*node.BasicRealm)
	store.Watch(m)

	return nil
}

func (m *Inspector) Signaled(s signal.Signal[node.Node]) {
	n := node.Snapshot(s.Receiver.(node.Node))
	log.Println("<- signaled", n.ID, s.Name, s.Args)
	m.updates.Publish(n)
	// log.Println("-> published", n.ID, s.Name, s.Args)
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
			// log.Println("update:", n.ID)
			nn := m.Object().Realm().Resolve(n.ID)
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
			// var ids []string
			// for id := range stateBuffer {
			// 	ids = append(ids, id)
			// }
			// log.Println("stateBuffer:", ids)
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
		// var ids []string
		// for id := range state {
		// 	ids = append(ids, id)
		// }
		// log.Println("flush:", ids)
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
