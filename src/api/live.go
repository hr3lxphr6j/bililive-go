package api

import (
	"net/url"
)

type Info struct {
	Url                *url.URL
	HostName, RoomName string
	Status             bool
}

type Live interface {
	GetRoom() (*Info, error)
	GetUrls() ([]*url.URL, error)
}

type RoomNotExistsError struct {
	Url *url.URL
}

func (e *RoomNotExistsError) Error() string {
	return "room not exists"
}

func IsRoomNotExistsError(err error) bool {
	_, ok := err.(*RoomNotExistsError)
	return ok
}

func NewLive(url *url.URL) Live {
	switch url.Host {
	case "www.panda.tv":
		return &PandaLive{url}
	case "live.bilibili.com":
		return &BiliBiliLive{url}
	case "www.zhanqi.tv":
		return &ZhanQiLive{url}
	default:
		return nil
	}
}
