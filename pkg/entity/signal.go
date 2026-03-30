package entity

import (
	"errors"
	"runtime"
	"strings"

	"github.com/progrium/rig/pkg/signal"
)

type Signal = signal.Signal[E]

type SignalEntity interface {
	Signaled(s Signal)
}

func Signaled(v any, s Signal) {
	if e := ToEntity(v); e != nil {
		if se, ok := e.(SignalEntity); ok {
			se.Signaled(s)
		}
	}
}

func Send(v any, sig string, args ...any) error {
	if e := ToEntity(v); e != nil {
		if sig == "" {
			sig = fnCaller(1)
		}
		if se, ok := e.(SignalEntity); ok {
			signal.Send(se, sig, args...)
			return nil
		}
	}
	return errors.ErrUnsupported
}

func Watch(v any, fn signal.Func[E]) error {
	if e := ToEntity(v); e != nil {
		if w, ok := e.(signal.Watcher[E]); ok {
			w.Watch(fn)
		}
	}
	return errors.ErrUnsupported
}

func Unwatch(v any, fn signal.Func[E]) error {
	if e := ToEntity(v); e != nil {
		if w, ok := e.(signal.Watcher[E]); ok {
			w.Unwatch(fn)
		}
	}
	return errors.ErrUnsupported
}

func fnCaller(n int) string {
	pc, _, _, _ := runtime.Caller(n + 1)
	details := runtime.FuncForPC(pc)
	fqn := strings.Split(details.Name(), ".")
	return fqn[len(fqn)-1]
}
