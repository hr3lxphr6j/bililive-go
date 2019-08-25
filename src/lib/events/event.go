package events

type EventType string

type EventHandler func(event *Event)

type Event struct {
	Type   EventType
	Object interface{}
}

func NewEvent(eventType EventType, object interface{}) *Event {
	return &Event{eventType, object}
}

type EventListener struct {
	Handler EventHandler
}

func NewEventListener(handler EventHandler) *EventListener {
	return &EventListener{handler}
}
