package node

import "errors"

type ParentEntity interface {
	StoreEntity
	GetParent() E
}

func Parent(v any) E {
	if e := ToEntity(v); e != nil {
		if ee, ok := e.(ParentEntity); ok {
			return ee.GetParent()
		}
	}
	return nil
}

type ParentIDEntity interface {
	E
	GetParentID() string
}

func ParentID(v any) string {
	if e := ToEntity(v); e != nil {
		if ee, ok := e.(ParentIDEntity); ok {
			return ee.GetParentID()
		}
	}
	return ""
}

type SetParentEntity interface {
	E
	SetParent(id string) error
}

func SetParent(v any, id string) error {
	if e := ToEntity(v); e != nil {
		if ee, ok := e.(SetParentEntity); ok {
			defer Send(e, "", id)
			return ee.SetParent(id)
		}
	}
	return errors.ErrUnsupported
}

func Parents(v any) (parents []E) {
	if e := ToEntity(v); e != nil {
		cur := e
		for cur != nil {
			cur = Parent(cur)
			if cur != nil {
				parents = append(parents, cur)
			}
		}
	}
	return
}
