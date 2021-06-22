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

	regRenderData = `<script id="RENDER_DATA" type="application/json">(.*?)</script>`
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
	resp, err := requests.Get(l.Url.String(), live.CommonUserAgent)
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
	rawData := utils.Match1(regRenderData, body)
	if rawData == "" {
		return nil, fmt.Errorf("failed to get RENDER_DATA from page, %w", live.ErrInternalError)
	}
	unescapedRawData, err := url.QueryUnescape(rawData)
	if err != nil {
		return nil, err
	}
	result := gjson.Parse(unescapedRawData)
	return &result, nil
}

func (l *Live) GetInfo() (info *live.Info, err error) {
	data, err := l.getData()
	if err != nil {
		return nil, err
	}
	info = &live.Info{
		Live:     l,
		HostName: data.Get("routeInitialProps.roomInfo.room.owner.nickname").String(),
		RoomName: data.Get("routeInitialProps.roomInfo.anchor.nickname").String(),
		Status:   data.Get("routeInitialProps.roomInfo.room.status").Int() == 2,
	}
	return
}

func (l *Live) GetStreamUrls() (us []*url.URL, err error) {
	data, err := l.getData()
	if err != nil {
		return nil, err
	}
	var urls []string
	data.Get("routeInitialProps.roomInfo.room.stream_url.flv_pull_url").ForEach(func(key, value gjson.Result) bool {
		urls = append(urls, value.String())
		return true
	})
	return utils.GenUrls(urls...)
}

func (l *Live) GetPlatformCNName() string {
	return cnName
}
