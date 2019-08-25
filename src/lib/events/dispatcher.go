package events

import (
	"container/list"
	"context"
	"sync"

	"github.com/hr3lxphr6j/bililive-go/src/instance"
)

func NewDispatcher(ctx context.Context) Dispatcher {
	ed := &dispatcher{
		saver: make(map[EventType]*list.List),
	}
	instance.GetInstance(ctx).EventDispatcher = ed
	return ed
}

type Dispatcher interface {
	AddEventListener(eventType EventType, listener *EventListener)
	RemoveEventListener(eventType EventType, listener *EventListener)
	RemoveAllEventListener(eventType EventType)
	DispatchEvent(event *Event)
}

type dispatcher struct {
	sync.RWMutex
	saver map[EventType]*list.List // map<EventType, List<*EventListener>>
}

func (e *dispatcher) Start(ctx context.Context) error {
	return nil
}

func (e *dispatcher) Close(ctx context.Context) {

}

func (e *dispatcher) AddEventListener(eventType EventType, listener *EventListener) {
	e.Lock()
	defer e.Unlock()
	listeners, ok := e.saver[eventType]
	if !ok || listener == nil {
		listeners = list.New()
		e.saver[eventType] = listeners
	}
	listeners.PushBack(listener)
}

func (e *dispatcher) RemoveEventListener(eventType EventType, listener *EventListener) {
	e.Lock()
	defer e.Unlock()
	listeners, ok := e.saver[eventType]
	if !ok || listeners == nil {
		return
	}
	for e := listeners.Front(); e != nil; e = e.Next() {
		if e.Value == listener {
			listeners.Remove(e)
		}
	}
	if listeners.Len() == 0 {
		delete(e.saver, eventType)
	}
}

func (e *dispatcher) RemoveAllEventListener(eventType EventType) {
	e.Lock()
	defer e.Unlock()
	e.saver = make(map[EventType]*list.List)
}

func (e *dispatcher) DispatchEvent(event *Event) {
	if event == nil {
		return
	}
	e.RLock()
	defer e.RUnlock()
	listeners, ok := e.saver[event.Type]
	if !ok || listeners == nil {
		return
	}
	for e := listeners.Front(); e != nil; e = e.Next() {
		handler := e.Value.(*EventListener)
		go handler.Handler(event)
	}
}
