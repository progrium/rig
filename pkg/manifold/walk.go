package manifold

import "errors"

type WalkFunc func(node Node) error

var SkipAll = errors.New("skip everything and stop the walk")
var SkipNode = errors.New("skip this node")

func Walk(n Node, fn WalkFunc) error {
	err := walk(n, fn)
	if err == SkipNode || err == SkipAll {
		return nil
	}
	return err
}

func walk(n Node, fn WalkFunc) error {
	if err := fn(n); err != nil {
		return err
	}
	for _, com := range n.Components().Nodes() {
		if err := walk(com, fn); err != nil {
			if err == SkipNode {
				continue
			}
			if err == SkipAll {
				return nil
			}
			return err
		}
	}
	for _, child := range n.Objects().Nodes() {
		if err := walk(child, fn); err != nil {
			if err == SkipNode {
				continue
			}
			if err == SkipAll {
				return nil
			}
			return err
		}
	}
	return nil
}
