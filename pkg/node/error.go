package node

import "errors"

func Error(v any) error {
	errStr := Attr(v, "error")
	if errStr == "" {
		return nil
	}
	return errors.New(errStr)
}
