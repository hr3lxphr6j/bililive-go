package core

import (
	"context"
)

type key int

const (
	InstanceKey key = 114514
)

func GetInstance(ctx context.Context) *Instance {
	if s, ok := ctx.Value(InstanceKey).(*Instance); ok {
		return s
	}
	return nil
}
