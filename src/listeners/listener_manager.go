package listeners

import (
	"context"
	"github.com/hr3lxphr6j/bililive-go/src/api"
	"github.com/hr3lxphr6j/bililive-go/src/instance"
	"net/url"
	"sync"
)

func NewIListenerManager(ctx context.Context) IListenerManager {
	lm := &ListenerManager{
		savers: make(map[api.Live]*Listener),
		lock:   new(sync.RWMutex),
	}
	instance.GetInstance(ctx).ListenerManager = lm
	return lm
}

// 监听管理器接口
type IListenerManager interface {
	AddListener(ctx context.Context, live api.Live) error
	RemoveListener(ctx context.Context, live api.Live) error
	HasListener(ctx context.Context, live api.Live) bool
}

type ListenerManager struct {
	savers map[api.Live]*Listener
	lock   *sync.RWMutex
}

// 验证直播间有效性
func (l *ListenerManager) verifyLive(live api.Live) bool {
	for i := 0; i < 3; i++ {
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
	if live == nil || !l.verifyLive(live) {
		return roomNotExistError
	}

	l.lock.Lock()
	defer l.lock.Unlock()

	if _, ok := l.savers[live]; ok {
		return listenerExistError
	}
	listener := NewListener(ctx, live)
	listener.Start()
	l.savers[live] = listener
	return nil
}

func (l *ListenerManager) RemoveListener(ctx context.Context, live api.Live) error {
	l.lock.Lock()
	defer l.lock.Unlock()

	if listener, ok := l.savers[live]; !ok {
		return listenerNotExistError
	} else {
		listener.Close()
		delete(l.savers, live)
		return nil
	}
}

func (l *ListenerManager) HasListener(ctx context.Context, live api.Live) bool {
	l.lock.RLock()
	defer l.lock.RUnlock()
	_, ok := l.savers[live]
	return ok
}

func (l *ListenerManager) Start(ctx context.Context) error {
	inst := instance.GetInstance(ctx)
	inst.WaitGroup.Add(1)

	for _, room := range instance.GetInstance(ctx).Config.LiveRooms {
		u, err := url.Parse(room)
		if err != nil {
			instance.GetInstance(ctx).Logger.Error(err)
		}
		err = l.AddListener(ctx, api.NewLive(u))
		if err != nil {
			instance.GetInstance(ctx).Logger.WithFields(map[string]interface{}{"Url": room}).Error(err)
		}
	}
	return nil
}

func (l *ListenerManager) Close(ctx context.Context) {
	l.lock.Lock()
	defer l.lock.Unlock()

	for _, listener := range l.savers {
		go listener.Close()
	}
	inst := instance.GetInstance(ctx)
	inst.WaitGroup.Done()
}
