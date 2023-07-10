package recorders

import "github.com/matyle/bililive-go/src/pkg/events"

const (
	RecorderStart events.EventType = "RecorderStart"
	RecorderStop  events.EventType = "RecorderStop"
	RecorderRestart  events.EventType = "RecorderRestart"
)
