package manifold

import (
	"fmt"
	"io"
	"os"
)

// func Dump(b Bus) {
// 	roots, err := r.Roots()
// 	if err != nil {
// 		panic(err)
// 	}
// 	for _, n := range roots {
// 		DumpNode(n)
// 	}
// }

func DumpNode(n Node) {
	dumpNode(os.Stdout, n, "", "  ")
}

func dumpNode(w io.Writer, n Node, prefix, indent string) {
	dumpf(w, n, prefix, indent, func(n Node) string {
		return fmt.Sprintf("%s [%s:%s]\n", n.Name(), "", n.ID())
	})
}

func dumpf(w io.Writer, n Node, prefix, indent string, format func(Node) string) {
	fmt.Fprintf(w, "%s%s\n", prefix, format(n))
	coms := n.Components().Nodes()
	if len(coms) > 0 {
		fmt.Fprintln(w, prefix+indent+"Components")
		for _, com := range coms {
			dumpf(w, com, prefix+indent+indent, indent, format)
		}
	}
	for _, child := range n.Children().Nodes() {
		dumpf(w, child, prefix+indent, indent, format)
	}
}
