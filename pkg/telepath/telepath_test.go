package telepath

import (
	"path/filepath"
	"reflect"
	"testing"

	"tractor.dev/toolkit-go/duplex/fn"
)

type testData struct {
	StringValue string
	IntValue    int
	BoolValue   bool
	MapValue    map[string]string
	SliceValue  []string
}

func (td *testData) Foo() string {
	return "foo"
}

func (td testData) Bar() string {
	return "bar"
}

func (td testData) Sum(a, b int) int {
	return a + b
}

func (td *testData) PtrSum(a, b int) int {
	return a + b
}

type testStructure struct {
	MapValue    map[string]*testData
	SliceValue  []testData
	StructValue testData
	PtrValue    *testData
}

func newTestStructure() testStructure {
	ptr := newTestData("qux")
	return testStructure{
		MapValue: map[string]*testData{
			"one": newTestData("foo"),
			"two": newTestData("bar"),
		},
		SliceValue: []testData{
			*newTestData("one"),
			*newTestData("two"),
		},
		StructValue: *newTestData("foobar"),
		PtrValue:    ptr,
	}
}

func newTestData(s string) *testData {
	return &testData{
		StringValue: s,
		IntValue:    100,
		BoolValue:   true,
		MapValue: map[string]string{
			"one": s,
			"two": s,
		},
		SliceValue: []string{"one", "two"},
	}
}

func TestFuncValue(t *testing.T) {
	data := newTestStructure()
	root := New(&data)
	foo, _ := root.Select("PtrValue/Foo").Value()
	if foo != "foo" {
		t.Fatal("unexpected value from Foo")
	}
	var ret string
	if err := root.Select("PtrValue/Foo").Call(fn.Args{}, &ret); err != nil {
		t.Fatal(err)
	}
	if ret != "foo" {
		t.Fatal("unexpected value from Foo")
	}
	bar, err := root.Select("StructValue/Bar").Value()
	if err != nil {
		t.Fatal(err)
	}
	if bar != "bar" {
		t.Fatal("unexpected value from Bar")
	}
	if err := root.Select("StructValue/Bar").Call(fn.Args{}, &ret); err != nil {
		t.Fatal(err)
	}
	if ret != "bar" {
		t.Fatal("unexpected value from Bar")
	}
	var num int
	if err := root.Select("StructValue/Sum").Call(fn.Args{1, 2}, &num); err != nil {
		t.Fatal(err)
	}
	if num != 3 {
		t.Fatal("unexpected value from Sum")
	}
	if err := root.Select("PtrValue/PtrSum").Call(fn.Args{2, 3}, &num); err != nil {
		t.Fatal(err)
	}
	if num != 5 {
		t.Fatal("unexpected value from PtrSum")
	}
}

func TestCursorValue(t *testing.T) {
	data := newTestStructure()
	view := New(&data)

	for _, tt := range []struct {
		in  string
		out string
	}{
		{"StructValue/StringValue", "foobar"},
		{"PtrValue/StringValue", "qux"},
		{"MapValue/one/StringValue", "foo"},
	} {
		t.Run(tt.in, func(t *testing.T) {
			got, _ := view.Select(tt.in).Value()

			if got != tt.out {
				t.Fatalf("expected '%#v' but got '%#v'", tt.out, got)
			}
		})
	}

}

func TestCursorValueTo(t *testing.T) {
	data := newTestStructure()
	view := New(&data)

	// STRINGS

	for _, tt := range []struct {
		in  string
		out string
	}{
		{"StructValue/StringValue", "foobar"},
		{"PtrValue/StringValue", "qux"},
		{"MapValue/one/StringValue", "foo"},
	} {
		t.Run(tt.in, func(t *testing.T) {
			var got string
			view.Select(tt.in).ValueTo(&got)

			if got != tt.out {
				t.Fatalf("expected '%#v' but got '%#v'", tt.out, got)
			}
		})
	}

	// MAPS

	for _, tt := range []struct {
		in  string
		out string
	}{
		{"StructValue/MapValue", "foobar"},
		{"PtrValue/MapValue", "qux"},
		{"MapValue/one/MapValue", "foo"},
	} {
		t.Run(tt.in, func(t *testing.T) {
			var got map[string]string
			view.Select(tt.in).ValueTo(&got)

			if got["one"] != tt.out || got["two"] != tt.out {
				t.Fatalf("expected values '%#v' but got '%#v'", tt.out, got)
			}
		})
	}

}

func TestCursorSet(t *testing.T) {
	data := newTestStructure()
	view := New(&data)

	// strings

	for _, tt := range []struct {
		in  string
		out string
	}{
		{"StructValue/StringValue", "foobar2"},
		{"SliceValue/0/StringValue", "foobar2"},
		{"MapValue/one/StringValue", "foobar2"},
	} {
		t.Run(tt.in, func(t *testing.T) {
			cur := view.Select(tt.in)
			cur.Set(tt.out)
			got, _ := cur.Value()

			if got != tt.out {
				t.Fatalf("expected '%#v' but got '%#v'", tt.out, got)
			}
		})
	}

	// maps

	for _, tt := range []struct {
		in  string
		out interface{}
	}{
		{"MapValue/one", newTestData("one")},
		{"StructValue/MapValue/two", "test"},
	} {
		t.Run(tt.in, func(t *testing.T) {
			cur := view.Select(tt.in)
			cur.Set(tt.out)
			got, _ := cur.Value()

			if !reflect.DeepEqual(got, New(tt.out).Value) {
				t.Fatalf("expected '%#v' but got '%#v'", tt.out, got)
			}
		})
	}

}

func TestCursorDelete(t *testing.T) {
	data := newTestStructure()
	view := New(&data)

	// strings

	for _, tt := range []struct {
		in  string
		out string
	}{
		{"StructValue/StringValue", ""},
		{"SliceValue/0/StringValue", ""},
	} {
		t.Run(tt.in, func(t *testing.T) {
			cur := view.Select(tt.in)
			cur.Delete()
			got, _ := cur.Value()

			if got != tt.out {
				t.Fatalf("expected '%#v' but got '%#v'", tt.out, got)
			}
		})
	}

	// maps

	for _, tt := range []struct {
		in  string
		out []string
	}{
		{"MapValue/one", []string{"two"}},
		{"StructValue/MapValue/two", []string{"one"}},
	} {
		t.Run(tt.in, func(t *testing.T) {
			cur := view.Select(filepath.Dir(tt.in))
			cur.Select(filepath.Base(tt.in)).Delete()
			got, _ := cur.List()

			if !reflect.DeepEqual(got, tt.out) {
				t.Fatalf("expected '%#v' but got '%#v'", tt.out, got)
			}
		})
	}

}

// func TestCursorAppend(t *testing.T) {
// 	data := newTestStructure()
// 	view := New(&data)

// 	for _, tt := range []struct {
// 		in  string
// 		out []string
// 	}{
// 		{"StructValue/SliceValue", []string{"one", "two", "three", "four"}},
// 	} {
// 		t.Run(tt.in, func(t *testing.T) {
// 			cur := view.Select(tt.in)
// 			cur.Append("three")
// 			cur.Append("four")
// 			got := cur.Value().([]string)

// 			if !reflect.DeepEqual(got, tt.out) {
// 				t.Fatalf("expected '%#v' but got '%#v'", tt.out, got)
// 			}
// 		})
// 	}
// }

func TestCursorInsert(t *testing.T) {
	data := newTestStructure()
	view := New(&data)

	for _, tt := range []struct {
		in  string
		out []string
	}{
		{"StructValue/SliceValue", []string{"one", "one-half", "two"}},
	} {
		t.Run(tt.in, func(t *testing.T) {
			cur := view.Select(tt.in)
			cur.Insert(1, "one-half")
			got, _ := cur.Value()

			if !reflect.DeepEqual(got, tt.out) {
				t.Fatalf("expected '%#v' but got '%#v'", tt.out, got)
			}
		})
	}
}

// func TestMetaSchema(t *testing.T) {
// 	root := New(&testData{})
// 	md, err := root.Meta("")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	b, _ := json.MarshalIndent(md.Schema, "", "  ")
// 	fmt.Println(string(b))
// }
