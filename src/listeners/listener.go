//go:generate mockgen -package listeners -destination mock_test.go github.com/hr3lxphr6j/bililive-go/src/listeners Listener,Manager
package listeners

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/hr3lxphr6j/bililive-go/src/configs"
	"github.com/hr3lxphr6j/bililive-go/src/instance"
	"github.com/hr3lxphr6j/bililive-go/src/interfaces"
	"github.com/hr3lxphr6j/bililive-go/src/live"
	"github.com/hr3lxphr6j/bililive-go/src/pkg/events"
)

const (
	begin uint32 = iota
	pending
	running
	stopped
)

type Listener interface {
	Start() error
	Close()
}

func NewListener(ctx context.Context, live live.Live) Listener {
	inst := instance.GetInstance(ctx)
	return &listener{
		Live:   live,
		liveRoomName:  "",
		status: false,
		config: inst.Config,
		stop:   make(chan struct{}),
		ed:     inst.EventDispatcher.(events.Dispatcher),
		logger: inst.Logger,
		state:  begin,
	}
}

type listener struct {
	Live   live.Live
	liveRoomName string
	status bool

	config *configs.Config
	ed     events.Dispatcher
	logger *interfaces.Logger

	state uint32
	stop  chan struct{}
}

func (l *listener) Start() error {
	if !atomic.CompareAndSwapUint32(&l.state, begin, pending) {
		return nil
	}
	defer atomic.CompareAndSwapUint32(&l.state, pending, running)

	l.ed.DispatchEvent(events.NewEvent(ListenStart, l.Live))
	l.refresh()
	go l.run()
	return nil
}

func (l *listener) Close() {
	if !atomic.CompareAndSwapUint32(&l.state, running, stopped) {
		return
	}
	l.ed.DispatchEvent(events.NewEvent(ListenStop, l.Live))
	close(l.stop)
}

func (l *listener) refresh() {
	info, err := l.Live.GetInfo()
	if err != nil {
		l.logger.
			WithError(err).
			WithField("url", l.Live.GetRawUrl()).
			Error("failed to load room info")
		return
	}

	// if live is still inactive, return 
	if ! info.Status && ! l.status {
		return
	}
// 	l.status = info.Status // Put this line later

	var (
		evtTyp  events.EventType
		logInfo string
		fields  = map[string]interface{}{
			"room": info.RoomName,
			"host": info.HostName,
		}
	)
	if info.Status && ! l.status {
		l.Live.SetLastStartTime(time.Now())
		evtTyp = LiveStart
		logInfo = "Live Start"
	} else if info.Status && l.status {
		// When RoomName is changed, restart the live
		if l.config.Feature.RestartRenamedRooms && info.RoomName != l.liveRoomName { 
			// In order to restart live, send an LiveEnd event, and then send an LiveStart event
			l.ed.DispatchEvent(events.NewEvent(LiveEnd, l.Live))
			evtTyp = LiveStart
			fields["reason"] = "Room name is changed"
			logInfo = "Live Restarted"
		} else {
			return
		}
	} else if ! info.Status && l.status {
		evtTyp = LiveEnd
		logInfo = "Live end"
	}
	
	// Update listener's status & liveRoomName
	l.status = info.Status
	l.liveRoomName = info.RoomName
	
	l.ed.DispatchEvent(events.NewEvent(evtTyp, l.Live))
	l.logger.WithFields(fields).Info(logInfo)
}

func (l *listener) run() {
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
