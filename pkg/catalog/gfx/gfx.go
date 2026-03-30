package gfx

import (
	"context"
	"time"

	"github.com/progrium/rig/pkg/manifold"
	"github.com/progrium/rig/pkg/node"
)

type Vector3 struct {
	X, Y, Z float64
}

type Vector2 struct {
	X, Y float64
}

type UpdateLoop struct {
	Framerate int

	ticker *time.Ticker
	done   chan bool
	manifold.Component
}

type Updater interface {
	Update(deltaTime time.Duration)
}

type UpdateFlusher interface {
	FlushUpdates()
}

func (l *UpdateLoop) Initialize() {
	if l.Framerate == 0 {
		l.Framerate = 60
	}
	l.done = make(chan bool)
}

func (l *UpdateLoop) Activate(ctx context.Context) error {
	l.ticker = time.NewTicker(time.Second / time.Duration(l.Framerate))
	go func() {
		lastUpdate := time.Now()
		for {
			select {
			case t := <-l.ticker.C:
				// Calculate time delta
				deltaTime := t.Sub(lastUpdate)
				lastUpdate = t

				updaters := node.GetAll[Updater](l.Object(), node.Include{Descendants: true})
				for _, updater := range updaters {
					updater.Update(deltaTime)
				}

				flushers := node.GetAll[UpdateFlusher](l.Object(), node.Include{Descendants: true, Parents: true})
				for _, flusher := range flushers {
					flusher.FlushUpdates()
				}
			case <-l.done:
				return
			}
		}
	}()
	return nil
}

func (l *UpdateLoop) Deactivate(ctx context.Context) error {
	if l.ticker != nil {
		l.ticker.Stop()
	}
	l.done <- true
	return nil
}
