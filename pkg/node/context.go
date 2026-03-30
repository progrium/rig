package node

import (
	"context"

	"github.com/progrium/rig/pkg/entity"
)

type contextKeyType string

const contextKeyNode = contextKeyType("@node")

func FromContext(ctx context.Context) entity.Node {
	v := ctx.Value(contextKeyNode)
	if v == nil {
		return nil
	}

	return v.(entity.Node)
}

func Context(ctx context.Context, n entity.Node) context.Context {
	return context.WithValue(ctx, contextKeyNode, n)
}
