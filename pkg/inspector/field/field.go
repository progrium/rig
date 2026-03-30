package field

import (
	"fmt"
	"reflect"
)

type Range struct {
	Min, Step int
	Max       uint
}

type Type interface {
	TypeName() string
	Type() reflect.Type
}

type BasicType interface {
	Type
	Parse(s string) (Value, error)
	Format() string
	Enum() []string
	Range() *Range
	Default() string
}

type StructType interface {
	Type
	Fields() []Field
}

// maps (order keys)
// arrays/slices
type CollectionType interface {
	Type
	IndexType() Type
	ElementType() Type
	ElementFields() []Field
}

type PointerType interface {
	PointerType() reflect.Type
}

type Field interface {
	Type
	Name() string
	ID() string
	Flags() Flag
}

type Value interface {
	Field
	Value() reflect.Value
}

type FieldInfo struct {
	name string
	id   string
	mask Flag
}

func WithFieldInfo(name string, id string, flags ...Flag) FieldInfo {
	var mask Flag
	for _, f := range flags {
		mask = mask | f
	}
	return FieldInfo{
		name: name,
		id:   id,
		mask: mask,
	}
}

type Flag uint

const (
	FlagRequired Flag = 1 << iota
	FlagHidden
	FlagReadonly
	FlagSecret
	FlagDisabled
)

var flags = map[string]func(f Field) bool{
	"required": Required,
	"hidden":   Hidden,
	"readonly": Readonly,
	"secret":   Secret,
	"disabled": Disabled,
}

func (f FieldInfo) ID() string   { return f.id }
func (f FieldInfo) Name() string { return f.name }
func (f FieldInfo) Flags() Flag  { return f.mask }

func default_(v, d string) string {
	if v != "" {
		return v
	}
	return d
}

func Required(f Field) bool {
	return (f.Flags() & FlagRequired) != 0
}

func Hidden(f Field) bool {
	return (f.Flags() & FlagHidden) != 0
}

func Readonly(f Field) bool {
	return (f.Flags() & FlagReadonly) != 0
}

func Secret(f Field) bool {
	return (f.Flags() & FlagSecret) != 0
}

func Disabled(f Field) bool {
	return (f.Flags() & FlagDisabled) != 0
}

type Data struct {
	TypeName   string
	IdxType    *string `json:",omitempty"`
	ElemType   *string `json:",omitempty"`
	ElemFields []Data
	Default    string
	Enum       []string
	Range      *Range `json:",omitempty"`
	Name       string
	Flags      []string
	ID         *string      `json:",omitempty"`
	Value      *interface{} `json:",omitempty"`
	Fields     []Data
	Annots     map[string]any
}

func ToData(f Field) (d Data) {
	return ToDataWithFunc(f, nil)
}

func ToDataWithFunc(f Field, fn func(d *Data, f Field)) (d Data) {
	d.Annots = make(map[string]any)
	d.TypeName = f.TypeName()
	if bf, ok := f.(BasicType); ok {
		d.Default = bf.Default()
		d.Enum = bf.Enum()
		d.Range = bf.Range()
	}
	if sf, ok := f.(StructType); ok {
		d.Fields = make([]Data, 0)
		for _, ff := range sf.Fields() {
			d.Fields = append(d.Fields, ToDataWithFunc(ff, fn))
		}
	}
	if cf, ok := f.(CollectionType); ok {
		idx := cf.IndexType().TypeName()
		d.IdxType = &idx
		el := cf.ElementType().TypeName()
		d.ElemType = &el
		d.ElemFields = make([]Data, 0)
		for _, ef := range cf.ElementFields() {
			d.ElemFields = append(d.ElemFields, ToDataWithFunc(ef, fn))
		}
	}
	if pt, ok := f.(PointerType); ok {
		el := pt.PointerType().Name()
		d.ElemType = &el
		if d.TypeName == TypeInterface {
			d.Annots["pkgpath"] = pt.PointerType().PkgPath()
		}
	}
	d.Name = f.Name()
	if f.ID() != "" {
		s := f.ID()
		d.ID = &s
	}
	for name, check := range flags {
		if check(f) {
			d.Flags = append(d.Flags, name)
		}
	}
	if fv, ok := f.(Value); ok {
		v := fv.Value().Interface()
		d.Value = &v
	}
	if fn != nil {
		fn(&d, f)
	}
	return
}

func FromType(t reflect.Type) Type {
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return Integer{AllowNegative: true}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return Integer{}
	case reflect.Float32, reflect.Float64:
		return Float{}
	case reflect.Bool:
		return Boolean{}
	case reflect.String:
		return String{}
	case reflect.Struct:
		return Struct{}
	case reflect.Slice:
		return Slice{}
	case reflect.Array:
		return Array{}
	case reflect.Map:
		return Map{}
	case reflect.Interface:
		return Interface{}
	default:
		panic(fmt.Sprintf("FromType: type kind not supported: %s (%v)", t.Kind(), t))
	}
}

func FromValue(v interface{}, fi FieldInfo) Field {
	// hack to support pointers .. sometimes
	if rv, ok := v.(reflect.Value); ok {
		switch rv.Kind() {
		case reflect.Pointer:
			return PointerFrom(rv, fi)
		case reflect.Interface:
			return InterfaceFrom(rv, fi)
		default:
			panic("unexpected use of reflect.Value for ValueFrom")
		}
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Invalid {
		panic("TODO: handle invalid?")
	}
	rv = reflect.Indirect(rv)
	switch rv.Type().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return IntegerFrom(int(rv.Int()), fi)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return IntegerFromUint(uint(rv.Uint()), fi)
	case reflect.Float32, reflect.Float64:
		return FloatFrom(rv.Float(), fi)
	case reflect.Bool:
		return BooleanFrom(rv.Bool(), fi)
	case reflect.String:
		return StringFrom(rv.Interface(), fi)
	case reflect.Struct:
		return StructFrom(rv.Interface(), fi)
	case reflect.Map:
		return MapFrom(rv.Interface(), fi)
	case reflect.Slice:
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			// i guess just assumes its bytes
			return BlobFrom(rv.Bytes(), fi)
		}
		return SliceFrom(rv.Interface(), fi)
	case reflect.Array:
		return ArrayFrom(rv.Interface(), fi)
	default:
		panic(fmt.Sprintf("FromValue: type kind not supported: %s (%#v)", rv.Type().Kind(), v))
	}
}

type Enumer interface {
	Enum() []string
}

func Enum(v interface{}) []string {
	if ev, ok := v.(Enumer); ok {
		return ev.Enum()
	}
	return nil
}
