package longzhu

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/tidwall/gjson"

	"github.com/hr3lxphr6j/bililive-go/src/lib/http"
	"github.com/hr3lxphr6j/bililive-go/src/lib/utils"
	"github.com/hr3lxphr6j/bililive-go/src/live"
	"github.com/hr3lxphr6j/bililive-go/src/live/internal"
)

const (
	domain = "star.longzhu.com"
	cnName = "龙珠"

	mobileUrl  = "http://m.longzhu.com/"
	roomApiUrl = "http://liveapi.plu.cn/liveapp/roomstatus"
	liveApiUrl = "http://livestream.plu.cn/live/getlivePlayurl"
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
	realId string
}

func (l *Live) parseRealId() error {
	paths := strings.Split(l.Url.Path, "/")
	if len(paths) < 2 {
		return live.ErrRoomUrlIncorrect
	}
	dom, err := http.Get(fmt.Sprintf("%s%s", mobileUrl, paths[1]), nil, nil)
	if err != nil {
		return err
	}
	realId := utils.Match1(`var\s*roomId\s*=\s*(\d+);`, string(dom))
	if realId == "" {
		return live.ErrRoomNotExist
	}
	l.realId = realId
	return nil
}

func (l *Live) GetInfo() (info *live.Info, err error) {
	if l.realId == "" {
		if err := l.parseRealId(); err != nil {
			return nil, err
		}
	}
	body, err := http.Get(roomApiUrl, nil, map[string]string{"roomId": l.realId})
	if err != nil {
		return nil, err
	}
	info = &live.Info{
		Live:     l,
		HostName: gjson.GetBytes(body, "userName").String(),
		RoomName: gjson.GetBytes(body, "title").String(),
		Status:   len(gjson.GetBytes(body, "streamUri").String()) > 4,
	}
	return info, nil

}

func (l *Live) GetStreamUrls() (us []*url.URL, err error) {
	if l.realId == "" {
		if err := l.parseRealId(); err != nil {
			return nil, err
		}
	}
	body, err := http.Get(liveApiUrl, nil, map[string]string{"roomId": l.realId})
	if err != nil {
		return nil, err
	}
	urls := make([]string, 0, 0)
	gjson.GetBytes(body, "playLines.0.urls.#.securityUrl").ForEach(func(key, value gjson.Result) bool {
		urls = append(urls, value.String())
		return true
	})
	return utils.GenUrls(urls...)
}

func (l *Live) GetPlatformCNName() string {
	return cnName
}
