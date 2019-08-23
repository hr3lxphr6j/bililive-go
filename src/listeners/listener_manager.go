package listeners

import (
	"context"
	"sync"

	"github.com/hr3lxphr6j/bililive-go/src/instance"
	"github.com/hr3lxphr6j/bililive-go/src/live"
)

func NewIListenerManager(ctx context.Context) IListenerManager {
	lm := &ListenerManager{
		savers: make(map[live.ID]*Listener),
		lock:   new(sync.RWMutex),
	}
	instance.GetInstance(ctx).ListenerManager = lm
	return lm
}

// 监听管理器接口
type IListenerManager interface {
	AddListener(ctx context.Context, live live.Live) error
	RemoveListener(ctx context.Context, liveId live.ID) error
	GetListener(ctx context.Context, liveId live.ID) (*Listener, error)
	HasListener(ctx context.Context, liveId live.ID) bool
}

type ListenerManager struct {
	savers map[live.ID]*Listener
	lock   *sync.RWMutex
}

func (l *ListenerManager) Start(ctx context.Context) error {
	inst := instance.GetInstance(ctx)
	if inst.Config.RPC.Enable || len(inst.Lives) > 0 {
		inst.WaitGroup.Add(1)
	}
	return nil
}

func (l *ListenerManager) Close(ctx context.Context) {
	l.lock.Lock()
	defer l.lock.Unlock()
	for id, listener := range l.savers {
		listener.Close()
		delete(l.savers, id)
	}
	inst := instance.GetInstance(ctx)
	inst.WaitGroup.Done()
}

func (l *ListenerManager) AddListener(ctx context.Context, live live.Live) error {
	l.lock.Lock()
	defer l.lock.Unlock()

	if _, ok := l.savers[live.GetLiveId()]; ok {
		return listenerExistError
	}
	listener := NewListener(ctx, live)
	listener.Start()
	l.savers[live.GetLiveId()] = listener
	return nil
}

func (l *ListenerManager) RemoveListener(ctx context.Context, liveId live.ID) error {
	l.lock.Lock()
	defer l.lock.Unlock()

	if listener, ok := l.savers[liveId]; !ok {
		return listenerNotExistError
	} else {
		listener.Close()
		delete(l.savers, liveId)
		return nil
	}
}

func (l *ListenerManager) GetListener(ctx context.Context, liveId live.ID) (*Listener, error) {
	l.lock.RLock()
	defer l.lock.RUnlock()
	if r, ok := l.savers[liveId]; !ok {
		return nil, listenerNotExistError
	} else {
		return r, nil
	}
}

func (l *ListenerManager) HasListener(ctx context.Context, liveId live.ID) bool {
	l.lock.RLock()
	defer l.lock.RUnlock()
	_, ok := l.savers[liveId]
	return ok
}
