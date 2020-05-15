package lang

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/tidwall/gjson"

	"github.com/hr3lxphr6j/bililive-go/src/live"
	"github.com/hr3lxphr6j/bililive-go/src/live/internal"
	"github.com/hr3lxphr6j/bililive-go/src/pkg/http"
	"github.com/hr3lxphr6j/bililive-go/src/pkg/utils"
)

const (
	domain = "play.lang.live"
	cnName = "浪live"

	liveInfoAPIUrl = "https://api.kingkongapp.com/webapi/v1/room/info"
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

func (l *Live) getData() (*gjson.Result, error) {
	paths := strings.Split(l.Url.Path, "/")
	if len(paths) < 2 {
		return nil, live.ErrRoomUrlIncorrect
	}
	body, err := http.Get(liveInfoAPIUrl, nil, map[string]string{
		"room_id": paths[1],
	})
	if err != nil || gjson.GetBytes(body, "ret_code").Int() != 0 {
		return nil, live.ErrRoomNotExist
	}
	data := gjson.GetBytes(body, "data")
	return &data, nil
}

func (l *Live) GetInfo() (info *live.Info, err error) {
	data, err := l.getData()
	if err != nil {
		return nil, err
	}

	return &live.Info{
		Live:     l,
		HostName: data.Get("live_info.nickname").String(),
		RoomName: data.Get("live_info.room_title").String(),
		Status:   data.Get("live_info.live_status").Int() == 1,
	}, nil
}

func (l *Live) GetStreamUrls() (us []*url.URL, err error) {
	data, err := l.getData()
	if err != nil {
		return nil, err
	}
	return utils.GenUrls(data.Get(fmt.Sprintf("live_info.stream_items.#(id==%d).video", data.Get("live_info.stream_id").Int())).String())
}

func (l *Live) GetPlatformCNName() string {
	return cnName
}
