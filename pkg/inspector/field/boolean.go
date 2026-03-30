package field

import (
	"fmt"
	"reflect"
	"strings"
)

var TypeBoolean = "Boolean"

type Boolean struct {
	FieldInfo
	v bool
}

func BooleanFrom(v bool, fi FieldInfo) Boolean {
	return Boolean{v: v, FieldInfo: fi}
}

func (f Boolean) TypeName() string     { return TypeBoolean }
func (f Boolean) Type() reflect.Type   { return reflect.TypeOf(f.v) }
func (f Boolean) Value() reflect.Value { return reflect.ValueOf(f.v) }
func (f Boolean) Default() string      { return "false" }
func (f Boolean) Enum() []string       { return nil }
func (f Boolean) Range() *Range        { return nil }

func (f Boolean) Parse(s string) (Value, error) {
	v := f
	switch strings.ToLower(s) {
	case "true", "t", "y", "1":
		v.v = true
	case "false", "f", "n", "0", "":
		v.v = false
	default:
		return nil, fmt.Errorf("Parse: invalid Boolean value: %v", s)
	}
	return &v, nil
}

func (f Boolean) Format() string {
	switch f.v {
	case true:
		return "true"
	case false:
		return "false"
	default:
		return ""
	}
}
