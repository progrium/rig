package entity

import "errors"

// type ErrorEntity interface {
// 	E
// 	GetError() error
// }

func Error(v any) error {
	errStr := Attr(v, "error")
	if errStr == "" {
		return nil
	}
	return errors.New(errStr)
}

// type SetErrorEntity interface {
// 	E
// 	SetError(err error) error
// }

// func SetError(v any, err error) error {
// 	if e := ToEntity(v); e != nil {
// 		if Error(e) == err {
// 			return nil
// 		}
// 		if ee, ok := e.(SetErrorEntity); ok {
// 			defer Send(e, "", err)
// 			return ee.SetError(err)
// 		}
// 	}
// 	return errors.ErrUnsupported
// }
