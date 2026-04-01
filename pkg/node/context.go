package node

import (
	"context"
)

type contextKeyType string

const contextKeyNode = contextKeyType("@node")

// FromContext returns the node stored in ctx, or nil when none is set.
func FromContext(ctx context.Context) Node {
	v := ctx.Value(contextKeyNode)
	if v == nil {
		return nil
	}

	return v.(Node)
}

// Context returns a derived context containing n as the current node.
func Context(ctx context.Context, n Node) context.Context {
	return context.WithValue(ctx, contextKeyNode, n)
}
