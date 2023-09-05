package douyin

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/hr3lxphr6j/requests"
	"github.com/tidwall/gjson"

	"github.com/hr3lxphr6j/bililive-go/src/live"
	"github.com/hr3lxphr6j/bililive-go/src/live/internal"
	"github.com/hr3lxphr6j/bililive-go/src/pkg/utils"
)

const (
	domain = "live.douyin.com"
	cnName = "抖音"

	randomCookieChars  = "1234567890abcdef"
	roomIdCatcherRegex = `{\\"webrid\\":\\"([^"]+)\\"}`
)

var roomInfoApiForSprintf = "https://live.douyin.com/webcast/room/web/enter/?aid=6383&app_name=douyin_web&live_id=1&device_platform=web&language=zh-CN&browser_language=zh-CN&browser_platform=Win32&browser_name=Chrome&browser_version=116.0.0.0&web_rid=%s"

func init() {
	live.Register(domain, new(builder))
}

type builder struct{}

func (b *builder) Build(url *url.URL, opt ...live.Option) (live.Live, error) {
	return &Live{
		BaseLive:        internal.NewBaseLive(url, opt...),
		responseCookies: make(map[string]string),
	}, nil
}

func createRandomCookie() string {
	return utils.GenRandomString(21, randomCookieChars)
}

type Live struct {
	internal.BaseLive
	responseCookies map[string]string
}

func (l *Live) getRoomId() (string, error) {
	cookies := l.Options.Cookies.Cookies(l.Url)
	cookieKVs := make(map[string]string)
	cookieKVs["__ac_nonce"] = createRandomCookie()
	for _, item := range cookies {
		cookieKVs[item.Name] = item.Value
	}
	resp, err := requests.Get(
		l.Url.String(),
		live.CommonUserAgent,
		requests.Cookies(cookieKVs),
	)
	if err != nil {
		return "", err
	}
	switch code := resp.StatusCode; code {
	case http.StatusOK:
	default:
		return "", fmt.Errorf("failed to get page, code: %v, %w", code, live.ErrInternalError)
	}
	body, err := resp.Text()
	if err != nil {
		return "", err
	}
	roomId := utils.Match1(roomIdCatcherRegex, body)
	if roomId == "" {
		fmt.Println(body)
		return "", fmt.Errorf("failed to get RoomId from page, %w", live.ErrInternalError)
	}
	for _, cookie := range resp.Cookies() {
		l.responseCookies[cookie.Name] = cookie.Value
	}
	return roomId, nil
}

func (l *Live) getRoomInfo() (*gjson.Result, error) {
	roomId, err := l.getRoomId()
	if err != nil {
		return nil, err
	}
	cookies := l.Options.Cookies.Cookies(l.Url)
	cookieKVs := make(map[string]string)
	cookieKVs["__ac_nonce"] = createRandomCookie()
	for _, item := range cookies {
		cookieKVs[item.Name] = item.Value
	}
	for key, value := range l.responseCookies {
		cookieKVs[key] = value
	}
	roomInfoApi := fmt.Sprintf(roomInfoApiForSprintf, roomId)
	resp, err := requests.Get(
		roomInfoApi,
		live.CommonUserAgent,
		requests.Cookies(cookieKVs),
		requests.Headers(map[string]interface{}{
			"Cache-Control": "no-cache",
		}))
	if err != nil {
		return nil, err
	}
	switch code := resp.StatusCode; code {
	case http.StatusOK:
	case http.StatusNotFound:
		return nil, live.ErrRoomNotExist
	default:
		return nil, fmt.Errorf("failed to get page, code: %v, %w", code, live.ErrInternalError)
	}

	body, err := resp.Text()
	if err != nil {
		return nil, err
	}
	result := gjson.Parse(body)
	return &result, nil
}

func (l *Live) GetInfo() (info *live.Info, err error) {
	data, err := l.getRoomInfo()
	// data, err := l.getData()
	if err != nil {
		return nil, err
	}
	info = &live.Info{
		Live:     l,
		HostName: data.Get("data.user.nickname").String(),
		RoomName: data.Get("data.data.0.title").String(),
		Status:   data.Get("data.data.0.status").Int() == 2,
	}
	return
}

func (l *Live) GetStreamUrls() (us []*url.URL, err error) {
	data, err := l.getRoomInfo()
	if err != nil {
		return nil, err
	}
	var urls []string
	data.Get("data.data.0.stream_url.flv_pull_url").ForEach(func(key, value gjson.Result) bool {
		urls = append(urls, value.String())
		return true
	})
	streamData := gjson.Parse(data.Get("data.data.0.stream_url.live_core_sdk_data.pull_data.stream_data").String())
	if streamData.Exists() {
		url := streamData.Get("data.origin.main.flv")
		if url.Exists() {
			urls = append([]string{url.String()}, urls...)
		}
	}
	return utils.GenUrls(urls...)
}

func (l *Live) GetPlatformCNName() string {
	return cnName
}
