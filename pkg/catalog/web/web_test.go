package web

import (
	"context"
	"os"
	"testing"

	"github.com/progrium/rig/pkg/node"
	"github.com/progrium/rig/pkg/util"
)

func TestWebActivation(t *testing.T) {
	d := os.DirFS("/tmp")
	fs := &FileServer{}
	s := &Server{}
	nl := &TCPListener{ListenAddr: ":0"}
	n := node.New("", d, fs, s, nl)
	oa := util.ObjectActivator{}
	if err := oa.ActivateObject(node.Context(context.Background(), n)); err != nil {
		t.Fatal(err)
	}
	if fs.FS == nil {
		t.Fatal("file server fs not set")
	}
	if s.Listener != nl {
		t.Fatal("server listener not net listener")
	}
}
