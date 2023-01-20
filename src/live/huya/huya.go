package huya

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hr3lxphr6j/requests"

	"github.com/hr3lxphr6j/bililive-go/src/live"
	"github.com/hr3lxphr6j/bililive-go/src/live/internal"
	"github.com/hr3lxphr6j/bililive-go/src/pkg/utils"
)

const (
	domain = "www.huya.com"
	cnName = "虎牙"
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
}

func (l *Live) GetInfo() (info *live.Info, err error) {
	resp, err := requests.Get(l.Url.String(), live.CommonUserAgent)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, live.ErrRoomNotExist
	}
	body, err := resp.Text()
	if err != nil {
		return nil, err
	}

	if res := utils.Match1("哎呀，虎牙君找不到这个主播，要不搜索看看？", body); res != "" {
		return nil, live.ErrRoomNotExist
	}

	if strings.Contains(body, "该主播涉嫌违规，正在整改中") {
		return &live.Info{
			Live:     l,
			HostName: "该主播涉嫌违规，正在整改中",
			RoomName: "该主播涉嫌违规，正在整改中",
			Status:   false,
		}, nil
	}

	var (
		strFilter = utils.NewStringFilterChain(utils.ParseUnicode, utils.UnescapeHTMLEntity)
		hostName  = strFilter.Do(utils.Match1(`"nick":"([^"]*)"`, body))
		roomName  = strFilter.Do(utils.Match1(`"introduction":"([^"]*)"`, body))
		status    = strFilter.Do(utils.Match1(`"isOn":([^,]*),`, body))
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
	resp, err := requests.Get(l.Url.String(), live.CommonUserAgent)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, live.ErrRoomNotExist
	}
	body, err := resp.Text()
	if err != nil {
		return nil, err
	}

	// Decode stream part.
	streamStr := utils.Match1(`(?m)stream: (.*?)$`, body)

	var (
		sStreamName  = utils.Match1(`"sStreamName":"([^"]*)"`, streamStr)
		sFlvUrl      = strings.ReplaceAll(utils.Match1(`"sFlvUrl":"([^"]*)"`, streamStr), `\/`, `/`)
		sFlvAntiCode = utils.Match1(`"sFlvAntiCode":"([^"]*)"`, streamStr)
		iLineIndex   = utils.Match1(`"iLineIndex":(\d*),`, streamStr)
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
