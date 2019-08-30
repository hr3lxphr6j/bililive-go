//go:generate mockgen -package mock -destination mock/mock.go github.com/hr3lxphr6j/bililive-go/src/live Live
package live

import (
	"errors"
	"net/url"
	"time"
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

func New(url *url.URL) (live Live, err error) {
	builder, ok := getBuilder(url.Host)
	if !ok {
		return nil, errors.New("not support this url")
	}
	live, err = builder.Build(url)
	if err != nil {
		return
	}
	for i := 0; i < 3; i++ {
		if _, err = live.GetInfo(); err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	return
}
