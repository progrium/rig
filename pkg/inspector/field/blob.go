package field

import "reflect"

var TypeBlob = "Blob"

type Blob struct {
	FieldInfo
	v []byte
}

func BlobFrom(v []byte, fi FieldInfo) Blob {
	return Blob{v: v, FieldInfo: fi}
}

func (f Blob) TypeName() string     { return TypeBlob }
func (f Blob) Type() reflect.Type   { return reflect.TypeOf(f.v) }
func (f Blob) Value() reflect.Value { return reflect.ValueOf(f.v) }
func (f Blob) Default() string      { return "" }
func (f Blob) Enum() []string       { return nil }
func (f Blob) Range() *Range        { return nil }

func (f Blob) Parse(s string) (Value, error) {
	v := f
	v.v = []byte(s)
	return &v, nil
}

func (f Blob) Format() string {
	return string(f.v)
}
