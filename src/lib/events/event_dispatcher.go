package events

import (
	"context"
	"github.com/hr3lxphr6j/bililive-go/src/instance"
)

func NewIEventDispatcher(ctx context.Context) IEventDispatcher {
	ed := &EventDispatcher{
		saver: make(map[EventType]eventListenerSet),
	}
	instance.GetInstance(ctx).EventDispatcher = ed
	return ed
}

// 事件分发器接口
type IEventDispatcher interface {
	AddEventListener(eventType EventType, listener *EventListener)
	RemoveEventListener(eventType EventType, listener *EventListener)
	RemoveAllEventListener(eventType EventType)
	DispatchEvent(event *Event)
}

// 事件分发器
type EventDispatcher struct {
	saver map[EventType]eventListenerSet
}

func (e *EventDispatcher) Start(ctx context.Context) error {
	return nil
}

func (e *EventDispatcher) Close(ctx context.Context) {

}

func (e *EventDispatcher) AddEventListener(eventType EventType, listener *EventListener) {
	_, isExist := e.saver[eventType]
	if !isExist {
		e.saver[eventType] = make(map[*EventListener]bool)
	}
	e.saver[eventType][listener] = true
}

func (e *EventDispatcher) RemoveEventListener(eventType EventType, listener *EventListener) {
	delete(e.saver[eventType], listener)
}

func (e *EventDispatcher) RemoveAllEventListener(eventType EventType) {
	delete(e.saver, eventType)
}

func (e *EventDispatcher) DispatchEvent(event *Event) {
	for l := range e.saver[event.Type] {
		go l.Handler(event)
	}
}
