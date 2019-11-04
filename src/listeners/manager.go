package listeners

import (
	"context"
	"sync"

	"github.com/hr3lxphr6j/bililive-go/src/instance"
	"github.com/hr3lxphr6j/bililive-go/src/interfaces"
	"github.com/hr3lxphr6j/bililive-go/src/live"
)

// for test
var newListener = NewListener

func NewManager(ctx context.Context) Manager {
	lm := &manager{
		savers: make(map[live.ID]Listener),
	}
	instance.GetInstance(ctx).ListenerManager = lm
	return lm
}

type Manager interface {
	interfaces.Module
	AddListener(ctx context.Context, live live.Live) error
	RemoveListener(ctx context.Context, liveId live.ID) error
	GetListener(ctx context.Context, liveId live.ID) (Listener, error)
	HasListener(ctx context.Context, liveId live.ID) bool
}

type manager struct {
	lock   sync.RWMutex
	savers map[live.ID]Listener
}

func (m *manager) Start(ctx context.Context) error {
	inst := instance.GetInstance(ctx)
	if inst.Config.RPC.Enable || len(inst.Lives) > 0 {
		inst.WaitGroup.Add(1)
	}
	return nil
}

func (m *manager) Close(ctx context.Context) {
	m.lock.Lock()
	defer m.lock.Unlock()
	for id, listener := range m.savers {
		listener.Close()
		delete(m.savers, id)
	}
	inst := instance.GetInstance(ctx)
	inst.WaitGroup.Done()
}

func (m *manager) AddListener(ctx context.Context, live live.Live) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if _, ok := m.savers[live.GetLiveId()]; ok {
		return ErrListenerExist
	}
	listener := newListener(ctx, live)
	m.savers[live.GetLiveId()] = listener
	return listener.Start()
}

func (m *manager) RemoveListener(ctx context.Context, liveId live.ID) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	listener, ok := m.savers[liveId]
	if !ok {
		return ErrListenerNotExist
	}
	listener.Close()
	delete(m.savers, liveId)
	return nil
}

func (m *manager) GetListener(ctx context.Context, liveId live.ID) (Listener, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	listener, ok := m.savers[liveId]
	if !ok {
		return nil, ErrListenerNotExist
	}
	return listener, nil
}

func (m *manager) HasListener(ctx context.Context, liveId live.ID) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()
	_, ok := m.savers[liveId]
	return ok
}
