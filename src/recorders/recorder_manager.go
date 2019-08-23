package recorders

import (
	"context"
	"sync"

	"github.com/hr3lxphr6j/bililive-go/src/instance"
	"github.com/hr3lxphr6j/bililive-go/src/lib/events"
	"github.com/hr3lxphr6j/bililive-go/src/listeners"
	"github.com/hr3lxphr6j/bililive-go/src/live"
)

func NewIRecorderManager(ctx context.Context) IRecorderManager {
	rm := &RecorderManager{
		savers: make(map[live.ID]*Recorder),
		lock:   new(sync.RWMutex),
	}
	instance.GetInstance(ctx).RecorderManager = rm
	return rm
}

type IRecorderManager interface {
	AddRecorder(ctx context.Context, live live.Live) error
	RemoveRecorder(ctx context.Context, liveId live.ID) error
	GetRecorder(ctx context.Context, liveId live.ID) (*Recorder, error)
	HasRecorder(ctx context.Context, liveId live.ID) bool
}

type RecorderManager struct {
	savers map[live.ID]*Recorder
	lock   *sync.RWMutex
}

func (r *RecorderManager) Start(ctx context.Context) error {
	inst := instance.GetInstance(ctx)
	if inst.Config.RPC.Enable || len(inst.Lives) > 0 {
		inst.WaitGroup.Add(1)
	}
	ed := inst.EventDispatcher.(events.IEventDispatcher)

	// 开播事件
	ed.AddEventListener(listeners.LiveStart, events.NewEventListener(func(event *events.Event) {
		live := event.Object.(live.Live)
		if err := r.AddRecorder(ctx, live); err != nil {
			instance.GetInstance(ctx).Logger.
				Errorf("failed to add recorder, err: %v", err)
		}
	}))

	// 下播事件
	ed.AddEventListener(listeners.LiveEnd, events.NewEventListener(func(event *events.Event) {
		live := event.Object.(live.Live)
		if !r.HasRecorder(ctx, live.GetLiveId()) {
			return
		}
		if err := r.RemoveRecorder(ctx, live.GetLiveId()); err != nil {
			instance.GetInstance(ctx).Logger.
				Errorf("failed to remove recorder, err: %v", err)
		}
	}))

	// 监听关闭事件
	ed.AddEventListener(listeners.ListenStop, events.NewEventListener(func(event *events.Event) {
		live := event.Object.(live.Live)
		if !r.HasRecorder(ctx, live.GetLiveId()) {
			return
		}
		if err := r.RemoveRecorder(ctx, live.GetLiveId()); err != nil {
			instance.GetInstance(ctx).Logger.
				Errorf("failed to remove recorder, err: %v", err)
		}
	}))

	return nil
}

func (r *RecorderManager) Close(ctx context.Context) {
	r.lock.Lock()
	defer r.lock.Unlock()
	for id, recorder := range r.savers {
		recorder.Close()
		delete(r.savers, id)
	}
	inst := instance.GetInstance(ctx)
	inst.WaitGroup.Done()
}

func (r *RecorderManager) AddRecorder(ctx context.Context, live live.Live) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	if _, ok := r.savers[live.GetLiveId()]; ok {
		return recorderExistError
	}
	recorder, err := NewRecorder(ctx, live)
	if err != nil {
		return err
	}
	r.savers[live.GetLiveId()] = recorder
	recorder.Start()
	return nil

}

func (r *RecorderManager) RemoveRecorder(ctx context.Context, liveId live.ID) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	if recorder, ok := r.savers[liveId]; !ok {
		return recorderNotExistError
	} else {
		recorder.Close()
		delete(r.savers, liveId)
		return nil
	}
}

func (r *RecorderManager) GetRecorder(ctx context.Context, liveId live.ID) (*Recorder, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	if r, ok := r.savers[liveId]; !ok {
		return nil, recorderNotExistError
	} else {
		return r, nil
	}
}

func (r *RecorderManager) HasRecorder(ctx context.Context, liveId live.ID) bool {
	r.lock.RLock()
	defer r.lock.RUnlock()
	_, ok := r.savers[liveId]
	return ok
}
