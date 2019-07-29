package listeners

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/hr3lxphr6j/bililive-go/src/api"
	"github.com/hr3lxphr6j/bililive-go/src/configs"
	"github.com/hr3lxphr6j/bililive-go/src/instance"
	"github.com/hr3lxphr6j/bililive-go/src/interfaces"
	"github.com/hr3lxphr6j/bililive-go/src/lib/events"
)

const (
	begin uint32 = iota
	pending
	running
	stopped
)

func NewListener(ctx context.Context, live api.Live) *Listener {
	inst := instance.GetInstance(ctx)
	return &Listener{
		Live:   live,
		status: false,
		config: inst.Config,
		stop:   make(chan struct{}),
		ed:     inst.EventDispatcher.(events.IEventDispatcher),
		logger: inst.Logger,
		state:  begin,
	}
}

type Listener struct {
	Live   api.Live
	status bool

	config *configs.Config
	ed     events.IEventDispatcher
	logger *interfaces.Logger

	state uint32
	stop  chan struct{}
}

func (l *Listener) Start() error {
	if !atomic.CompareAndSwapUint32(&l.state, begin, pending) {
		return nil
	}
	defer atomic.CompareAndSwapUint32(&l.state, pending, running)

	l.logger.WithFields(l.Live.GetInfoMap()).Info("Listener Start")
	l.ed.DispatchEvent(events.NewEvent(ListenStart, l.Live))
	l.refresh()
	go l.run()
	return nil
}

func (l *Listener) Close() {
	if !atomic.CompareAndSwapUint32(&l.state, running, stopped) {
		return
	}
	l.logger.WithFields(l.Live.GetInfoMap()).Info("Listener Close")
	l.ed.DispatchEvent(events.NewEvent(ListenStop, l.Live))
	close(l.stop)
}

func (l *Listener) refresh() {
	info, err := l.Live.GetInfo()
	if err != nil {
		return
	}
	if info.Status == l.status {
		return
	}
	l.status = info.Status
	if l.status {
		l.Live.SetLastStartTime(time.Now())
		l.logger.WithFields(l.Live.GetInfoMap()).Info("Live Start")
		l.ed.DispatchEvent(events.NewEvent(LiveStart, l.Live))
	} else {
		l.logger.WithFields(l.Live.GetInfoMap()).Info("Live End")
		l.ed.DispatchEvent(events.NewEvent(LiveEnd, l.Live))
	}
}

func (l *Listener) run() {
	ticker := time.NewTicker(time.Duration(l.config.Interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-l.stop:
			return
		case <-ticker.C:
			l.refresh()
		}
	}
}
