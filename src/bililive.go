package core

import (
	"bililive/src/lib/events"
	"bililive/src/listeners"
	"bililive/src/configs"
)

type Instance struct {
	Config          *configs.Config
	EventDispatcher events.IEventDispatcher
	ListenerManager listeners.IListenerManager
}
