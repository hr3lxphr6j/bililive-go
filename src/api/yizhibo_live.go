package api

import (
	"net/url"
	"strings"

	"github.com/tidwall/gjson"

	"github.com/hr3lxphr6j/bililive-go/src/lib/http"
)

const yizhiboApiUrl = "http://www.yizhibo.com/live/h5api/get_basic_live_info"

type YiZhiBoLive struct {
	abstractLive
}

func (y *YiZhiBoLive) requestRoomInfo() ([]byte, error) {
	scid := strings.Split(strings.Split(y.Url.Path, "/")[2], ".")[0]
	body, err := http.Get(yizhiboApiUrl, map[string]string{"scid": scid}, nil)
	if err != nil {
		return nil, err
	}
	if gjson.GetBytes(body, "result").Int() != 1 {
		return nil, &RoomNotExistsError{y.Url}
	} else {
		return body, nil
	}
}

func (y *YiZhiBoLive) GetInfo() (info *Info, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()

	data, err := y.requestRoomInfo()
	if err != nil {
		return nil, err
	}
	info = &Info{
		Live:     y,
		HostName: gjson.GetBytes(data, "data.nickname").String(),
		RoomName: gjson.GetBytes(data, "data.live_title").String(),
		Status:   gjson.GetBytes(data, "data.status").Int() == 10,
	}
	y.cachedInfo = info
	return info, nil
}

func (y *YiZhiBoLive) GetStreamUrls() (us []*url.URL, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()

	data, err := y.requestRoomInfo()
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(gjson.GetBytes(data, "data.play_url").String())
	if err != nil {
		return nil, err
	}
	return []*url.URL{u}, nil
}
