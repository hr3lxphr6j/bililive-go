package recorders

import (
	"context"
	"github.com/bluele/gcache"
	"net/url"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/hr3lxphr6j/bililive-go/src/configs"
	"github.com/hr3lxphr6j/bililive-go/src/instance"
	"github.com/hr3lxphr6j/bililive-go/src/live"
	livemock "github.com/hr3lxphr6j/bililive-go/src/live/mock"
)

func TestManagerAddAndRemoveRecorder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.WithValue(context.Background(), instance.Key, &instance.Instance{
		Config: new(configs.Config),
	})
	m := NewManager(ctx)
	backup := newRecorder
	newRecorder = func(ctx context.Context, live live.Live) (Recorder, error) {
		r := NewMockRecorder(ctrl)
		r.EXPECT().Start(ctx).Return(nil)
		r.EXPECT().Close()
		return r, nil
	}
	defer func() { newRecorder = backup }()
	l := livemock.NewMockLive(ctrl)
	l.EXPECT().GetLiveId().Return(live.ID("test")).AnyTimes()
	assert.NoError(t, m.AddRecorder(context.Background(), l))
	assert.Equal(t, ErrRecorderExist, m.AddRecorder(context.Background(), l))
	ln, err := m.GetRecorder(context.Background(), "test")
	assert.NoError(t, err)
	assert.NotNil(t, ln)
	assert.True(t, m.HasRecorder(context.Background(), "test"))
	assert.NoError(t, m.RestartRecorder(context.Background(), l))
	assert.NoError(t, m.RemoveRecorder(context.Background(), "test"))
	assert.Equal(t, ErrRecorderNotExist, m.RemoveRecorder(context.Background(), "test"))
	_, err = m.GetRecorder(context.Background(), "test")
	assert.Equal(t, ErrRecorderNotExist, err)
	assert.False(t, m.HasRecorder(context.Background(), "test"))
}

func TestManager_Time(t *testing.T) {

	ot, _ := time.Parse("2006-01-02 15:04:05", "2023-07-15 23:01:00")

	tt := ot.Truncate(time.Hour)

	thisHour := ot.Truncate(time.Hour)
	nextHour := thisHour.Add(time.Hour)

	println(nextHour.Sub(ot).String())

	println("tets" + tt.String())
}

func TestManager_AddRecorder(t *testing.T) {
	mgr := manager{}
	ctx := context.Background()
	_url, _ := url.Parse("https://www.douyu.com/10976478")
	live, _ := live.New(_url, gcache.New(1024).LRU().Build(), nil)
	mgr.AddRecorder(ctx, live)
}
