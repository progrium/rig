package state_test

import (
	"fmt"
	"testing"

	"github.com/progrium/rig/pkg/state"
)

type Demo struct {
	state.State
}

const (
	Loading state.Enum = 100
	Ready   state.Enum = 200
	Warning state.Enum = 300
)

func (c Demo) States() state.Transitions {
	return state.Transitions{
		Loading: []state.Enum{},
		Ready:   []state.Enum{Warning, Loading},
	}
}

func (c Demo) Transition(to, from state.Enum) error {
	fmt.Println(from, "=>", to)
	return nil
}

func TestDemo(t *testing.T) {
	d := &Demo{}
	if d.CurrentState() != state.Reset {
		t.Fatal("bad state:", d.CurrentState())
	}
	err := state.To(d, Loading)
	if err != nil {
		t.Fatal("to:", err)
	}
	if d.CurrentState() != Loading {
		t.Fatal("bad state:", d.CurrentState())
	}
}
