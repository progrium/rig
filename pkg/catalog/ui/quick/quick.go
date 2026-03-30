package quick

import "github.com/progrium/rig/pkg/manifold"

type Handler interface {
	HandleUI(event string, data map[string]string)
}

type Element interface {
	Element() // just a tag interface for now
}

type Layout struct {
	Handler Handler
	manifold.Component
}

func (l *Layout) Assemble(h Handler) {
	if l.Handler == nil {
		l.Handler = h
	}
}

func (l *Layout) RelatedComponents() []string {
	return []string{"."}
}

type Column struct {
}

func (e *Column) Element() {}

type Text struct {
	Text string
}

func (e *Text) Element() {}

type Input struct {
	Key    string
	Events bool
}

func (e *Input) Element() {}

type Button struct {
	Label string
	Key   string
}

func (e *Button) Element() {}

type Checkbox struct {
	Label  string
	Key    string
	Events string
}

func (e *Checkbox) Element() {}

type Radio struct {
	Label  string
	Key    string
	Events string
}

func (e *Radio) Element() {}

type Separator struct {
}

func (e *Separator) Element() {}
