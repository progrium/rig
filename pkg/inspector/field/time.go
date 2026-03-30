package field

import (
	"reflect"
	"time"
)

var TypeTime = "Time"

type Time struct {
	FieldInfo
	v time.Time

	Layout string
}

func TimeFrom(v time.Time, fi FieldInfo) Time {
	return Time{v: v, Layout: time.RFC3339, FieldInfo: fi}
}

func (f Time) TypeName() string     { return TypeTime }
func (f Time) Type() reflect.Type   { return reflect.TypeOf(f.v) }
func (f Time) Value() reflect.Value { return reflect.ValueOf(f.v) }
func (f Time) Default() string      { return "" }
func (f Time) Enum() []string       { return nil }
func (f Time) Range() *Range        { return nil }

func (f Time) Parse(s string) (v Value, err error) {
	ff := f
	ff.v, err = time.Parse(time.RFC3339, s)
	v = ff
	return
}

func (f Time) Format() string {
	return f.v.Format(f.Layout)
}
