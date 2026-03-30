package runtime

import (
	"context"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/progrium/rig/pkg/entity"
	"github.com/progrium/rig/pkg/inspector"
	"github.com/progrium/rig/pkg/manifold"
	"github.com/progrium/rig/pkg/misc"
	"github.com/progrium/rig/pkg/module"
	"github.com/progrium/rig/pkg/node"
	signals "github.com/progrium/rig/pkg/signal"
	"tractor.dev/toolkit-go/duplex/codec"
	"tractor.dev/toolkit-go/duplex/fn"
	"tractor.dev/toolkit-go/duplex/mux"
	"tractor.dev/toolkit-go/duplex/rpc"
	"tractor.dev/toolkit-go/duplex/talk"
)

type Awaker interface {
	Awake()
}

type VSCodeBridge struct {
	*talk.Peer

	changeFns sync.Map
}

func (b *VSCodeBridge) BufferChanged(name string, data []byte) {
	v, ok := b.changeFns.Load(name)
	if ok {
		v.(func(data []byte))(data)
	}
}

func (b *VSCodeBridge) BufferListen(name string, fn func(data []byte)) {
	b.changeFns.Store(name, fn)
}

func Run(mainFacets ...any) {
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)

	// desktop.Start(func() {

	// if os.Getenv("STATE_RESET") != "" {
	// 	log.Println("clearing manifold state")
	// 	os.Remove("state.manifold")
	// }

	os.MkdirAll("/data", 0755)

	main := node.NewID("@main", "Main", mainFacets...)
	mod, err := module.LoadFrom("main", module.NewJSONProvider("/data/state.manifold"), main)
	if err != nil {
		log.Fatal(err)
	}

	manifold.Walk(mod.Main(), func(n manifold.Node) error {
		if n.Kind() == node.Component {
			if a, ok := n.Value().(Awaker); ok {
				a.Awake()
			}
		}
		return nil
	})

	handleProcSignals(mod)

	bridge := &VSCodeBridge{}
	inspector := inspector.Server{}
	super := node.New(node.RootID, bridge, inspector)
	if err := mod.Main().Store().Store(super); err != nil {
		log.Fatal(err)
	}
	super.Children = []string{"@main"}
	if err := mod.Main().SetParent(manifold.FromEntity(super)); err != nil {
		log.Fatal(err)
	}
	wb := New(super)

	go func() {
		time.Sleep(1 * time.Second)
		if err := wb.Toggle(node.RootID, true); err != nil {
			log.Fatal(err)
		}
	}()

	runService(mod, wb, "/var/run/manifold.sock", bridge)

	// 	desktop.Stop()
	// 	os.Exit(0)
	// })
}

func handleProcSignals(m *module.M) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	main := m.Main()
	go func() {
		for sig := range sigs {
			if sig == syscall.SIGTERM || sig == syscall.SIGINT {
				log.Println("signaled to shutdown")
				if err := node.Deactivate(context.TODO(), main); err != nil {
					log.Println(err)
				}
				// log.Println("saving module")
				if err := m.Save(); err != nil {
					log.Fatal(err)
				}
				// log.Println("module saved")
				// desktop.Stop()
				os.Exit(0)
			}
		}

	}()
}

func runService(mod *module.M, wb *Workbench, socketPath string, bridge *VSCodeBridge) {
	if err := os.RemoveAll(socketPath); err != nil {
		log.Fatal(err)
	}
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatal("Listen error:", err)
	}
	defer listener.Close()

	var watcher signals.Watcher[entity.E]
	if s := entity.GetStore(mod.Main().Entity()); s != nil {
		if sw, ok := s.(signals.Watcher[entity.E]); ok {
			watcher = sw
		}
	}

	var connected atomic.Bool
	logs := misc.NewBufferedPipe()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("Accept error:", err)
		}

		log.Println("session connected:", conn.RemoteAddr())
		go func(sess mux.Session) {

			peer := talk.NewPeer(sess, codec.CBORCodec{})
			peer.Handle("/", fn.HandlerFrom(wb))
			peer.Handle("/bridge/", fn.HandlerFrom(bridge))
			peer.Handle("/ping", fn.HandlerFrom(func() bool {
				return true
			}))
			peer.Handle("/logfeed", rpc.HandlerFunc(func(r rpc.Responder, c *rpc.Call) {
				c.Receive(nil)
				ch, err := r.Continue(nil)
				if err != nil {
					log.Println(err)
					return
				}
				defer ch.Close()
				if _, err = io.Copy(logs, ch); err != nil && err != io.EOF {
					log.Println(err)
				}
			}))
			peer.Handle("/logs", rpc.HandlerFunc(func(r rpc.Responder, c *rpc.Call) {
				ch, err := r.Continue(nil)
				if err != nil {
					log.Println(err)
					return
				}
				defer ch.Close()
				if _, err = io.Copy(ch, logs); err != nil && err != io.EOF {
					log.Println(err)
				}
			}))
			peer.Handle("/session", rpc.HandlerFunc(func(r rpc.Responder, c *rpc.Call) {
				ch, err := r.Continue(nil)
				if err != nil {
					log.Println(err)
					return
				}
				defer ch.Close()

				// log.Println("session started")
				connected.Store(true)

				// latest connection takes over bridge
				bridge.Peer = peer

				signaler := signals.Func[entity.E](func(s node.Signal) {
					e := s.Receiver.(entity.E)
					// log.Println("signal:", s.Name, e.GetName(), s.Args)
					if _, err := peer.Call(context.TODO(), "signaled", fn.Args{e.GetID()}, nil); err != nil {
						if err != io.EOF && !strings.Contains(err.Error(), "use of closed network connection") {
							log.Println(err)
						}
					}
				})
				watcher.Watch(signaler)

				for {
					// little keep alive
					<-time.After(1 * time.Second)
					if err := r.Send(nil); err != nil {
						if err != io.EOF {
							log.Println(err)
						}
						break
					}
				}

				watcher.Unwatch(signaler)

				connected.Store(false)
				go func() {
					<-time.After(1 * time.Second)
					if !connected.Load() {
						wb.Shutdown()
						os.Exit(0)
					}
				}()
			}))

			peer.Respond()

		}(mux.New(conn))
	}
}
