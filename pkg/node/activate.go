package node

import (
	"context"
	"fmt"

	"github.com/progrium/rig/pkg/entity"
	"tractor.dev/toolkit-go/engine"
)

type Activator interface {
	Activate(ctx context.Context) error
}

type Deactivator interface {
	Activator
	Deactivate(ctx context.Context) error
}

type ActivationStrategy interface {
	ActivateObject(ctx context.Context) error
}

type DeactivationStrategy interface {
	DeactivateObject(ctx context.Context) error
}

func Activate(ctx context.Context, n entity.Node) error {
	if err := entity.SetAttr(n, "busy", "true"); err != nil {
		return err
	}
	defer func() {
		if err := entity.DelAttr(n, "busy"); err != nil {
			panic(err)
		}
	}()

	if strat := Get[ActivationStrategy](n); strat != nil {
		if err := strat.ActivateObject(Context(ctx, n)); err != nil {
			if e := entity.SetAttr(n, "error", err.Error()); e != nil {
				panic(err)
			}
			return err
		}
		return nil
	}

	stateful := false
	asm, _ := engine.New()
	for _, com := range entity.Entities(n, Component) {
		v := entity.Value(com)
		if err := asm.Add(v); err != nil {
			panic(err)
		}
	}
	for _, com := range entity.Entities(n, Component) {
		v := entity.Value(com)
		if err := asm.Assemble(v); err != nil {
			panic(err)
		}
		activator, ok := v.(Activator)
		if ok && entity.ComponentEnabled(com) {
			if err := activator.Activate(Context(ctx, com)); err != nil {
				if e := entity.SetAttr(n, "error", fmt.Sprintf("%s: %s", entity.Name(com), err.Error())); e != nil {
					panic(err)
				}
				return err
			}
		}
		if _, ok = entity.Value(com).(Deactivator); ok {
			stateful = true
		}
	}
	if stateful {
		if err := entity.SetAttr(n, "activated", "true"); err != nil {
			return err
		}
	}
	return nil
}

func Deactivate(ctx context.Context, n entity.Node) error {
	if err := entity.SetAttr(n, "busy", "true"); err != nil {
		return err
	}
	defer func() {
		if err := entity.DelAttr(n, "busy"); err != nil {
			panic(err)
		}
	}()

	if strat := Get[DeactivationStrategy](n); strat != nil {
		if err := strat.DeactivateObject(Context(ctx, n)); err != nil {
			if e := entity.SetAttr(n, "error", err.Error()); e != nil {
				panic(err)
			}
			return err
		}
		return entity.SetAttr(n, "activated", "false")
	}

	for _, com := range entity.Entities(n, Component) {
		activator, ok := entity.Value(com).(Deactivator)
		if ok {
			if err := activator.Deactivate(Context(ctx, com)); err != nil {
				if e := entity.SetAttr(n, "error", fmt.Sprintf("%s: %s", entity.Name(com), err.Error())); e != nil {
					panic(err)
				}
				return err
			}
		}
	}
	return entity.SetAttr(n, "activated", "false")
}
