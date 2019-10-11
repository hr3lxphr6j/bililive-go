package zhanqi

import (
	"encoding/base64"
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
	domain = "www.zhanqi.tv"
	cnName = "战旗"

	apiUrl = "https://www.zhanqi.tv/api/static/v2.1/room/domain/%s.json"
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

func (l *Live) requestRoomInfo() ([]byte, error) {
	body, err := http.Get(fmt.Sprintf(apiUrl, strings.Split(l.Url.Path, "/")[1]), nil, nil)
	if err != nil {
		return nil, err
	}
	if gjson.GetBytes(body, "code").Int() == 0 {
		return body, nil
	} else {
		return nil, live.ErrRoomNotExist
	}
}

func (l *Live) GetInfo() (info *live.Info, err error) {
	body, err := l.requestRoomInfo()
	if err != nil {
		return nil, err
	}
	info = &live.Info{
		Live:     l,
		HostName: gjson.GetBytes(body, "data.nickname").String(),
		RoomName: gjson.GetBytes(body, "data.title").String(),
		Status:   gjson.GetBytes(body, "data.status").String() == "4",
	}
	return info, nil
}

func (l *Live) GetStreamUrls() (us []*url.URL, err error) {
	body, err := l.requestRoomInfo()
	if err != nil {
		return nil, err
	}
	videoLevels := gjson.GetBytes(body, "data.flashvars.VideoLevels").String()
	data, err := base64.StdEncoding.DecodeString(videoLevels)
	if err != nil {
		return nil, err
	}
	return utils.GenUrls(gjson.GetBytes(data, "streamUrl").String())
}

func (l *Live) GetPlatformCNName() string {
	return cnName
}
