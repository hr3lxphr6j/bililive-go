package listeners

import (
	"bililive/src/api"
	"sync"
	"context"
	"bililive/src"
	"time"
)

type IListenerManager interface {
	AddListener(ctx context.Context, live api.Live) error
	RemoveListener(ctx context.Context, live api.Live) error
	HasListener(ctx context.Context, live api.Live) bool
}

type ListenerManager struct {
	savers map[api.Live]*Listener
	lock   sync.RWMutex
}

func (l *ListenerManager) verifyLive(live api.Live) bool {
	for i := 0; i < 5; i++ {
		_, err := live.GetRoom()
		if err == nil {
			return true
		}
		if api.IsRoomNotExistsError(err) {
			return false
		}
	}
	return false
}

func (l *ListenerManager) AddListener(ctx context.Context, live api.Live) error {

	if !l.verifyLive(live) {
		return roomNotExistError
	}

	l.lock.Lock()
	defer l.lock.Unlock()

	_, ok := l.savers[live]
	if ok {
		return listenerExistError
	}
	listener := &Listener{
		Live:   live,
		ticker: time.NewTicker(core.GetInterval(ctx)),
		ed:     core.GetInstance(ctx).EventDispatcher,
		stop:   make(chan struct{}),
		status: false,
	}
	l.savers[live] = listener
	listener.Start()
	return nil
}

func (l *ListenerManager) RemoveListener(ctx context.Context, live api.Live) error {
	l.lock.Lock()
	defer l.lock.Unlock()

	listener, ok := l.savers[live]
	if !ok {
		return listenerNotExistError
	}
	listener.Close()
	delete(l.savers, live)
	return nil
}

func (l *ListenerManager) HasListener(ctx context.Context, live api.Live) bool {
	l.lock.RLock()
	defer l.lock.RUnlock()
	_, ok := l.savers[live]
	return ok
}
