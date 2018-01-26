package events

// 事件种类
type EventType string

// 事件回调
type EventHandler func(event *Event)

// 事件类型基类
type Event struct {
	Type   EventType
	Object interface{}
}

func NewEvent(eventType EventType, object interface{}) *Event {
	return &Event{eventType, object}
}

// 事件监听器
type EventListener struct {
	Handler EventHandler
}

func NewEventListener(handler EventHandler) *EventListener {
	return &EventListener{handler}
}

// 事件监听器集合
type eventListenerSet map[*EventListener]bool
