package huomao

import (
	"fmt"
	"net/url"
	"time"

	"github.com/tidwall/gjson"

	"github.com/hr3lxphr6j/bililive-go/src/lib/http"
	"github.com/hr3lxphr6j/bililive-go/src/lib/utils"
	"github.com/hr3lxphr6j/bililive-go/src/live"
	"github.com/hr3lxphr6j/bililive-go/src/live/internal"
)

const (
	domain = "www.huomao.com"
	cnName = "火猫"

	liveApiUrl = "http://www.huomao.com/swf/live_data"
	salt       = "6FE26D855E1AEAE090E243EB1AF73685"
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
	isDuanbo bool
}

func (l *Live) GetInfo() (info *live.Info, err error) {
	dom, err := http.Get(l.Url.String(), nil, nil)
	if err != nil {
		return nil, err
	}
	l.isDuanbo = utils.Match1(`face_label\s?=\s?(\d*);`, string(dom)) == "1"
	var (
		hostNameReg string
		roomNameReg string
		statusReg   string
	)
	if l.isDuanbo {
		hostNameReg = `live_yz_h_nickName\s?=\s?"([^"]*)";`
		roomNameReg = `live_yz_h_channelName\s?=\s?"([^"]*)";`
		statusReg = `is_live\s?=\s?"?(\d*)"?;`
	} else {
		hostNameReg = `"nickname":"([^"]*)",`
		roomNameReg = `"channel":"([^"]*)"`
		statusReg = `"is_live":"?(\d*)"?,`
	}
	var (
		hostName = utils.Match1(hostNameReg, string(dom))
		roomName = utils.Match1(roomNameReg, string(dom))
		status   = utils.Match1(statusReg, string(dom))
	)
	if hostName == "" || roomName == "" || status == "" {
		return nil, live.ErrInternalError
	}
	info = &live.Info{
		Live:     l,
		HostName: hostName,
		RoomName: roomName,
		Status:   status == "1",
	}
	return info, nil
}

func (l *Live) GetStreamUrls() ([]*url.URL, error) {
	dom, err := http.Get(l.Url.String(), nil, nil)
	if err != nil {
		return nil, err
	}
	var streamReg string
	if !l.isDuanbo {
		streamReg = `"stream":"([^"]*)"`
	} else {
		streamReg = `getFlash\("\d*","([^"]*)","\d*"\);`
	}
	var (
		from     = "huomaoh5room"
		t        = fmt.Sprintf("%d", time.Now().Unix())
		streamID = utils.Match1(streamReg, string(dom))
		token    = utils.GetMd5String([]byte(fmt.Sprintf("%s%s%s%s", streamID, from, t, salt)))
	)
	body, err := http.Post(liveApiUrl, nil, map[string]string{
		"VideoIDS":   streamID,
		"streamtype": "live",
		"cdns":       "1",
		"from":       from,
		"time":       t,
		"token":      token,
	}, nil)
	if err != nil {
		return nil, err
	}
	urls := make([]string, 0, 0)
	gjson.GetBytes(body, "streamList.#.list.#.url").ForEach(func(key, value gjson.Result) bool {
		for _, u := range value.Array() {
			urls = append(urls, u.String())
		}
		return true
	})
	return utils.GenUrls(urls...)
}

func (l *Live) GetPlatformCNName() string {
	return cnName
}
