package node

import (
	"context"

)

type contextKeyType string

const contextKeyNode = contextKeyType("@node")

func FromContext(ctx context.Context) Node {
	v := ctx.Value(contextKeyNode)
	if v == nil {
		return nil
	}

	return v.(Node)
}

func Context(ctx context.Context, n Node) context.Context {
	return context.WithValue(ctx, contextKeyNode, n)
}
