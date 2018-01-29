package recorders

import (
	"context"
	"bililive/src/api"
	"time"
)

type IRecorderManager interface {
	AddRecorder(ctx context.Context, live api.Live) error
	RemoveRecorder(ctx context.Context, live api.Live) (time.Duration, error)
}
type RecorderManager struct {
}
