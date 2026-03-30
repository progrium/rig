package web

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os/exec"
	goruntime "runtime"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/progrium/rig/pkg/manifold"
	"github.com/progrium/rig/pkg/misc"
	"github.com/progrium/rig/pkg/node"
	"github.com/progrium/rig/pkg/runtime"
	netws "golang.org/x/net/websocket"
	"tractor.dev/toolkit-go/duplex/fn"
)

type TCPListener struct {
	ListenAddr   string
	net.Listener `json:"-"`
}

func (l *TCPListener) Initialize() {
	l.ListenAddr = ":0"
}

func (l *TCPListener) URL() string {
	if l.Listener == nil {
		return ""
	}
	host := strings.ReplaceAll(l.Listener.Addr().String(), "0.0.0.0", "localhost")
	return fmt.Sprintf("http://%s", host)
}

func (l *TCPListener) Activate(ctx context.Context) (err error) {
	addr := l.ListenAddr
	if addr == "" {
		addr = misc.ListenAddr()
	}
	l.Listener, err = net.Listen("tcp4", addr)
	if err != nil {
		log.Println("listen:", err)
		return
	}
	return nil
}

func (l *TCPListener) Deactivate(ctx context.Context) error {
	if l.Listener != nil {
		err := l.Listener.Close()
		if err != nil && strings.Contains(err.Error(), "use of closed network connection") {
			return nil
		}
		return err
	}
	return nil
}

func (l *TCPListener) Provides() Listener {
	return l
}

type Server struct {
	LaunchBrowser bool
	LaunchPath    string

	Listener Listener
	Handler  http.Handler

	*http.Server `json:"-"`
	manifold.Component
}

func (s *Server) Assemble(l Listener, h http.Handler) {
	s.Handler = h
	s.Listener = l
}

func (s *Server) Activate(ctx context.Context) error {
	go func() {
		s.Server = &http.Server{
			Handler: s.Handler,
		}
		if err := s.Server.Serve(s.Listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			if !strings.Contains(err.Error(), "broken pipe") && !strings.Contains(err.Error(), "use of closed") {
				log.Println("serve:", err)
			}
		}
	}()
	if s.LaunchBrowser {
		url := misc.URLJoin(s.Listener.URL(), s.LaunchPath)
		bridge := node.Get[*runtime.VSCodeBridge](s.Object(), node.Include{Parents: true})
		if bridge != nil {
			if _, err := bridge.Call(context.TODO(), "execCommand", fn.Args{"manifold.openBrowser", fn.Args{url}}, nil); err != nil {
				return err
			}
		} else {
			openBrowser(url)
		}
	}
	return nil
}

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch goruntime.GOOS {
	case "darwin":
		cmd = "open"
	case "windows":
		cmd = "rundll32"
		args = append(args, "url.dll,FileProtocolHandler")
	case "linux":
		cmd = "xdg-open"
	default:
		return fmt.Errorf("unsupported platform")
	}
	args = append(args, url)

	return exec.Command(cmd, args...).Start()
}

func (s *Server) Deactivate(ctx context.Context) error {
	if s.Server != nil {
		if err := s.Server.Shutdown(ctx); err != nil {
			return err
		}
		s.Server = nil
	}
	return nil
}

type FileServer struct {
	FS  fs.FS
	Dir string

	http.Handler `json:"-" cbor:"-"`
}

func (s *FileServer) Assemble(fsys fs.FS) {
	s.FS = fsys
	s.Handler = http.FileServerFS(misc.Must(fs.Sub(s.FS, s.Dir)))
}

func (s *FileServer) Provides() http.Handler {
	return s.Handler
}

type Router struct {
	manifold.Component
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	mux := http.NewServeMux()
	for _, route := range node.GetAll[*Route](r.Object(), node.Include{Children: true}) {
		path := route.Path
		if path == "" {
			path = "/" + route.Object().Name()
		}
		mux.Handle(path, route.Handler)
	}
	mux.ServeHTTP(w, req)
}

type Route struct {
	Path    string
	Handler http.Handler

	manifold.Component
}

func (r *Route) Assemble(h http.Handler) {
	r.Handler = h
}

type Matcher interface {
	MatchHTTP(req *http.Request) bool
}

type MatchRouter struct {
	manifold.Component
}

func (m *MatchRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, handler := range node.GetAll[http.Handler](m.Object(), node.Include{Children: true}) {
		if handler == m {
			continue
		}
		matcher, ok := handler.(Matcher)
		if ok {
			if matcher.MatchHTTP(r) {
				handler.ServeHTTP(w, r)
				return
			} else {
				continue
			}
		}
		handler.ServeHTTP(w, r)
		return
	}
}

type SocketHandler interface {
	HandleWS(conn *websocket.Conn)
}

type SocketHandlerAlt interface {
	HandleWS(conn *netws.Conn)
}

type SocketUpgrader struct {
	Handler    SocketHandler
	HandlerAlt SocketHandlerAlt
	// Fallback http.Handler

	upgrader websocket.Upgrader
}

func (u *SocketUpgrader) Initialize() {
	u.upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
}

func (u *SocketUpgrader) Assemble(h SocketHandler, alt SocketHandlerAlt) {
	u.Handler = h
	u.HandlerAlt = alt
	// u.Fallback = f
}

func (u *SocketUpgrader) MatchHTTP(req *http.Request) bool {
	return websocket.IsWebSocketUpgrade(req)
}

func (u *SocketUpgrader) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if websocket.IsWebSocketUpgrade(r) {
		if u.Handler != nil {
			conn, err := u.upgrader.Upgrade(w, r, nil)
			if err != nil {
				log.Println(err)
				return
			}
			u.Handler.HandleWS(conn)
		}
		if u.HandlerAlt != nil {
			netws.Handler(func(c *netws.Conn) {
				u.HandlerAlt.HandleWS(c)
			}).ServeHTTP(w, r)
		}
	}
	// if u.Fallback != nil {
	// 	u.Fallback.ServeHTTP(w, r)
	// }
}
