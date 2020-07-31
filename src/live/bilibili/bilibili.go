package bilibili

import (
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
	domain = "live.bilibili.com"
	cnName = "哔哩哔哩"

	roomInitUrl = "https://api.live.bilibili.com/room/v1/Room/room_init"
	roomApiUrl  = "https://api.live.bilibili.com/room/v1/Room/get_info"
	userApiUrl  = "https://api.live.bilibili.com/live_user/v1/UserInfo/get_anchor_in_room"
	liveApiUrl  = "https://api.live.bilibili.com/room/v1/Room/playUrl"
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

func (l *Live) parseRealId() error {
	paths := strings.Split(l.Url.Path, "/")
	if len(paths) < 2 {
		return live.ErrRoomUrlIncorrect
	}
	resp, err := requests.Get(roomInitUrl, live.CommonUserAgent, requests.Query("id", paths[1]))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return live.ErrRoomNotExist
	}
	body, err := resp.Bytes()
	if err != nil || gjson.GetBytes(body, "code").Int() != 0 {
		return live.ErrRoomNotExist
	}
	l.realID = gjson.GetBytes(body, "data.room_id").String()
	return nil
}

func (l *Live) GetInfo() (info *live.Info, err error) {
	// Parse the short id from URL to full id
	if l.realID == "" {
		if err := l.parseRealId(); err != nil {
			return nil, err
		}
	}
	resp, err := requests.Get(
		roomApiUrl,
		live.CommonUserAgent,
		requests.Query("room_id", l.realID),
		requests.Query("from", "room"),
	)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, live.ErrRoomNotExist
	}
	body, err := resp.Bytes()
	if err != nil {
		return nil, err
	}
	if gjson.GetBytes(body, "code").Int() != 0 {
		return nil, live.ErrRoomNotExist
	}

	info = &live.Info{
		Live:     l,
		RoomName: gjson.GetBytes(body, "data.title").String(),
		Status:   gjson.GetBytes(body, "data.live_status").Int() == 1,
	}

	resp, err = requests.Get(userApiUrl, live.CommonUserAgent, requests.Query("roomid", l.realID))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, live.ErrInternalError
	}
	body, err = resp.Bytes()
	if err != nil {
		return nil, err
	}
	if gjson.GetBytes(body, "code").Int() != 0 {
		return nil, live.ErrInternalError
	}

	info.HostName = gjson.GetBytes(body, "data.info.uname").String()
	return info, nil
}

func (l *Live) GetStreamUrls() (us []*url.URL, err error) {
	if l.realID == "" {
		if err := l.parseRealId(); err != nil {
			return nil, err
		}
	}
	resp, err := requests.Get(liveApiUrl, live.CommonUserAgent, requests.Queries(map[string]string{
		"cid":      l.realID,
		"quality":  "4",
		"platform": "web",
	}))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, live.ErrRoomNotExist
	}
	body, err := resp.Bytes()
	if err != nil {
		return nil, err
	}
	urls := make([]string, 0, 4)
	gjson.GetBytes(body, "data.durl.#.url").ForEach(func(_, value gjson.Result) bool {
		urls = append(urls, value.String())
		return true
	})
	return utils.GenUrls(urls...)
}

func (l *Live) GetPlatformCNName() string {
	return cnName
}
