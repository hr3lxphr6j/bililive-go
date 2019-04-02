package parser

import "net/url"

type Parser interface {
	ParseLiveStream(url *url.URL, file string) error
	Stop() error
}
