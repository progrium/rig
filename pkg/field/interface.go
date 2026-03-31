package field

import (
	"errors"
	"fmt"
	"reflect"
)

var TypeInterface = "Interface"

type Interface struct {
	FieldInfo
	v reflect.Value
}

func InterfaceFrom(v reflect.Value, fi FieldInfo) Interface {
	return Interface{v: v, FieldInfo: fi}
}

func (f Interface) TypeName() string          { return TypeInterface }
func (f Interface) Type() reflect.Type        { return reflect.TypeOf(f) }
func (f Interface) Value() reflect.Value      { return reflect.ValueOf(fmt.Sprintf("%#v", f.v.Interface())) }
func (f Interface) Default() string           { return "" }
func (f Interface) Enum() []string            { return nil }
func (f Interface) PointerType() reflect.Type { return f.v.Type() }

func (f Interface) Range() *Range {
	return nil
}

func (f Interface) Parse(s string) (v Value, err error) {
	err = errors.New("unable to parse Interface field values")
	return
}

func (f Interface) Format() string {
	return fmt.Sprint(f.Value().Interface())
}
