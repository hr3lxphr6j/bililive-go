package listeners

import (
	"context"
	"time"

	"github.com/hr3lxphr6j/bililive-go/src/api"
	"github.com/hr3lxphr6j/bililive-go/src/instance"
	"github.com/hr3lxphr6j/bililive-go/src/interfaces"
	"github.com/hr3lxphr6j/bililive-go/src/lib/events"
)

func NewListener(ctx context.Context, live api.Live) *Listener {
	inst := instance.GetInstance(ctx)
	return &Listener{
		Live:   live,
		status: false,
		ticker: time.NewTicker(time.Duration(inst.Config.Interval) * time.Second),
		stop:   make(chan struct{}),
		ed:     inst.EventDispatcher.(events.IEventDispatcher),
		logger: inst.Logger,
	}
}

type Listener struct {
	Live api.Live

	status bool
	ticker *time.Ticker
	stop   chan struct{}
	ed     events.IEventDispatcher
	logger *interfaces.Logger
}

func (l *Listener) Start() error {
	l.logger.WithFields(l.Live.GetInfoMap()).Info("Listener Start")
	l.ed.DispatchEvent(events.NewEvent(ListenStart, l.Live))
	l.refresh()
	go l.run()
	return nil
}

func (l *Listener) Close() {
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
		l.logger.WithFields(l.Live.GetInfoMap()).Info("Live Start")
		l.ed.DispatchEvent(events.NewEvent(LiveStart, l.Live))
	} else {
		l.logger.WithFields(l.Live.GetInfoMap()).Info("Live End")
		l.ed.DispatchEvent(events.NewEvent(LiveEnd, l.Live))
	}
}

func (l *Listener) run() {
	defer func() {
		l.ticker.Stop()
	}()

	for {
		select {
		case <-l.stop:
			return
		case <-l.ticker.C:
			l.refresh()
		}
	}
}
