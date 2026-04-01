package node

import "errors"

// WalkFunc is called for each node during Walk.
// Return nil to continue. Return SkipNode or SkipAll to alter traversal (see their docs).
type WalkFunc func(node Node) error

// SkipAll requests that the walk stop. Walk returns nil to its caller when the walk ends this way.
var SkipAll = errors.New("skip everything and stop the walk")

// SkipNode requests skipping the current node's descendants when returned from enterVisitor
// before children are visited, or advancing past the current branch when returned while
// visiting children (the parent continues with its next subnode).
var SkipNode = errors.New("skip this node")

// Walk performs a depth-first traversal starting at n.
// enterVisitor runs before visiting a node's component subnodes; exitVisitor runs after.
// Either visitor may be nil. Only nodes implementing SubnodesNode are descended, via
// GetSubnodes with kind TypeComponent or TypeObject.
//
// If the inner walk returns SkipNode or SkipAll, Walk returns nil (those values are not
// propagated to Walk's caller).
func Walk(n Node, enterVisitor, exitVisitor WalkFunc) error {
	err := walk(n, enterVisitor, exitVisitor)
	if err == SkipNode || err == SkipAll {
		return nil
	}
	return err
}

func walk(n Node, enterVisitor, exitVisitor WalkFunc) error {
	if enterVisitor != nil {
		if err := enterVisitor(n); err != nil {
			return err
		}
	}
	if sn, ok := Unwrap[SubnodesNode](n); ok {
		for _, com := range sn.GetSubnodes(TypeComponent) {
			if err := walk(com, enterVisitor, exitVisitor); err != nil {
				if err == SkipNode {
					continue
				}
				if err == SkipAll {
					return nil
				}
				return err
			}
		}
		for _, child := range sn.GetSubnodes(TypeObject) {
			if err := walk(child, enterVisitor, exitVisitor); err != nil {
				if err == SkipNode {
					continue
				}
				if err == SkipAll {
					return nil
				}
				return err
			}
		}
	}
	if exitVisitor != nil {
		if err := exitVisitor(n); err != nil {
			return err
		}
	}
	return nil
}
