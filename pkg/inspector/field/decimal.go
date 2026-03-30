package field

import (
	"reflect"

	"github.com/shopspring/decimal"
)

var TypeDecimal = "Decimal"

type Decimal struct {
	FieldInfo
	v decimal.Decimal

	DefaultDecimal string
	Precision      int
}

func DecimalFrom(v decimal.Decimal, fi FieldInfo) Decimal {
	return Decimal{v: v, FieldInfo: fi}
}

func (f Decimal) TypeName() string     { return TypeDecimal }
func (f Decimal) Type() reflect.Type   { return reflect.TypeOf(f.v) }
func (f Decimal) Value() reflect.Value { return reflect.ValueOf(f.v) }
func (f Decimal) Default() string      { return default_(f.DefaultDecimal, "0.0") }
func (f Decimal) Enum() []string       { return nil }
func (f Decimal) Range() *Range        { return nil }

func (f Decimal) Parse(s string) (v Value, err error) {
	if s == "." {
		s = f.Default()
	}
	ff := f
	ff.v, err = decimal.NewFromString(s)
	v = ff
	return
}

func (f Decimal) Format() string {
	return f.v.StringFixed(int32(f.Precision))
}
