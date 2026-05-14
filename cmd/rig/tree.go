package main

import (
	"fmt"
	"log"
	"path"
	"strings"

	"github.com/progrium/rig/pkg/manifold"
	"github.com/progrium/rig/pkg/node"
	"tractor.dev/toolkit-go/engine/cli"
)

func treeCmd() *cli.Command {
	var showID bool
	cmd := &cli.Command{
		Usage: "tree [--id] [<node>]",
		Short: "Display a hierarchy of objects",
		Long:  `Display a hierarchy of objects. Use --id to show node IDs.`,
		Args:  cli.RangeArgs(0, 1),
		Run: func(_ *cli.Context, args []string) {
			tree(args, showID)
		},
	}
	cmd.Flags().BoolVar(&showID, "id", false, "print node id after each label in braces")
	return cmd
}

func tree(args []string, showID bool) {
	_, realm := dialManifold()
	id := "@main"
	if len(args) == 1 {
		id = args[0]
	}
	raw := realm.Resolve(id)
	if raw == nil {
		log.Fatal("node not found")
	}
	n := manifold.FromNode(raw)
	writeTree(n, "", false, true, showID)
}

// treeChildren matches manifold.Walk order: components first, then object children.
func treeChildren(n manifold.Node) []manifold.Node {
	var out []manifold.Node
	out = append(out, n.Components().Nodes()...)
	out = append(out, n.Children().Nodes()...)
	return out
}

func treeLabel(n manifold.Node) string {
	if n.Kind() == node.TypeComponent {
		t := n.ComponentType()
		if t == "" {
			t = n.Name()
		}
		return "(" + path.Base(t) + ")"
	}
	return n.Name()
}

func treeLineLabel(n manifold.Node, showID bool) string {
	s := treeLabel(n)
	if showID {
		s += " {" + n.NodeID() + "}"
	}
	return s
}

func writeTree(n manifold.Node, prefix string, isLast bool, isRoot bool, showID bool) {
	var line strings.Builder
	if isRoot {
		line.WriteString(treeLineLabel(n, showID))
	} else {
		if isLast {
			line.WriteString(prefix)
			line.WriteString("└─ ")
		} else {
			line.WriteString(prefix)
			line.WriteString("├─ ")
		}
		line.WriteString(treeLineLabel(n, showID))
	}
	fmt.Println(line.String())

	children := treeChildren(n)
	for i, c := range children {
		last := i == len(children)-1
		var childPrefix string
		if isRoot {
			childPrefix = ""
		} else if isLast {
			childPrefix = prefix + "   "
		} else {
			childPrefix = prefix + "│  "
		}
		writeTree(c, childPrefix, last, false, showID)
	}
}
