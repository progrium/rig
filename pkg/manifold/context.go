package manifold

import (
	"context"

	"github.com/progrium/rig/pkg/node"
)

func FromContext(ctx context.Context) Node {
	e := node.FromContext(ctx)
	if e == nil {
		return nil
	}
	return FromEntity(e)
}
