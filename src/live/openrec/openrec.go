package openrec

import (
	"net/url"
	"strings"

	"github.com/hr3lxphr6j/bililive-go/src/lib/http"
	"github.com/hr3lxphr6j/bililive-go/src/lib/utils"
	"github.com/hr3lxphr6j/bililive-go/src/live"
	"github.com/hr3lxphr6j/bililive-go/src/live/internal"
)

const (
	domain = "www.openrec.tv"
	cnName = "openrec"
)

type Live struct {
	internal.BaseLive
}

func init() {
	live.Register(domain, new(builder))
}

type builder struct{}

func (b *builder) Build(url *url.URL) (live.Live, error) {
	return &Live{
		BaseLive: internal.NewBaseLive(url),
	}, nil
}

func (l *Live) GetInfo() (info *live.Info, err error) {
	dom, err := http.Get(l.Url.String(), nil, nil)
	if err != nil {
		return nil, err
	}
	var (
		roomName = utils.ParseString(
			utils.Match1(`"title":"([^:]*)",`, string(dom)),
			utils.StringFilterFunc(strings.TrimSpace),
			utils.UnescapeHTMLEntity,
		)
		hostName = utils.ParseString(
			utils.Match1(`"name":"([^:]*)",`, string(dom)),
			utils.ParseUnicode,
			utils.UnescapeHTMLEntity,
		)
		status = utils.Match1(`"onairStatus":(\d),`, string(dom))
	)
	if roomName == "" || hostName == "" || status == "" {
		return nil, live.ErrInternalError
	}
	info = &live.Info{
		Live:     l,
		RoomName: roomName,
		HostName: hostName,
		Status:   status == "1",
	}
	return info, nil
}

func (l *Live) GetStreamUrls() (us []*url.URL, err error) {
	dom, err := http.Get(l.Url.String(), nil, nil)
	if err != nil {
		return nil, err
	}
	return utils.GenUrls(utils.Match1(`{"url":"(\S*m3u8)",`, string(dom)))
}

func (l *Live) GetPlatformCNName() string {
	return cnName
}
