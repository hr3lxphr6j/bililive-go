package parser

import (
	"errors"
	"net/url"
)

type Builder interface {
	Build() (Parser, error)
}

type Parser interface {
	ParseLiveStream(url *url.URL, file string) error
	Stop() error
}

var m = make(map[string]Builder)

func Register(name string, b Builder) {
	m[name] = b
}

func New(name string) (Parser, error) {
	builder, ok := m[name]
	if !ok {
		return nil, errors.New("unknown parser")
	}
	return builder.Build()
}
