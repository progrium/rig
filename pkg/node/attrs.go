package node

import (
	"errors"
)

// AttrNode is implemented by nodes that expose string attributes.
type AttrNode interface {
	Node
	GetAttrs() []string
	GetAttr(key string) string
}

// Attrs returns the list of attribute keys for n.
// It returns nil when n does not implement AttrNode.
func Attrs(n Node) []string {
	if an, ok := Unwrap[AttrNode](n); ok {
		return an.GetAttrs()
	}
	return nil
}

// Attr returns the attribute value for key on n.
// It returns an empty string when n does not implement AttrNode.
func Attr(n Node, key string) string {
	if an, ok := Unwrap[AttrNode](n); ok {
		return an.GetAttr(key)
	}
	return ""
}

// HasAttr reports whether n has an attribute named key.
func HasAttr(n Node, key string) bool {
	for _, attr := range Attrs(n) {
		if attr == key {
			return true
		}
	}
	return false
}

// AttrMap returns all attributes on n as a key/value map.
// It returns an empty map when n does not implement AttrNode.
func AttrMap(n Node) map[string]string {
	if an, ok := Unwrap[AttrNode](n); ok {
		attrs := make(map[string]string)
		for _, attr := range an.GetAttrs() {
			attrs[attr] = an.GetAttr(attr)
		}
		return attrs
	}
	return map[string]string{}
}

// SetAttrNode is implemented by nodes that can mutate attributes.
type SetAttrNode interface {
	Node
	SetAttr(key, value string) error
}

// SetAttr updates key with value on n when supported.
// It is a no-op if the value is already set, emits a signal on success path,
// and returns errors.ErrUnsupported when n does not implement SetAttrNode.
func SetAttr(n Node, key, value string) error {
	if Attr(n, key) == value {
		return nil
	}
	if snn, ok := Unwrap[SetAttrNode](n); ok {
		defer Send(n, "", key, value)
		return snn.SetAttr(key, value)
	}
	return errors.ErrUnsupported
}

// DelAttrNode is implemented by nodes that can delete attributes.
type DelAttrNode interface {
	Node
	DelAttr(key string) error
}

// DelAttr removes key from n when supported.
// It is a no-op when key is missing, emits a signal on success path,
// and returns errors.ErrUnsupported when n does not implement DelAttrNode.
func DelAttr(n Node, key string) error {
	if !HasAttr(n, key) {
		return nil
	}
	if dnn, ok := Unwrap[DelAttrNode](n); ok {
		defer Send(n, "", key)
		return dnn.DelAttr(key)
	}
	return errors.ErrUnsupported
}
