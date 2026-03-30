package file

import (
	"context"
	"log"
	"path/filepath"
	"sync"

	"github.com/progrium/rig/pkg/manifold"
	"github.com/progrium/rig/pkg/node"
	"github.com/progrium/rig/pkg/runtime"
	"tractor.dev/toolkit-go/duplex/fn"
)

// maybe this can just be an interface?
type Buffer struct {
	Data []byte

	mu sync.Mutex
	manifold.Component
}

func (b *Buffer) WriteFile(buf []byte) error {
	b.mu.Lock()
	b.Data = buf
	b.mu.Unlock()
	return nil
}

func (b *Buffer) ReadFile() ([]byte, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.Data, nil
}

func (b *Buffer) OnSelected() {
	path := filepath.Join(b.Object().ID(), b.Object().Name())
	bridge := node.Get[*runtime.VSCodeBridge](b.Object(), node.Include{Parents: true})
	if bridge != nil {
		bridge.BufferListen(path, func(data []byte) {
			b.WriteFile(data)
		})
		data, _ := b.ReadFile()
		if _, err := bridge.Call(context.TODO(), "execCommand", fn.Args{"manifold.editBuffer", fn.Args{path, data}}, nil); err != nil {
			log.Println(err)
		}
	}
}
