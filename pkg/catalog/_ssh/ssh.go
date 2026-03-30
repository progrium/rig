package ssh

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"log"
	"net"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/terminfo"
	"github.com/gliderlabs/ssh"
	xssh "golang.org/x/crypto/ssh"
	"tractor.dev/hack/pkg/misc"
)

type Session = ssh.Session

type SessionHandler interface {
	HandleSSH(ssh.Session)
}

type Server struct {
	Addr    string
	Handler SessionHandler

	*ssh.Server `json:"-"`
}

func (s *Server) Assemble(h SessionHandler) {
	if s.Handler == nil {
		s.Handler = h
	}
}

func (s *Server) Activate(ctx context.Context) (err error) {
	if s.Addr == "" {
		s.Addr = misc.ListenAddr()
	}
	l, err := net.Listen("tcp4", s.Addr)
	if err != nil {
		return err
	}
	hostKey, err := generateHostKey()
	if err != nil {
		return err
	}
	go func() {
		s.Server = &ssh.Server{
			Handler: func(sess ssh.Session) {
				if s.Handler != nil {
					s.Handler.HandleSSH(sess)
				}
			},
		}
		s.Server.AddHostKey(hostKey)
		if err := s.Server.Serve(l); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Println(err)
		}
	}()
	return nil
}

func (s *Server) Deactivate(ctx context.Context) error {
	if s.Server != nil {
		return s.Server.Shutdown(ctx)
	}
	return nil
}

func generateHostKey() (xssh.Signer, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	return xssh.ParsePrivateKey(privateKeyPEM)
}

func NewScreen(s ssh.Session) (tcell.Screen, error) {
	pi, ch, ok := s.Pty()
	if !ok {
		return nil, errors.New("no pty requested")
	}
	ti, err := terminfo.LookupTerminfo(pi.Term)
	if err != nil {
		return nil, err
	}
	screen, err := tcell.NewTerminfoScreenFromTtyTerminfo(&screenTTY{
		Session: s,
		size: tcell.WindowSize{
			Width:  pi.Window.Width,
			Height: pi.Window.Height,
		},
		ch: ch,
	}, ti)
	if err != nil {
		return nil, err
	}
	log.Println("screen made:", pi)
	return screen, nil
}

type screenTTY struct {
	ssh.Session
	size     tcell.WindowSize
	ch       <-chan ssh.Window
	resizecb func()
	mu       sync.Mutex
}

func (t *screenTTY) Start() error {
	go func() {
		for win := range t.ch {
			t.mu.Lock()
			t.size = tcell.WindowSize{
				Width:  win.Width,
				Height: win.Height,
			}
			if t.resizecb != nil {
				t.resizecb()
			}
			t.mu.Unlock()
		}
	}()
	return nil
}

func (t *screenTTY) Stop() error {
	return nil
}

func (t *screenTTY) Drain() error {
	return nil
}

func (t *screenTTY) WindowSize() (tcell.WindowSize, error) {
	return t.size, nil
}

func (t *screenTTY) NotifyResize(cb func()) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.resizecb = cb
}
