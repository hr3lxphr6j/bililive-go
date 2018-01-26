package core

import (
	"context"
	"time"
)

type key int

const (
	InstanceKey key = 114514
	IntervalKey key = 1919
)

func GetInstance(ctx context.Context) *Instance {
	if s, ok := ctx.Value(InstanceKey).(*Instance); ok {
		return s
	}
	return nil
}

func GetInterval(ctx context.Context) time.Duration {
	if s, ok := ctx.Value(IntervalKey).(time.Duration); ok {
		return s
	}
	return 15 * time.Second
}
