package node

import "errors"

// ValueNode is implemented by nodes that expose a value.
type ValueNode interface {
	Node
	GetValue() any
}

// Value returns n's value when n implements ValueNode.
// It returns nil when no value is available.
func Value(n Node) any {
	if vn, ok := Unwrap[ValueNode](n); ok {
		return vn.GetValue()
	}
	return nil
}

// SetValueNode is implemented by nodes whose value can be updated.
type SetValueNode interface {
	Node
	SetValue(v any) error
}

// SetValue updates n's value when supported.
// It is a no-op if the value is unchanged, emits a signal on success path,
// and returns errors.ErrUnsupported when n does not implement SetValueNode.
func SetValue(n Node, val any) error {
	if Value(n) == val {
		return nil
	}
	if snn, ok := Unwrap[SetValueNode](n); ok {
		defer Send(n, "", val)
		return snn.SetValue(val)
	}
	return errors.ErrUnsupported
}
