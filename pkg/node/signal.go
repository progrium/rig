package node

import (
	"github.com/progrium/rig/pkg/entity"
	"github.com/progrium/rig/pkg/signal"
)

type Signal = signal.Signal[entity.E]

func (r *Raw) Signaled(s Signal) {
	if s.Receiver.(entity.E).GetID() == r.ID {
		// signal local watchers
		r.Dispatcher.Signaled(s)
		// signal store if it wants
		store := entity.GetStore(r)
		if store != nil {
			if sn, ok := store.(signal.Receiver[entity.E]); ok {
				sn.Signaled(s)
			}
		}
	}
	// maybe instead of a goroutine per signal per node, we have a single
	// goroutine and we just send to it via channel?
	go func() {
		// send to components if any
		if entity.Kind(r) == Object && entity.EntityCount(r, Component) > 0 {
			for _, com := range entity.Entities(r, Component) {
				if sr, ok := entity.Value(com).(signal.Receiver[entity.E]); ok {
					sr.Signaled(s)
				}
			}
		}

		// signal parent to propagate up ancestors if it can
		if sr, ok := entity.Parent(r).(signal.Receiver[entity.E]); ok {
			sr.Signaled(s)
		}
	}()
}

func Send(n entity.Node, sig string, args []any) {
	rn, ok := n.Entity().(*Raw)
	if !ok {
		panic("unable to send signal")
	}
	signal.Send(rn, sig, args)
}
