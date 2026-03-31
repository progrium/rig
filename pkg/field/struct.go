package field

import (
	"fmt"
	"reflect"
	"strings"
)

var TypeStruct = "Struct"

type Struct struct {
	FieldInfo
	v      interface{}
	fields []Field
}

func StructFrom(v interface{}, fi FieldInfo) Struct {
	f := Struct{v: v, FieldInfo: fi}
	for _, fieldname := range fields(f.Type()) {
		sf, ok := fromField(f, fieldname)
		if ok {
			f.fields = append(f.fields, sf)
		}
	}
	return f
}

func (f Struct) TypeName() string     { return TypeStruct }
func (f Struct) Type() reflect.Type   { return f.Value().Type() }
func (f Struct) Value() reflect.Value { return reflect.Indirect(reflect.ValueOf(f.v)) }
func (f Struct) Fields() []Field      { return f.fields }

func fields(t reflect.Type) []string {
	var f []string
	for i := 0; i < t.NumField(); i++ {
		name := t.Field(i).Name
		if name[0] == strings.ToUpper(name)[0] {
			f = append(f, name)
		}
	}
	return f
}

func fromField(f Struct, fname string) (Field, bool) {
	id := ""
	if f.ID() != "" {
		id = fmt.Sprintf("%s/%s", f.ID(), fname)
	}
	fi := WithFieldInfo(fname, id)
	sf, ok := f.Type().FieldByName(fname)
	if !ok {
		return nil, false
	}
	if sf.Tag.Get("hidden") == "true" {
		return nil, false
	}
	rv := f.Value().FieldByName(fname)
	if rv.Kind() == reflect.Func {
		return nil, false
	}
	if rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface {
		return FromValue(rv, fi), true
	}
	return FromValue(rv.Interface(), fi), true
}
