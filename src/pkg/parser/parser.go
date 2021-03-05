//go:generate mockgen -package mock -destination mock/mock.go github.com/hr3lxphr6j/bililive-go/src/pkg/parser Parser
package parser

import (
	"errors"
	"net/url"

	"github.com/hr3lxphr6j/bililive-go/src/live"
)

type Builder interface {
	Build(cfg map[string]string) (Parser, error)
}

type Parser interface {
	ParseLiveStream(url *url.URL, live live.Live, file string) error
	Stop() error
}

type StatusParser interface {
	Parser
	Status() (map[string]string, error)
}

var m = make(map[string]Builder)

func Register(name string, b Builder) {
	m[name] = b
}

func New(name string, cfg map[string]string) (Parser, error) {
	builder, ok := m[name]
	if !ok {
		return nil, errors.New("unknown parser")
	}
	return builder.Build(cfg)
}
