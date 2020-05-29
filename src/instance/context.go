package instance

import (
	"context"
)

type key int

const (
	Key key = 114514
)

func GetInstance(ctx context.Context) *Instance {
	if s, ok := ctx.Value(Key).(*Instance); ok {
		return s
	}
	return nil
}
