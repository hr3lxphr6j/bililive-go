package listeners

import (
	"context"
	"errors"
	"testing"

	"github.com/bluele/gcache"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/hr3lxphr6j/bililive-go/src/configs"
	"github.com/hr3lxphr6j/bililive-go/src/instance"
	"github.com/hr3lxphr6j/bililive-go/src/lib/events"
	evtmock "github.com/hr3lxphr6j/bililive-go/src/lib/events/mock"
	livepkg "github.com/hr3lxphr6j/bililive-go/src/live"
	livemock "github.com/hr3lxphr6j/bililive-go/src/live/mock"
	"github.com/hr3lxphr6j/bililive-go/src/log"
)

func TestRefresh(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ed := evtmock.NewMockDispatcher(ctrl)
	ctx := context.WithValue(context.Background(), instance.InstanceKey, &instance.Instance{
		EventDispatcher: ed,
		Config:          new(configs.Config),
	})
	log.New(ctx)
	live := livemock.NewMockLive(ctrl)
	l := NewListener(ctx, live).(*listener)

	// false -> false
	live.EXPECT().GetInfo().Return(&livepkg.Info{Status: false}, nil)
	l.refresh()
	assert.False(t, l.status)

	// false -> true
	live.EXPECT().GetInfo().Return(&livepkg.Info{Status: true}, nil)
	live.EXPECT().SetLastStartTime(gomock.Any())
	ed.EXPECT().DispatchEvent(events.NewEvent(LiveStart, live))
	l.refresh()
	assert.True(t, l.status)

	// true -> false
	live.EXPECT().GetInfo().Return(&livepkg.Info{Status: false}, nil)
	ed.EXPECT().DispatchEvent(events.NewEvent(LiveEnd, live))
	l.refresh()
	assert.False(t, l.status)
}

func TestRefreshWithError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ed := evtmock.NewMockDispatcher(ctrl)
	cache := gcache.New(4).LRU().Build()
	ctx := context.WithValue(context.Background(), instance.InstanceKey, &instance.Instance{
		EventDispatcher: ed,
		Cache:           cache,
		Config:          new(configs.Config),
	})
	log.New(ctx)
	live := livemock.NewMockLive(ctrl)
	l := NewListener(ctx, live).(*listener)

	live.EXPECT().GetInfo().Return(nil, errors.New("this is error"))
	live.EXPECT().GetRawUrl().Return("")
	l.refresh()
	assert.False(t, l.status)
}

func TestListenerStartAndClose(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ed := evtmock.NewMockDispatcher(ctrl)
	cache := gcache.New(4).LRU().Build()
	ctx := context.WithValue(context.Background(), instance.InstanceKey, &instance.Instance{
		EventDispatcher: ed,
		Cache:           cache,
		Config:          &configs.Config{Interval: 5},
	})
	log.New(ctx)
	live := livemock.NewMockLive(ctrl)
	live.EXPECT().GetInfo().Return(&livepkg.Info{Status: false}, nil)
	ed.EXPECT().DispatchEvent(gomock.Any()).Times(2)
	l := NewListener(ctx, live)
	assert.NoError(t, l.Start())
	assert.NoError(t, l.Start())
	l.Close()
	l.Close()
}
