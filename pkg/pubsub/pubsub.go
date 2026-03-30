package pubsub

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

const DefaultSize = 1024

type Topic[T any] struct {
	mu      sync.Mutex
	subs    []chan T
	queue   chan T
	closed  bool
	running atomic.Bool
	size    int
}

func New[T any]() *Topic[T] {
	return NewWithSize[T](DefaultSize)
}

func NewWithSize[T any](size int) *Topic[T] {
	return &Topic[T]{
		subs:  make([]chan T, 0),
		queue: make(chan T, size),
		size:  size,
	}
}

func (t *Topic[T]) IsRunning() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return !t.closed && t.running.Load()
}

func (t *Topic[T]) Dispatch() {
	if t.running.Load() {
		return
	}
	if t.closed {
		// starting up again
		t.closed = false
		t.subs = make([]chan T, 0)
		t.queue = make(chan T, t.size)
	}
	t.running.Store(true)
	for msg := range t.queue {
		t.mu.Lock()
		subs := t.subs
		t.mu.Unlock()
		for _, sub := range subs {
			select {
			case sub <- msg:
			default:
			}
		}
	}
	t.running.Store(false)
}

func (t *Topic[T]) Publish(msg T) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		log.Println("warning: publishing to topic not dispatching")
		panic("publishing to closed topic")
	}

	t.queue <- msg
}

func (t *Topic[T]) Subscribe(ch chan T) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return
	}

	if !t.running.Load() {
		log.Println("warning: subscribing to topic not dispatching")
	}

	t.subs = append(t.subs, ch)
}

func (t *Topic[T]) Unsubscribe(ch chan T) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return
	}

	t.subs = remove(t.subs, ch)

	go func() {
		defer recover()
		<-time.After(1 * time.Second)
		close(ch)
	}()
}

func remove[T any](subs []chan T, ch chan T) []chan T {
	for i, v := range subs {
		if v == ch {
			return append(subs[:i], subs[i+1:]...)
		}
	}
	return subs
}

func (t *Topic[T]) Close() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return
	}

	t.running.Store(false)
	t.closed = true
	close(t.queue)

	for _, sub := range t.subs {
		close(sub)
	}
}

// todo for below: separate wrapper? automatic via reflection?

func (t *Topic[T]) Activate(ctx context.Context) (err error) {
	go t.Dispatch()
	return nil
}

func (t *Topic[T]) Deactivate(ctx context.Context) error {
	t.Close()
	return nil
}
