package recorders

import (
	"context"
	"github.com/hr3lxphr6j/bililive-go/src/api"
	"github.com/hr3lxphr6j/bililive-go/src/instance"
	"github.com/hr3lxphr6j/bililive-go/src/lib/events"
	"github.com/hr3lxphr6j/bililive-go/src/listeners"
	"sync"
	"time"
)

func NewIRecorderManager(ctx context.Context) IRecorderManager {
	rm := &RecorderManager{
		saver: make(map[api.Live]*Recorder),
		lock:  new(sync.RWMutex),
	}
	instance.GetInstance(ctx).RecorderManager = rm
	return rm
}

type IRecorderManager interface {
	AddRecorder(ctx context.Context, live *api.Info) error
	GetRecorder(ctx context.Context, live api.Live) (*Recorder, error)
	RemoveRecorder(ctx context.Context, live api.Live) (time.Duration, error)
}

type RecorderManager struct {
	saver map[api.Live]*Recorder
	lock  *sync.RWMutex
}

func (r *RecorderManager) Start(ctx context.Context) error {
	inst := instance.GetInstance(ctx)
	inst.WaitGroup.Add(1)

	ed := inst.EventDispatcher.(events.IEventDispatcher)

	// 开播事件
	ed.AddEventListener(listeners.LiveStart, events.NewEventListener(func(event *events.Event) {
		r.AddRecorder(ctx, event.Object.(*api.Info))
	}))

	// 下播事件
	ed.AddEventListener(listeners.LiveEnd, events.NewEventListener(func(event *events.Event) {
		r.RemoveRecorder(ctx, event.Object.(*api.Info).Live)
	}))

	return nil
}

func (r *RecorderManager) Close(ctx context.Context) {

}

func (r *RecorderManager) AddRecorder(ctx context.Context, info *api.Info) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	if _, ok := r.saver[info.Live]; ok {
		return recorderExistError
	}
	recorder, err := NewRecorder(ctx, info)
	if err != nil {
		return err
	}
	r.saver[info.Live] = recorder
	recorder.Start()
	return nil

}

func (r *RecorderManager) GetRecorder(ctx context.Context, live api.Live) (*Recorder, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	if r, ok := r.saver[live]; !ok {
		return nil, recorderNotExistError
	} else {
		return r, nil
	}
}

func (r *RecorderManager) RemoveRecorder(ctx context.Context, live api.Live) (time.Duration, error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	if recorder, ok := r.saver[live]; !ok {
		return 0, recorderNotExistError
	} else {
		recorder.Close()
		delete(r.saver, live)
		return time.Now().Sub(recorder.StartTime), nil
	}
}
