package qq

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
	domain    = "egame.qq.com"
	cnName    = "企鹅电竞"
	mobileUrl = "https://m.egame.qq.com/live"
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

func (l *Live) GetInfo() (info *live.Info, err error) {
	paths := strings.Split(l.Url.Path, "/")
	if len(paths) < 2 {
		return nil, live.ErrRoomUrlIncorrect
	}
	anchorID := paths[1]
	dom, err := http.Get(l.Url.String(), nil, nil)
	if err != nil {
		return nil, err
	}
	roomName := utils.UnescapeHTMLEntity(utils.Match1(`title:"([^"]*)"`, string(dom)))
	hostName := utils.UnescapeHTMLEntity(utils.Match1(`nickName:"([^"]+)"`, string(dom)))
	dom2, err := http.Get(mobileUrl, nil, map[string]string{
		"anchorid": anchorID,
	})
	if err != nil {
		return nil, err
	}
	isLive := utils.Match1(`"isLive":(\d+)`, string(dom2))
	if roomName == "" || hostName == "" || isLive == "" {
		return nil, live.ErrInternalError
	}
	info = &live.Info{
		Live:     l,
		RoomName: string(roomName),
		HostName: string(hostName),
		Status:   isLive == "1",
	}
	return info, nil
}

func (l *Live) GetStreamUrls() (us []*url.URL, err error) {
	dom, err := http.Get(l.Url.String(), nil, nil)
	if err != nil {
		return nil, err
	}
	result := utils.Match1(`"urlArray":(\[[^\]]+\])`, string(dom))
	if result == "" {
		return nil, live.ErrInternalError
	}
	return utils.GenUrls(gjson.Get(result, "#[bitrate==0].playUrl").String())
}

func (l *Live) GetPlatformCNName() string {
	return cnName
}
