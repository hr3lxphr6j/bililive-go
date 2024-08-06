package internal

import (
	"fmt"
	"net/url"
	"time"

	"github.com/hr3lxphr6j/bililive-go/src/live"
	"github.com/hr3lxphr6j/bililive-go/src/pkg/utils"
)

type BaseLive struct {
	Url           *url.URL
	LastStartTime time.Time
	LiveId        live.ID
	Options       *live.Options
}

func genLiveId(url *url.URL) live.ID {
	return genLiveIdByString(fmt.Sprintf("%s%s", url.Host, url.Path))
}

func genLiveIdByString(value string) live.ID {
	return live.ID(utils.GetMd5String([]byte(value)))
}

func NewBaseLive(url *url.URL, opt ...live.Option) BaseLive {
	return BaseLive{
		Url:     url,
		LiveId:  genLiveId(url),
		Options: live.MustNewOptions(opt...),
	}
}

func (a *BaseLive) SetLiveIdByString(value string) {
	a.LiveId = genLiveIdByString(value)
}

func (a *BaseLive) GetLiveId() live.ID {
	return a.LiveId
}

func (a *BaseLive) GetRawUrl() string {
	return a.Url.String()
}

func (a *BaseLive) GetLastStartTime() time.Time {
	return a.LastStartTime
}

func (a *BaseLive) SetLastStartTime(time time.Time) {
	a.LastStartTime = time
}

// TODO: remove this method
func (a *BaseLive) GetStreamUrls() ([]*url.URL, error) {
	return nil, live.ErrNotImplemented
}

// TODO: remove this method
func (a *BaseLive) GetStreamInfos() ([]*live.StreamUrlInfo, error) {
	return nil, live.ErrNotImplemented
}
