package cc

import (
	"fmt"
	"net/url"

	"github.com/tidwall/gjson"

	"github.com/hr3lxphr6j/bililive-go/src/live"
	"github.com/hr3lxphr6j/bililive-go/src/live/internal"
	"github.com/hr3lxphr6j/bililive-go/src/pkg/http"
	"github.com/hr3lxphr6j/bililive-go/src/pkg/utils"
)

const (
	domain = "cc.163.com"
	cnName = "CC直播"

	apiUrl = "http://cgi.v.cc.163.com/video_play_url/"
	dataRe = `<script id="__NEXT_DATA__" type="application/json" crossorigin="anonymous">(.*?)</script>`
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
	dom, err := http.Get(l.Url.String(), nil, nil)
	if err != nil {
		return nil, err
	}
	data := utils.UnescapeHTMLEntity(utils.Match1(dataRe, string(dom)))
	if data == "" {
		return nil, live.ErrInternalError
	}
	result := gjson.Parse(data)
	return &result, nil
}

func (l *Live) getCcID() (string, error) {
	data, err := l.getData()
	if err != nil {
		return "", err
	}
	return data.Get("props.pageProps.roomInfoInitData.micfirst.ccid").String(), nil
}

func (l *Live) GetInfo() (info *live.Info, err error) {
	data, err := l.getData()
	if err != nil {
		return nil, err
	}
	var (
		hostName = data.Get("props.pageProps.roomInfoInitData.micfirst.nickname").String()
		roomName = data.Get("props.pageProps.roomInfoInitData.live.title").String()
	)

	if hostName == "" || roomName == "" {
		return nil, live.ErrInternalError
	}

	info = &live.Info{
		Live:     l,
		HostName: hostName,
		RoomName: roomName,
		Status:   data.Get("props.pageProps.roomInfoInitData.live.ccid").Int() != 0,
	}
	return info, nil
}

func (l *Live) GetStreamUrls() (us []*url.URL, err error) {
	ccid, err := l.getCcID()
	if err != nil {
		return nil, err
	}
	data, err := http.Get(fmt.Sprintf("%s%s", apiUrl, ccid), nil, nil)
	if err != nil {
		return nil, err
	}
	return utils.GenUrls(
		gjson.GetBytes(data, "videourl").String(),
		gjson.GetBytes(data, "bakvideourl").String(),
	)
}

func (l *Live) GetPlatformCNName() string {
	return cnName
}
