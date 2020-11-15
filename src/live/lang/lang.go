package lang

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strings"

	"github.com/hr3lxphr6j/requests"
	"github.com/tidwall/gjson"

	"github.com/hr3lxphr6j/bililive-go/src/live"
	"github.com/hr3lxphr6j/bililive-go/src/live/internal"
	"github.com/hr3lxphr6j/bililive-go/src/pkg/utils"
)

const (
	playLiveDomain = "play.lang.live"
	liveDomain     = "www.lang.live"
	cnName         = "浪live"

	playLiveInfoAPIUrl = "https://game-api.lang.live/webapi/v1/room/info"
)

var liveInfoAPIUrls = [...]string{
	"https://langapi.lv-show.com/langweb/v1/room/liveinfo",
	"https://api.lang.live/langweb/v1/room/liveinfo",
}

func init() {
	live.Register(playLiveDomain, new(builder))
	live.Register(liveDomain, new(builder))
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
	var (
		roomID string
		api    string
		paths  = strings.Split(l.Url.Path, "/")
	)
	switch l.Url.Host {
	case liveDomain:
		if len(paths) < 3 {
			return nil, live.ErrRoomUrlIncorrect
		}
		roomID = paths[2]
		// TODO: Request all APIs at the same time, use the fastest return.
		api = liveInfoAPIUrls[rand.Int()&1]
	case playLiveDomain:
		if len(paths) < 2 {
			return nil, live.ErrRoomUrlIncorrect
		}
		roomID = paths[1]
		api = playLiveInfoAPIUrl
	}

	resp, err := requests.Get(api, live.CommonUserAgent, requests.Query("room_id", roomID))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, live.ErrRoomNotExist
	}
	body, err := resp.Bytes()
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

	var (
		hostNamePath = "live_info.nickname"
		roomNamePath string
		statusPath   = "live_info.live_status"
	)
	switch l.Url.Host {
	case liveDomain:
		roomNamePath = "live_info.pretty_id"
	case playLiveDomain:
		roomNamePath = "live_info.room_title"
	}

	return &live.Info{
		Live:     l,
		HostName: data.Get(hostNamePath).String(),
		RoomName: data.Get(roomNamePath).String(),
		Status:   data.Get(statusPath).Int() == 1,
	}, nil
}

func (l *Live) GetStreamUrls() (us []*url.URL, err error) {
	data, err := l.getData()
	if err != nil {
		return nil, err
	}
	urls := make([]string, 0)
	switch l.Url.Host {
	case playLiveDomain:
		streamID := data.Get("live_info.stream_id").Int()
		if u := data.Get(fmt.Sprintf("live_info.stream_items.#(id==%d).video", streamID)).String(); u != "" {
			urls = append(urls, u)
		}
		if u := data.Get(fmt.Sprintf("live_info.hls_items.#(id==%d).video", streamID)).String(); u != "" {
			urls = append(urls, u)
		}

	case liveDomain:
		if u := data.Get("live_info.liveurl").String(); u != "" {
			urls = append(urls, u)
		}
		if u := data.Get("live_info.liveurl_hls").String(); u != "" {
			urls = append(urls, u)
		}
	}
	return utils.GenUrls(urls...)
}

func (l *Live) GetPlatformCNName() string {
	return cnName
}
