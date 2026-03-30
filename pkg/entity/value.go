package entity

import "errors"

type ValueEntity interface {
	E
	GetValue() any
}

func Value(v any) any {
	if e := ToEntity(v); e != nil {
		if ee, ok := e.(ValueEntity); ok {
			return ee.GetValue()
		}
	}
	return nil
}

type SetValueEntity interface {
	E
	SetValue(v any) error
}

func SetValue(v, val any) error {
	if e := ToEntity(v); e != nil {
		if Value(e) == val {
			return nil
		}
		if ee, ok := e.(SetValueEntity); ok {
			defer Send(e, "", val)
			return ee.SetValue(val)
		}
	}
	return errors.ErrUnsupported
}
