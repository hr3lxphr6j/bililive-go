package lang

import (
	"net/url"
	"strings"

	"github.com/tidwall/gjson"

	"github.com/hr3lxphr6j/bililive-go/src/live"
	"github.com/hr3lxphr6j/bililive-go/src/live/internal"
	"github.com/hr3lxphr6j/bililive-go/src/pkg/http"
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
	realID string
}

// 2132991
func (l *Live) GetInfo() (info *live.Info, err error) {
	paths := strings.Split(l.Url.Path, "/")
	if len(paths) < 2 {
		return nil, live.ErrRoomUrlIncorrect
	}
	body, err := http.Get(liveInfoAPIUrl, nil, map[string]string{
		"keyword": paths[1],
	})
	if err != nil || gjson.GetBytes(body, "ret_code").Int() != 0 {
		return nil, live.ErrRoomNotExist
	}

	roomData := gjson.GetBytes(body, "data.users.#(room_id==%s)")
	if !roomData.Exists() {
		return nil, live.ErrRoomNotExist
	}

	return &live.Info{
		Live:     l,
		HostName: roomData.Get("nickname").String(),
		RoomName: roomData.Get("").String(),
		Status:   roomData.Get("").Int() == 1,
	}, nil
}

func (l *Live) GetStreamUrls() (us []*url.URL, err error) {
	// TODO: Implement this method
	return nil, nil
}

func (l *Live) GetPlatformCNName() string {
	return cnName
}
