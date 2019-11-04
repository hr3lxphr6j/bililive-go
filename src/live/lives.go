//go:generate mockgen -package mock -destination mock/mock.go github.com/hr3lxphr6j/bililive-go/src/live Live
package live

import (
	"errors"
	"net/url"
	"time"

	"github.com/bluele/gcache"
)

var (
	m = make(map[string]Builder)
)

func Register(domain string, b Builder) {
	m[domain] = b
}

func getBuilder(domain string) (Builder, bool) {
	builder, ok := m[domain]
	return builder, ok
}

type Builder interface {
	Build(*url.URL) (Live, error)
}

type ID string

type Live interface {
	GetLiveId() ID
	GetRawUrl() string
	GetInfo() (*Info, error)
	GetStreamUrls() ([]*url.URL, error)
	GetPlatformCNName() string
	GetLastStartTime() time.Time
	SetLastStartTime(time.Time)
}

type wrappedLive struct {
	Live
	cache gcache.Cache
}

func newWrappedLive(live Live, cache gcache.Cache) Live {
	return &wrappedLive{
		Live:  live,
		cache: cache,
	}
}

func (w *wrappedLive) GetInfo() (*Info, error) {
	i, err := w.Live.GetInfo()
	if err != nil {
		return nil, err
	}
	if w.cache != nil {
		w.cache.Set(w, i)
	}
	return i, nil
}

func New(url *url.URL, cache gcache.Cache) (live Live, err error) {
	builder, ok := getBuilder(url.Host)
	if !ok {
		return nil, errors.New("not support this url")
	}
	live, err = builder.Build(url)
	if err != nil {
		return
	}
	live = newWrappedLive(live, cache)
	for i := 0; i < 3; i++ {
		if _, err = live.GetInfo(); err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	return
}
