package field

import (
	"reflect"
)

var TypeMap = "Map"

type Map struct {
	FieldInfo
	v reflect.Value
}

func MapFrom(v interface{}, fi FieldInfo) Map {
	return Map{v: reflect.ValueOf(v), FieldInfo: fi}
}

func (f Map) TypeName() string       { return TypeMap }
func (f Map) Type() reflect.Type     { return f.v.Type() }
func (f Map) Value() reflect.Value   { return f.v }
func (f Map) IndexType() Type        { return FromType(f.v.Type().Key()) }
func (f Map) ElementType() Type      { return FromType(f.v.Type().Elem()) }
func (f Map) ElementFields() []Field { return nil }
