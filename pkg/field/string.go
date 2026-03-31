package field

import "reflect"

var TypeString = "String"

type String struct {
	FieldInfo
	v interface{}

	DefaultString string
}

func StringFrom(v interface{}, fi FieldInfo) String {
	// TODO: check string kind?
	return String{v: v, FieldInfo: fi}
}

func (f String) TypeName() string     { return TypeString }
func (f String) Type() reflect.Type   { return reflect.TypeOf(f.v) }
func (f String) Value() reflect.Value { return reflect.ValueOf(f.v) }
func (f String) Default() string      { return default_(f.DefaultString, "") }
func (f String) Enum() []string       { return Enum(f.v) }
func (f String) Range() *Range        { return nil }

func (f String) Parse(s string) (Value, error) {
	ff := f
	ff.v = s
	return ff, nil
}

func (f String) Format() string {
	return f.v.(string)
}
