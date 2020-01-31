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
}

func genLiveId(url *url.URL) live.ID {
	return live.ID(utils.GetMd5String([]byte(fmt.Sprintf("%s%s", url.Host, url.Path))))
}

func NewBaseLive(url *url.URL) BaseLive {
	return BaseLive{
		Url:    url,
		LiveId: genLiveId(url),
	}
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
