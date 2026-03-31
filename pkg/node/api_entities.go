package node

import "errors"

type EntitiesEntity interface {
	GetEntities(kind string) []E
	GetEntityIndexOf(kind, id string) (int, bool)
}

func Entities(v any, kind string) []E {
	if e := ToEntity(v); e != nil {
		if ee, ok := e.(EntitiesEntity); ok {
			return ee.GetEntities(kind)
		}
	}
	return nil
}

func EntityIndexOf(v any, kind, id string) (int, bool) {
	if e := ToEntity(v); e != nil {
		if ee, ok := e.(EntitiesEntity); ok {
			return ee.GetEntityIndexOf(kind, id)
		}
	}
	return 0, false
}

type EntityCountEntity interface {
	GetEntityCount(kind string) int
}

func EntityCount(v any, kind string) int {
	if e := ToEntity(v); e != nil {
		if ee, ok := e.(EntityCountEntity); ok {
			return ee.GetEntityCount(kind)
		}
		return len(Entities(v, kind))
	}
	return 0
}

type AppenderEntity interface {
	AppendEntity(kind, id string) error
}

func AppendEntity(v any, kind, id string) error {
	if e := ToEntity(v); e != nil {
		if ee, ok := e.(AppenderEntity); ok {
			defer Send(e, "", kind, id)
			return ee.AppendEntity(kind, id)
		}
	}
	return errors.ErrUnsupported
}

type InsertAtEntity interface {
	InsertEntityAt(kind string, idx int, id string) error
}

func InsertEntityAt(v any, kind string, idx int, id string) error {
	if e := ToEntity(v); e != nil {
		if ee, ok := e.(InsertAtEntity); ok {
			defer Send(e, "", kind, idx, id)
			return ee.InsertEntityAt(kind, idx, id)
		}
	}
	return errors.ErrUnsupported
}

type RemoverEntity interface {
	RemoveEntity(kind, id string) error
}

func RemoveEntity(v any, kind, id string) error {
	if e := ToEntity(v); e != nil {
		if ee, ok := e.(RemoverEntity); ok {
			defer Send(e, "", kind, id)
			return ee.RemoveEntity(kind, id)
		}
	}
	return errors.ErrUnsupported
}

type RemoveAtEntity interface {
	RemoveEntityAt(kind string, idx int) error
}

func RemoveEntityAt(v any, kind string, idx int) error {
	if e := ToEntity(v); e != nil {
		if ee, ok := e.(RemoveAtEntity); ok {
			defer Send(e, "", kind, idx)
			return ee.RemoveEntityAt(kind, idx)
		}
	}
	return errors.ErrUnsupported
}

type MoverEntity interface {
	MoveEntity(kind string, idx, to int) error
}

func MoveEntity(v any, kind string, idx, to int) error {
	if e := ToEntity(v); e != nil {
		if ee, ok := e.(MoverEntity); ok {
			defer Send(e, "", kind, idx, to)
			return ee.MoveEntity(kind, idx, to)
		}
	}
	return errors.ErrUnsupported
}
