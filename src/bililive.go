package core

import (
	"bililive/src/lib/events"
	"bililive/src/listeners"
)

type Instance struct {
	EventDispatcher events.IEventDispatcher
	ListenerManager listeners.IListenerManager

}

