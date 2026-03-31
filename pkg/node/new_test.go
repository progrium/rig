package node_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/progrium/rig/pkg/node"
)

type ComponentA struct {
	FieldA string
}

func TestObject(t *testing.T) {
	obj := node.New("demo", ComponentA{FieldA: "Foo"}, node.Children{
		node.New("sub1", node.Attributes{"foo": "bar"}),
		node.New("sub2", node.Attributes{"foo": "bar"}, node.Children{
			node.New("subsub"),
		}),
	})
	v, _ := json.MarshalIndent(obj, "", "  ")
	fmt.Println(string(v))
}
