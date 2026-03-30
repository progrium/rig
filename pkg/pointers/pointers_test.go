package pointers

import (
	"reflect"
	"testing"
)

type TypeA struct {
	Slice []*TypeA
}
type TypeB struct {
	A *TypeA
	C *TypeC
}
type TypeC struct {
	Empty *TypeA
	A     *TypeA
	B     *TypeB
}

func TestPointersFrom(t *testing.T) {
	as0 := &TypeA{}
	as1 := &TypeA{}
	a := &TypeA{
		Slice: []*TypeA{
			as0,
			as1,
		},
	}
	ba := &TypeA{}
	b := &TypeB{
		A: ba,
	}
	v := &TypeC{
		A: a,
		B: b,
	}
	ptrs := From(v)
	if !reflect.DeepEqual(ptrs, map[string]any{
		"A":         a,
		"A/Slice/0": as0,
		"A/Slice/1": as1,
		"B":         b,
		"B/A":       ba,
	}) {
		t.Fatal("unexpected output")
	}
}
