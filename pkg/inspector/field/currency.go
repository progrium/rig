package field

import (
	"reflect"

	"github.com/shopspring/decimal"
)

var TypeCurrency = "Currency"

type Currency struct {
	FieldInfo
	v decimal.Decimal

	Symbol       string
	DefaultValue string
	Precision    int
}

func CurrencyFrom(v decimal.Decimal, fi FieldInfo) Currency {
	return Currency{v: v, Symbol: "$", Precision: 2, FieldInfo: fi}
}

func (f Currency) TypeName() string     { return TypeCurrency }
func (f Currency) Type() reflect.Type   { return reflect.TypeOf(f.v) }
func (f Currency) Value() reflect.Value { return reflect.ValueOf(f.v) }
func (f Currency) Default() string      { return default_(f.DefaultValue, "0.00") }
func (f Currency) Enum() []string       { return nil }
func (f Currency) Range() *Range        { return nil }

func (f Currency) Parse(s string) (v Value, err error) {
	if s == "." {
		s = f.Default()
	}
	ff := f
	ff.v, err = decimal.NewFromString(s)
	v = ff
	return
}

func (f Currency) Format() string {
	return f.Symbol + f.v.StringFixed(int32(f.Precision))
}
