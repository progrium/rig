package field

import (
	"fmt"
	"testing"
)

type TestStruct struct {
	Int   int
	Str   string
	Bool  bool
	Bytes []byte
}

func TestField(t *testing.T) {
	txt := TextFrom("Hello world", FieldInfo{})
	if txt.Value().String() != "Hello world" {
		t.Fatal("unexpected")
	}
	if txt.TypeName() != "Text" {
		t.Fatal("unexpected")
	}

	i := IntegerFrom(123, FieldInfo{})
	if i.Value().Int() != 123 {
		t.Fatal("unexpected")
	}
	if i.TypeName() != "Integer" {
		t.Fatal("unexpected")
	}

	fi := WithFieldInfo("Foobar", "", FlagHidden, FlagReadonly)
	if fi.Name() != "Foobar" {
		t.Fatal("unexpected")
	}
	f := StringFrom("Hello again", fi)
	if Hidden(f) != true {
		t.Fatal("unexpected")
	}
	if Readonly(f) != true {
		t.Fatal("unexpected")
	}
	if Required(f) != false {
		t.Fatal("unexpected")
	}

	fmt.Println(StructFrom(Decimal{}, FieldInfo{}).Fields())
}
