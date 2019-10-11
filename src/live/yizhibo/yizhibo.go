package yizhibo

import (
	"net/url"
	"strings"

	"github.com/tidwall/gjson"

	"github.com/hr3lxphr6j/bililive-go/src/lib/http"
	"github.com/hr3lxphr6j/bililive-go/src/lib/utils"
	"github.com/hr3lxphr6j/bililive-go/src/live"
	"github.com/hr3lxphr6j/bililive-go/src/live/internal"
)

const (
	domain = "www.yizhibo.com"
	cnName = "一直播"

	apiUrl = "http://www.yizhibo.com/live/h5api/get_basic_live_info"
)

func init() {
	live.Register(domain, new(builder))
}

type builder struct{}

func (b *builder) Build(url *url.URL) (live.Live, error) {
	return &Live{
		BaseLive: internal.NewBaseLive(url),
	}, nil
}

type Live struct {
	internal.BaseLive
}

func (l *Live) requestRoomInfo() ([]byte, error) {
	scid := strings.Split(strings.Split(l.Url.Path, "/")[2], ".")[0]
	body, err := http.Get(apiUrl, nil, map[string]string{"scid": scid})
	if err != nil {
		return nil, err
	}
	if gjson.GetBytes(body, "result").Int() != 1 {
		return nil, live.ErrRoomNotExist
	} else {
		return body, nil
	}
}

func (l *Live) GetInfo() (info *live.Info, err error) {
	data, err := l.requestRoomInfo()
	if err != nil {
		return nil, err
	}
	info = &live.Info{
		Live:     l,
		HostName: gjson.GetBytes(data, "data.nickname").String(),
		RoomName: gjson.GetBytes(data, "data.live_title").String(),
		Status:   gjson.GetBytes(data, "data.status").Int() == 10,
	}
	return info, nil
}

func (l *Live) GetStreamUrls() (us []*url.URL, err error) {
	data, err := l.requestRoomInfo()
	if err != nil {
		return nil, err
	}
	return utils.GenUrls(gjson.GetBytes(data, "data.play_url").String())
}

func (l *Live) GetPlatformCNName() string {
	return cnName
}
