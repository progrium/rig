package field

import (
	"reflect"
	"strconv"
)

var TypeFloat = "Float"

type Float struct {
	FieldInfo
	v float64

	DefaultFloat string
}

func FloatFrom(v float64, fi FieldInfo) Float {
	return Float{v: v, FieldInfo: fi}
}

func (f Float) TypeName() string     { return TypeFloat }
func (f Float) Type() reflect.Type   { return reflect.TypeOf(f.v) }
func (f Float) Value() reflect.Value { return reflect.ValueOf(f.v) }
func (f Float) Default() string      { return default_(f.DefaultFloat, "0.0") }
func (f Float) Enum() []string       { return nil }
func (f Float) Range() *Range        { return nil }

func (f Float) Parse(s string) (v Value, err error) {
	ff := f
	ff.v, err = strconv.ParseFloat(s, 64)
	v = f
	return
}

func (f Float) Format() string {
	return strconv.FormatFloat(f.v, 'f', -1, 64)
}
