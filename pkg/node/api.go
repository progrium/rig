package node

import (
	"errors"
	"fmt"
)

type E interface {
	Node
	GetID() string
	GetName() string
}

type Node interface {
	Entity() E
}

type Nodes []Node

func ToEntity(v any) E {
	if ee, ok := v.(Node); ok {
		return ee.Entity()
	}
	if e, ok := v.(E); ok {
		return e
	}
	return nil
}

func Name(v any) string {
	if e := ToEntity(v); e != nil {
		return e.GetName()
	}
	return fmt.Sprintf("?%#v", v)
}

type KindEntity interface {
	GetKind() string
}

func Kind(v any) string {
	if e := ToEntity(v); e != nil {
		if ee, ok := e.(KindEntity); ok {
			return ee.GetKind()
		}
	}
	return ""
}

type SetNameEntity interface {
	E
	SetName(name string) error
}

func SetName(v any, name string) error {
	if e := ToEntity(v); e != nil {
		if Name(e) == name {
			return nil
		}
		if ee, ok := e.(SetNameEntity); ok {
			defer Send(e, "", name)
			return ee.SetName(name)
		}
	}
	return errors.ErrUnsupported
}
