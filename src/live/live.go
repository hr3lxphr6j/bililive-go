package live

import (
	"net/url"
)

var commonHeader = map[string]string{
	"Accept":          "application/json, text/javascript, */*; q=0.01",
	"Accept-Encoding": "gzip, deflate",
	"Accept-Language": "zh-CN,zh;q=0.8,en-US;q=0.6,en;q=0.4,zh-TW;q=0.2",
	"Connection":      "keep-alive",
	"User-Agent":      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Safari/537.36",
}

type RoomNotExistsError struct {
	Url url.URL
}

func (e RoomNotExistsError) Error() string {
	return "room not exists"
}

func IsRoomNotExistsError(err error) bool {
	_, ok := err.(*RoomNotExistsError)
	return ok
}

type Info struct {
	url                string
	HostName, RoomName string
	Status             bool
}

type Live interface {
	GetRoom() (*Info, error)
	GetUrls() ([]url.URL, error)
}
