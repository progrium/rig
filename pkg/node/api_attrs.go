package node

import (
	"errors"
)

type AttrEntity interface {
	E
	GetAttrs() []string
	GetAttr(key string) string
}

func Attrs(v any) []string {
	if e := ToEntity(v); e != nil {
		if ee, ok := e.(AttrEntity); ok {
			return ee.GetAttrs()
		}
	}
	return nil
}

func Attr(v any, key string) string {
	if e := ToEntity(v); e != nil {
		if ee, ok := e.(AttrEntity); ok {
			return ee.GetAttr(key)
		}
	}
	return ""
}

func AttrMap(v any) map[string]string {
	if e := ToEntity(v); e != nil {
		if ee, ok := e.(AttrEntity); ok {
			attrs := make(map[string]string)
			for _, attr := range ee.GetAttrs() {
				attrs[attr] = ee.GetAttr(attr)
			}
			return attrs
		}
	}
	return nil
}

type SetAttrEntity interface {
	E
	SetAttr(key, value string) error
}

func SetAttr(v any, key, value string) error {
	if e := ToEntity(v); e != nil {
		if Attr(v, key) == value {
			return nil
		}
		if ee, ok := e.(SetAttrEntity); ok {
			defer Send(e, "", key, value)
			return ee.SetAttr(key, value)
		}
	}
	return errors.ErrUnsupported
}

type DelAttrEntity interface {
	E
	DelAttr(key string) error
}

func DelAttr(v any, key string) error {
	if e := ToEntity(v); e != nil {
		if !HasAttr(v, key) {
			return nil
		}
		if ee, ok := e.(DelAttrEntity); ok {
			defer Send(e, "", key)
			return ee.DelAttr(key)
		}
	}
	return errors.ErrUnsupported
}

func HasAttr(v any, key string) bool {
	for _, attr := range Attrs(v) {
		if attr == key {
			return true
		}
	}
	return false
}
