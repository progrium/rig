package node

import (
	"errors"
	"runtime"
	"strings"

	"github.com/progrium/rig/pkg/signal"
)

// Signal is a named event targeted at a Node receiver with optional arguments.
type Signal = signal.Signal[Node]

// SignalNode is implemented by nodes that handle signals delivered via Send.
type SignalNode interface {
	Node
	Signaled(s Signal)
}

// Signaled invokes n.Signaled(s) when n implements SignalNode; otherwise it is a no-op.
func Signaled(n Node, s Signal) {
	if sn, ok := Unwrap[SignalNode](n); ok {
		sn.Signaled(s)
	}
}

// Send delivers a signal to n when n implements SignalNode.
// If sig is empty, the name of the direct caller is used as the signal name.
// It returns errors.ErrUnsupported when n does not implement SignalNode.
func Send(n Node, sig string, args ...any) error {
	if sn, ok := Unwrap[SignalNode](n); ok {
		if sig == "" {
			sig = fnCaller(1)
		}
		signal.Send(sn, sig, args...)
		return nil
	}
	return errors.ErrUnsupported
}

// Watch registers fn as a signal receiver on n when n implements signal.Watcher[Node].
// It returns errors.ErrUnsupported otherwise.
func Watch(n Node, fn signal.Func[Node]) error {
	if sn, ok := Unwrap[signal.Watcher[Node]](n); ok {
		sn.Watch(fn)
		return nil
	}
	return errors.ErrUnsupported
}

// Unwatch removes a previously registered fn from n when n implements signal.Watcher[Node].
// It returns errors.ErrUnsupported otherwise.
func Unwatch(n Node, fn signal.Func[Node]) error {
	if w, ok := Unwrap[signal.Watcher[Node]](n); ok {
		w.Unwatch(fn)
		return nil
	}
	return errors.ErrUnsupported
}

func fnCaller(n int) string {
	pc, _, _, _ := runtime.Caller(n + 1)
	details := runtime.FuncForPC(pc)
	fqn := strings.Split(details.Name(), ".")
	return fqn[len(fqn)-1]
}

// Signaled handles incoming signals for Raw nodes.
// When s targets this node, it forwards to the embedded Dispatcher and, if the node's Realm
// implements signal.Receiver[Node], to the realm as well.
// It then asynchronously notifies component values (for object nodes with components) and the
// immediate parent when they implement signal.Receiver[Node].
func (r *Raw) Signaled(s Signal) {
	if s.Receiver.(Node).NodeID() == r.ID {
		r.Dispatcher.Signaled(s)
		realm := GetRealm(r)
		if realm != nil {
			if sn, ok := realm.(signal.Receiver[Node]); ok {
				sn.Signaled(s)
			}
		}
	}
	go func() {
		if Kind(r) == TypeObject && SubnodeCount(r, TypeComponent) > 0 {
			for _, com := range Subnodes(r, TypeComponent) {
				if sr, ok := Value(com).(signal.Receiver[Node]); ok {
					sr.Signaled(s)
				}
			}
		}

		if sr, ok := Parent(r).(signal.Receiver[Node]); ok {
			sr.Signaled(s)
		}
	}()
}
