package field

import (
	"errors"
	"fmt"
	"reflect"
)

var TypePointer = "Pointer"

type Pointer struct {
	FieldInfo
	v reflect.Value
}

func PointerFrom(v reflect.Value, fi FieldInfo) Pointer {
	return Pointer{v: v, FieldInfo: fi}
}

func (f Pointer) TypeName() string          { return TypePointer }
func (f Pointer) Type() reflect.Type        { return reflect.TypeOf(f) }
func (f Pointer) Value() reflect.Value      { return reflect.ValueOf(fmt.Sprintf("%#v", f.v.Interface())) }
func (f Pointer) Default() string           { return "" }
func (f Pointer) Enum() []string            { return nil }
func (f Pointer) PointerType() reflect.Type { return f.v.Type().Elem() }

func (f Pointer) Range() *Range {
	return nil
}

func (f Pointer) Parse(s string) (v Value, err error) {
	err = errors.New("unable to parse Pointer field values")
	return
}

func (f Pointer) Format() string {
	return fmt.Sprint(f.Value().Interface())
}
