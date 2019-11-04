package cc

import (
	"fmt"
	"net/url"
	"regexp"

	"github.com/tidwall/gjson"

	"github.com/hr3lxphr6j/bililive-go/src/lib/http"
	"github.com/hr3lxphr6j/bililive-go/src/lib/utils"
	"github.com/hr3lxphr6j/bililive-go/src/live"
	"github.com/hr3lxphr6j/bililive-go/src/live/internal"
)

const (
	domain = "cc.163.com"
	cnName = "CC直播"

	apiUrl = "http://cgi.v.cc.163.com/video_play_url/"
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
	ccid string
}

func (l *Live) parseCCId() error {
	dom, err := http.Get(l.Url.String(), nil, nil)
	if err != nil {
		return err
	}
	ccid := utils.Match1(`anchorCcId\s*:\s*'(\d*)'`, string(dom))
	if ccid == "" {
		return live.ErrInternalError
	}
	l.ccid = ccid
	return nil
}

func (l *Live) GetInfo() (info *live.Info, err error) {
	dom, err := http.Get(l.Url.String(), nil, nil)
	if err != nil {
		return nil, err
	}

	var (
		hostName = utils.UnescapeHTMLEntity(utils.Match1(`anchorName\s*:\s*'([^']*)',`, string(dom)))
		roomName = utils.UnescapeHTMLEntity(utils.Match1(`js-live-title nick" title\s*=\s*"([^"]*)"`, string(dom)))
	)

	if hostName == "" || roomName == "" {
		return nil, live.ErrInternalError
	}

	info = &live.Info{
		Live:     l,
		HostName: hostName,
		RoomName: roomName,
		Status:   len(regexp.MustCompile(`isLive\s*:\s*\d+,`).FindAll(dom, -1)) > 0,
	}
	return info, nil
}

func (l *Live) GetStreamUrls() (us []*url.URL, err error) {
	if l.ccid == "" {
		if err := l.parseCCId(); err != nil {
			return nil, err
		}
	}
	data, err := http.Get(fmt.Sprintf("%s%s", apiUrl, l.ccid), nil, nil)
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
