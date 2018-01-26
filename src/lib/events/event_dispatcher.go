package events

import (
	"sync"
)

// 事件分发器接口
type IEventDispatcher interface {
	AddEventListener(eventType EventType, listener *EventListener) error
	RemoveEventListener(eventType EventType, listener *EventListener) error
	AddEvent(eventType EventType) error
	RemoveEvent(eventType EventType) error
	HasEvent(eventType EventType) bool
	DispatchEvent(event *Event) error
}

func NewIEventDispatcher() IEventDispatcher {
	return new(EventDispatcher)
}

// 事件分发器
type EventDispatcher struct {
	saver map[EventType]*eventListenerSet
	lock  sync.RWMutex
}

func (e *EventDispatcher) AddEventListener(eventType EventType, listener *EventListener) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	set, ok := e.saver[eventType]
	if !ok {
		return eventNotExistError
	}

	_, ok = (*set)[listener]
	if ok {
		return listenerExistError
	} else {
		(*set)[listener] = true
		return nil
	}
}

func (e *EventDispatcher) RemoveEventListener(eventType EventType, listener *EventListener) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	set, ok := e.saver[eventType]
	if !ok {
		return eventNotExistError
	}

	_, ok = (*set)[listener]
	if ok {
		delete(*set, listener)
		return nil
	} else {
		return listenerNotExistError
	}
}

func (e *EventDispatcher) AddEvent(eventType EventType) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	_, ok := e.saver[eventType]
	if ok {
		return eventExistError
	} else {
		set := new(eventListenerSet)
		e.saver[eventType] = set
		return nil
	}
}

func (e *EventDispatcher) RemoveEvent(eventType EventType) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	set, ok := e.saver[eventType]
	if !ok {
		return eventNotExistError
	} else {
		delete(*set, eventType)
		return nil
	}
}

func (e *EventDispatcher) HasEvent(eventType EventType) bool {
	e.lock.RLock()
	defer e.lock.RLock()
	_, ok := e.saver[eventType]
	return ok
}

func (e *EventDispatcher) DispatchEvent(event *Event) error {
	e.lock.RLock()
	defer e.lock.RLock()
	set, ok := e.saver[event.Type]
	if !ok {
		return eventNotExistError
	}
	for l := range *set {
		l.Handler(event)
	}
	return nil
}
