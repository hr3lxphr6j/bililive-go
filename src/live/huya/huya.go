package huya

import (
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"time"

	"github.com/hr3lxphr6j/bililive-go/src/lib/http"
	"github.com/hr3lxphr6j/bililive-go/src/lib/utils"
	"github.com/hr3lxphr6j/bililive-go/src/live"
	"github.com/hr3lxphr6j/bililive-go/src/live/internal"
)

const (
	domain = "www.huya.com"
	cnName = "虎牙"
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
	dom, err := http.Get(l.Url.String(), nil, nil)
	if err != nil {
		return nil, err
	}
	if res := utils.Match1("哎呀，虎牙君找不到这个主播，要不搜索看看？", string(dom)); res != "" {
		return nil, live.ErrRoomNotExist
	}

	var (
		strFilter = utils.NewStringFilterChain(utils.ParseUnicode, utils.UnescapeHTMLEntity)
		hostName  = strFilter.Do(utils.Match1(`"nick":"([^"]*)"`, string(dom)))
		roomName  = strFilter.Do(utils.Match1(`"introduction":"([^"]*)"`, string(dom)))
		status    = strFilter.Do(utils.Match1(`"isOn":([^,]*),`, string(dom)))
	)

	if hostName == "" || roomName == "" || status == "" {
		return nil, live.ErrInternalError
	}

	info = &live.Info{
		Live:     l,
		HostName: hostName,
		RoomName: roomName,
		Status:   status == "true",
	}
	return info, nil
}

func (l *Live) GetStreamUrls() (us []*url.URL, err error) {
	dom, err := http.Get(l.Url.String(), nil, nil)
	if err != nil {
		return nil, err
	}
	var (
		sStreamName  = utils.Match1(`"sStreamName":"([^"]*)"`, string(dom))
		sFlvUrl      = strings.ReplaceAll(utils.Match1(`"sFlvUrl":"([^"]*)"`, string(dom)), `\/`, `/`)
		sFlvAntiCode = utils.Match1(`"sFlvAntiCode":"([^"]*)"`, string(dom))
		iLineIndex   = utils.Match1(`"iLineIndex":(\d*),`, string(dom))
		uid          = (time.Now().Unix()%1e7*1e6 + int64(1e3*rand.Float64())) % 4294967295
	)
	u, err := url.Parse(fmt.Sprintf("%s/%s.flv", sFlvUrl, sStreamName))
	if err != nil {
		return nil, err
	}
	value := url.Values{}
	value.Add("line", iLineIndex)
	value.Add("p2p", "0")
	value.Add("type", "web")
	value.Add("ver", "1805071653")
	value.Add("uid", fmt.Sprintf("%d", uid))
	u.RawQuery = fmt.Sprintf("%s&%s", value.Encode(), utils.UnescapeHTMLEntity(sFlvAntiCode))
	return []*url.URL{u}, nil
}

func (l *Live) GetPlatformCNName() string {
	return cnName
}
