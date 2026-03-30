package field

import (
	"net/url"
	"reflect"
)

var TypeURL = "URL"

type URL struct {
	FieldInfo
	v *url.URL
}

func URLFrom(v *url.URL, fi FieldInfo) URL {
	return URL{v: v, FieldInfo: fi}
}

func (f URL) TypeName() string     { return TypeURL }
func (f URL) Type() reflect.Type   { return reflect.TypeOf(f.v).Elem() }
func (f URL) Value() reflect.Value { return reflect.ValueOf(f.v).Elem() }
func (f URL) Default() string      { return "" }
func (f URL) Enum() []string       { return nil }
func (f URL) Range() *Range        { return nil }

func (f URL) Parse(s string) (v Value, err error) {
	ff := f
	ff.v, err = url.Parse(s)
	v = f
	return
}

func (f URL) Format() string {
	return f.v.String()
}
