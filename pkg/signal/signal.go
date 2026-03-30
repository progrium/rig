package signal

import (
	"context"
	"log"
	"reflect"
	"runtime"
	"strings"
	"sync"
)

type Signal[T any] struct {
	Receiver Receiver[T]
	Name     string
	Args     []any
}

type Receiver[T any] interface {
	Signaled(s Signal[T])
}

type Func[T any] func(s Signal[T])

func (fn Func[T]) Signaled(s Signal[T]) {
	fn(s)
}

type Watcher[T any] interface {
	Watch(n Receiver[T])
	Unwatch(n Receiver[T])
}

type Dispatcher[T any] struct {
	m sync.Map
}

func (w *Dispatcher[T]) Watch(n Receiver[T]) {
	w.m.Store(reflect.ValueOf(n), n)
}

func (w *Dispatcher[T]) Unwatch(n Receiver[T]) {
	w.m.Delete(reflect.ValueOf(n))
}

func (w *Dispatcher[T]) Signaled(s Signal[T]) {
	w.m.Range(func(k, v any) bool {
		if n, ok := v.(Receiver[T]); ok {
			n.Signaled(s)
		}
		return true
	})
}

func Send[T any](rcvr Receiver[T], signal string, args ...any) {
	if signal == "" {
		signal = fnCaller(1)
	}
	rcvr.Signaled(Signal[T]{
		Receiver: rcvr,
		Name:     signal,
		Args:     args,
	})
}

func Receive[T any](watcher Watcher[T], ctx context.Context, ch chan Signal[T]) {
	var fn Func[T]
	fn = Func[T](func(s Signal[T]) {
		if ctx != nil {
			// first check if the context is done
			select {
			case <-ctx.Done():
				watcher.Unwatch(fn)
				close(ch)
				return
			default:
			}
		}

		// now try to send
		select {
		case ch <- s:
		default:
			// if unable to send,
			// treat this watch as done
			watcher.Unwatch(fn)
			// logging for now to tune channel buffer
			log.Println("signals channel unable to keep up")
		}
	})
	watcher.Watch(fn)
}

func fnCaller(n int) string {
	pc, _, _, _ := runtime.Caller(n + 1)
	details := runtime.FuncForPC(pc)
	fqn := strings.Split(details.Name(), ".")
	return fqn[len(fqn)-1]
}
