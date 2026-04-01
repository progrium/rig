package node

import "errors"

// TODO: upgrade this to a runtime field on Raw?

// Error returns an error built from n's "error" attribute when it is non-empty.
// It returns nil when the attribute is missing or empty.
func Error(n Node) error {
	errStr := Attr(n, "error")
	if errStr == "" {
		return nil
	}
	return errors.New(errStr)
}

// SetError stores err on n as the "error" attribute, or removes that attribute when err is nil.
func SetError(n Node, err error) error {
	if err == nil {
		return DelAttr(n, "error")
	}
	return SetAttr(n, "error", err.Error())
}
