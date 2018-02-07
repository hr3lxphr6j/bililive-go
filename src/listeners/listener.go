package listeners

import (
	"context"
	"github.com/hr3lxphr6j/bililive-go/src/api"
	"github.com/hr3lxphr6j/bililive-go/src/instance"
	"github.com/hr3lxphr6j/bililive-go/src/lib/events"
	"time"
)

func NewListener(ctx context.Context, live api.Live) *Listener {
	inst := instance.GetInstance(ctx)
	return &Listener{
		Live:   live,
		status: false,
		ticker: time.NewTicker(time.Duration(inst.Config.Interval) * time.Second),
		stop:   make(chan struct{}),
		ed:     inst.EventDispatcher.(events.IEventDispatcher),
	}
}

type Listener struct {
	Live api.Live

	status bool
	ticker *time.Ticker
	stop   chan struct{}
	ed     events.IEventDispatcher
}

func (l *Listener) Start() error {
	info, _ := l.Live.GetRoom()
	l.ed.DispatchEvent(events.NewEvent(ListenStart, info))
	go l.run()
	return nil
}

func (l *Listener) Close() {
	close(l.stop)
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
			info, err := l.Live.GetRoom()
			if err != nil {
				continue
			}
			if info.Status == l.status {
				continue
			}
			l.status = info.Status
			if l.status {
				l.ed.DispatchEvent(events.NewEvent(LiveStart, info))
			} else {
				l.ed.DispatchEvent(events.NewEvent(LiveEnd, info))
			}
		}
	}
}
