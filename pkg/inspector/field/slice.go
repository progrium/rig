package field

import (
	"reflect"
)

var TypeSlice = "Slice"

type Slice struct {
	FieldInfo
	v          reflect.Value
	elemFields []Field
}

func SliceFrom(v interface{}, fi FieldInfo) Slice {
	f := Slice{v: reflect.ValueOf(v), FieldInfo: fi}
	if f.v.Type().Elem().Kind() == reflect.Struct {
		for _, fieldname := range fields(f.v.Type().Elem()) {
			sf, ok := fromField(StructFrom(reflect.Zero(f.v.Type().Elem()).Interface(), FieldInfo{}), fieldname)
			if ok {
				f.elemFields = append(f.elemFields, sf)
			}
		}
	}
	return f
}

func (f Slice) TypeName() string       { return TypeSlice }
func (f Slice) Type() reflect.Type     { return f.v.Type() }
func (f Slice) Value() reflect.Value   { return f.v }
func (f Slice) IndexType() Type        { return Integer{} }
func (f Slice) ElementType() Type      { return FromType(f.v.Type().Elem()) }
func (f Slice) ElementFields() []Field { return f.elemFields }
func (f Slice) Fields() (fields []Field) {
	for i := 0; i < f.v.Len(); i++ {
		fields = append(fields, FromValue(f.v.Index(i).Interface(), FieldInfo{}))
	}
	return
}
