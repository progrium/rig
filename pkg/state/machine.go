package state

import (
	"errors"
	"fmt"
)

// TODO: enum categories
// neutral (gray) 0-99
// transition (empty/spinner) 100-199
// good (green) 200-299
// warning (yellow) 300-399
// bad (red) <0

const (
	Error Enum = -1
	Reset Enum = 0
)

type Enum int

type Transitions map[Enum][]Enum

type Machine interface {
	CurrentState() Enum
	MachineStates() Transitions
}

type Transitioner interface {
	Transition(to, from Enum) error
}

type machine interface {
	Machine
	provider() *State
}

func To(v Machine, s Enum) error {
	m, ok := v.(machine)
	if !ok {
		return errors.New("not a state machine")
	}

	old := m.CurrentState()

	from, ok := m.MachineStates()[s]
	if !ok && s != Reset && s != Error {
		return fmt.Errorf("invalid state: %d", s)
	}
	var valid bool
	for _, f := range from {
		if f == old {
			valid = true
		}
	}
	if len(from) == 0 {
		valid = true
	}
	if s == old {
		valid = true
	}
	if !valid {
		return fmt.Errorf("invalid transition: %d => %d", old, s)
	}

	p := m.provider()
	p.S = s

	if t, ok := v.(Transitioner); ok {
		if err := t.Transition(s, old); err != nil {
			p.S = old
			return err
		}
	}

	return nil
}

type State struct {
	S Enum
}

func (s State) CurrentState() Enum {
	return s.S
}

func (s State) MachineStates() Transitions {
	return Transitions{}
}

func (s *State) provider() *State {
	return s
}
