package entity

import "errors"

func Root(v any) E {
	if e := ToEntity(v); e != nil {
		cur := e
		var root E
		for cur != nil {
			cur = Parent(cur)
			if cur != nil {
				root = cur
			}
		}
		return root
	}
	return nil
}

func Siblings(v any) (siblings []E) {
	if e := ToEntity(v); e != nil {
		kind := Kind(v)
		p := Parent(e)
		if p == nil {
			return
		}
		for _, child := range Entities(p, kind) {
			if child.GetID() != e.GetID() {
				siblings = append(siblings, child)
			}
		}
	}
	return
}

func Destroy(v any) error {
	if e := ToEntity(v); e != nil {
		p := Parent(v)
		if p != nil {
			if err := RemoveEntity(p, Kind(e), e.GetID()); err != nil {
				return err
			}
		}
		return GetStore(v).Destroy(e)
	}
	return errors.ErrUnsupported
}
