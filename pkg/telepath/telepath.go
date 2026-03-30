package telepath

import (
	"log"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/progrium/jsonschema"
)

// Telepath is embedded in Cursors so you can pass
// Cursors as values in remote Set/Insert calls
// indicating to do a local Select on the other
// end, allowing you to set by reference/pointers
// as long as addressable with Telepath
type Telepath struct {
	Telepath string // unique identifying field
}

func FromTelepath(root Root, v any) (Cursor, bool) {
	switch vv := v.(type) {
	case Telepath:
		return NewCursor(root, vv.Telepath), true
	case map[any]any:
		if path, ok := vv["Telepath"].(string); ok {
			return NewCursor(root, path), true
		}
		return nil, false
	case map[string]any:
		if path, ok := vv["Telepath"].(string); ok {
			return NewCursor(root, path), true
		}
		return nil, false
	default:
		return nil, false
	}
}

type Metadata struct {
	Type   Type
	Schema *jsonschema.Schema
	Len    int
}

type Type struct {
	Name    string
	Kind    string
	Elem    *Type
	Fields  []string
	Methods []string
}

type Virtual interface {
	PathValue(name string) reflect.Value
	PathNames() []string
}

type Deleter interface {
	PathDelete() error
}

func WalkTo(v any, path string) reflect.Value {
	parts := strings.Split(strings.TrimLeft(path, "/"), "/")
	curr := reflect.ValueOf(v)
	for _, name := range parts {
		if name == "" {
			continue
		}
		curr = Value(curr, name)
		if !curr.IsValid() {
			break
		}
	}
	return curr
}

// Value extracts a sub-value from the provided reflect.Value, based on the given name string.
// The function supports various types: if the input value v implements a Virtual interface,
// it calls the PathValue method. For slices and arrays, it tries to
// convert the name to an index and returns the element at that index. For pointers, it
// recursively calls itself with the de-referenced pointer. For maps, it returns the value
// mapped by the name key. For structs, it looks up the field with the given name, and also
// checks for the JSON tag equivalent to the name. If no field is found, it looks for a method
// with the given name. For functions, it calls the function (if it doesn’t have inputs
// and has a single output) and recursively processes the return value. If the input value
// is of an unsupported type or any other error occurs, the function panics.
func Value(v reflect.Value, name string) reflect.Value {
	vv := v.Interface()
	if p, ok := vv.(Virtual); ok {
		return p.PathValue(name)
	}
	rtyp := v.Type() // was TypeOf(vv), but Interface() can force interface to struct kind
	switch rtyp.Kind() {
	case reflect.Slice, reflect.Array:
		idx, err := strconv.Atoi(name)
		if err != nil {
			panic("non-numeric index given for slice")
		}
		return v.Index(idx)
	case reflect.Ptr, reflect.Interface:
		if v.IsNil() {
			return reflect.Value{}
		}
		m := v.MethodByName(name)
		if m.IsValid() {
			return m
		}
		return Value(v.Elem(), name)
	case reflect.Map:
		return v.MapIndex(reflect.ValueOf(name))
	case reflect.Struct:
		f := v.FieldByName(name)
		if f.IsValid() {
			return f
		}
		for i := 0; i < rtyp.NumField(); i++ {
			// check by json name
			field := rtyp.Field(i)
			tag := strings.Split(field.Tag.Get("json"), ",")
			if tag[0] == name {
				return v.FieldByName(field.Name)
			}
		}
		return v.MethodByName(name)
	case reflect.Func:
		if v.Type().NumIn() != 0 || v.Type().NumOut() != 1 {
			panic("func is not traversable")
		}
		return Value(v.Call([]reflect.Value{})[0], name)
	}
	log.Panicf("telepath: unexpected kind: %s %s %s %s", v, reflect.TypeOf(v.Interface()).Kind(), v.Type(), v.Type().Kind())
	//panic("unexpected kind: " + rtyp.Kind().String() + " " + reflect.ValueOf(vv).String())
	return reflect.Value{}
}

// Names returns the names that can be "pathed" into
// from this value using Value(). Pointer values are
// first dereferenced. For maps, this returns the keys.
// For structs names are the fields and methods. For
// slices and arrays these are strings of the index values.
//
// If the value implements Virtual, the result of
// PathNames() will be used instead. A nil or
// unsupported value returns an empty slice.
func Names(v any) (names []string) {
	if p, ok := v.(Virtual); ok {
		return p.PathNames()
	}
	if v == nil {
		return []string{}
	}
	rv := reflect.Indirect(reflect.ValueOf(v))
	switch rv.Type().Kind() {
	case reflect.Map:
		for _, key := range rv.MapKeys() {
			k, ok := key.Interface().(string)
			if !ok {
				continue
			}
			names = append(names, k)
		}
		sort.Sort(sort.StringSlice(names))
	case reflect.Struct:
		t := rv.Type()
		for i := 0; i < t.NumField(); i++ {
			name := t.Field(i).Name
			// first letter capitalized means exported
			if name[0] == strings.ToUpper(name)[0] {
				names = append(names, name)
			}
		}
		for m := 0; m < reflect.PointerTo(t).NumMethod(); m++ {
			names = append(names, reflect.PointerTo(t).Method(m).Name)
		}
	case reflect.Slice, reflect.Array:
		for n := 0; n < rv.Len(); n++ {
			names = append(names, strconv.Itoa(n))
		}
	default:
	}
	return
}

func ensureType(v reflect.Value, t reflect.Type) reflect.Value {
	nv := v
	if v.Type().Kind() == reflect.Slice && v.Type().Elem() != t {
		switch t.Kind() {
		case reflect.Array:
			nv = reflect.Indirect(reflect.New(t))
			for i := 0; i < v.Len(); i++ {
				vv := reflect.ValueOf(v.Index(i).Interface())
				nv.Index(i).Set(vv.Convert(nv.Type().Elem()))
			}
		case reflect.Slice:
			nv = reflect.MakeSlice(t, 0, 0)
			for i := 0; i < v.Len(); i++ {
				vv := reflect.ValueOf(v.Index(i).Interface())
				nv = reflect.Append(nv, vv.Convert(nv.Type().Elem()))
			}
		default:
			panic("unable to convert slice to non-array, non-slice type")
		}
	}
	if t.Kind() == reflect.Pointer && v.Kind() != reflect.Pointer {
		if v.Type() != t.Elem() {
			nv = nv.Convert(t.Elem())
		}
		pv := reflect.New(t.Elem())
		pv.Elem().Set(nv)
		return pv
	}
	if t.Kind() == reflect.Float64 && v.Kind() == reflect.String {
		f, err := strconv.ParseFloat(v.String(), 64)
		if err != nil {
			panic(err)
		}
		nv = reflect.ValueOf(f)
	}
	if t.Kind() == reflect.Float32 && v.Kind() == reflect.String {
		f, err := strconv.ParseFloat(v.String(), 32)
		if err != nil {
			panic(err)
		}
		nv = reflect.ValueOf(f)
	}
	if t.Kind() == reflect.Int && v.Kind() == reflect.String {
		i, err := strconv.Atoi(v.String())
		if err != nil {
			panic(err)
		}
		nv = reflect.ValueOf(i)
	}
	if t.Kind() == reflect.Bool && v.Kind() == reflect.String {
		b := false
		if v.String() == "true" {
			b = true
		}
		nv = reflect.ValueOf(b)
	}
	if v.Type() != t {
		nv = nv.Convert(t)
	}
	return nv
}
