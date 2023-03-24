package bilibili

import (
	"fmt"
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
	liveApiUrlv2    = "https://api.live.bilibili.com/xlive/web-room/v2/index/getRoomPlayInfo"
)

func init() {
	live.Register(domain, new(builder))
}

type builder struct{}

func (b *builder) Build(url *url.URL, opt ...live.Option) (live.Live, error) {
	return &Live{
		BaseLive: internal.NewBaseLive(url, opt...),
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
	cookies := l.Options.Cookies.Cookies(l.Url)
	cookieKVs := make(map[string]string)
	for _, item := range cookies {
		cookieKVs[item.Name] = item.Value
	}
	resp, err := requests.Get(roomInitUrl, live.CommonUserAgent, requests.Query("id", paths[1]), requests.Cookies(cookieKVs))
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
	cookies := l.Options.Cookies.Cookies(l.Url)
	cookieKVs := make(map[string]string)
	for _, item := range cookies {
		cookieKVs[item.Name] = item.Value
	}
	resp, err := requests.Get(
		roomApiUrl,
		live.CommonUserAgent,
		requests.Query("room_id", l.realID),
		requests.Query("from", "room"),
		requests.Cookies(cookieKVs),
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
	cookies := l.Options.Cookies.Cookies(l.Url)
	cookieKVs := make(map[string]string)
	for _, item := range cookies {
		cookieKVs[item.Name] = item.Value
	}
	// format: 0 for http_stream(flv), 1 for http_tls(m3u8), 2 for unknown, default 0,1,2
	// codec: 0 for avc(原画), 1 for hevc(原画Pro)
	query := fmt.Sprintf("?room_id=%s&protocol=0,1&format=0,1,2&codec=0,1&qn=10000&platform=web&ptype=8&dolby=5&panorama=1", l.realID)
	resp, err := requests.Get(liveApiUrlv2+query, live.CommonUserAgent, requests.Cookies(cookieKVs))

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

	addr := ""

	if l.Options.Quality == 0 && gjson.GetBytes(body, "data.playurl_info.playurl.stream.1.format.1.codec.#").Int() > 1 {
		addr = "data.playurl_info.playurl.stream.1.format.1.codec.1" // hevc m3u8
	} else {
		addr = "data.playurl_info.playurl.stream.0.format.0.codec.0" // avc flv
	}

	baseURL := gjson.GetBytes(body, addr+".base_url").String()
	gjson.GetBytes(body, addr+".url_info").ForEach(func(_, value gjson.Result) bool {
		hosts := gjson.Get(value.String(), "host").String()
		queries := gjson.Get(value.String(), "extra").String()
		urls = append(urls, hosts+baseURL+queries)
		return true
	})

	return utils.GenUrls(urls...)
}

func (l *Live) GetPlatformCNName() string {
	return cnName
}
