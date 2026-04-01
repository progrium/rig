package manifold

import (
	"testing"

	"github.com/progrium/rig/pkg/node"
)

type FooCom struct {
	Foo string
}

type BarCom struct {
	Bar string
}

func TestNode(t *testing.T) {
	n := FromNode(node.New("root",
		node.Attributes{
			"foo": "bar",
		},
		node.Children{
			node.NewRaw("child1", nil, ""),
			node.NewRaw("child2", "value", ""),
		},
		&FooCom{
			Foo: "foo!",
		},
		&BarCom{
			Bar: "bar!",
		},
	))

	if n.Attr("foo") != "bar" {
		t.Fatal("unexpected attribute")
	}

	if n.Objects().Count() != 2 {
		t.Fatal("unexpected children count")
	}

	if n.Components().Count() != 2 {
		t.Fatal("unexpected component count")
	}

	if n.Objects().Nodes()[1].Value() != "value" {
		t.Fatal("unexpected child value")
	}

	if node.Get[*BarCom](n).Bar != "bar!" {
		t.Fatal("unexpected component value")
	}

}
