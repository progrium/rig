package field

import (
	"math"
	"reflect"
	"strconv"
)

var TypeInteger = "Integer"

type Integer struct {
	FieldInfo
	sv int
	uv uint

	DefaultInteger string
	AllowNegative  bool
}

func IntegerFrom(v int, fi FieldInfo) Integer {
	return Integer{sv: v, AllowNegative: true, FieldInfo: fi}
}

func IntegerFromUint(v uint, fi FieldInfo) Integer {
	return Integer{uv: v, FieldInfo: fi}
}

func (f Integer) TypeName() string     { return TypeInteger }
func (f Integer) Type() reflect.Type   { return reflect.TypeOf(integerValue(f)) }
func (f Integer) Value() reflect.Value { return reflect.ValueOf(integerValue(f)) }
func (f Integer) Default() string      { return default_(f.DefaultInteger, "0") }
func (f Integer) Enum() []string       { return nil }

func (f Integer) Range() *Range {
	if f.AllowNegative {
		return &Range{
			Min:  math.MinInt64,
			Max:  math.MaxInt64,
			Step: 1,
		}
	}
	return &Range{
		Min:  0,
		Max:  math.MaxUint64,
		Step: 1,
	}
}

func (f Integer) Parse(s string) (v Value, err error) {
	ff := f
	if f.AllowNegative {
		var i int64
		i, err = strconv.ParseInt(s, 10, 0)
		ff.sv = int(i)
	} else {
		var i uint64
		i, err = strconv.ParseUint(s, 10, 0)
		ff.uv = uint(i)
	}
	v = ff
	return
}

func (f Integer) Format() string {
	if f.AllowNegative {
		return strconv.FormatInt(int64(f.sv), 10)
	}
	return strconv.FormatUint(uint64(f.uv), 10)
}

func integerValue(f Integer) interface{} {
	if f.AllowNegative {
		return f.sv
	}
	return f.uv
}
