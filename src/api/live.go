package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/hr3lxphr6j/bililive-go/src/lib/utils"
)

type LiveId string

var LivePlatformCNNameMap = map[string]string{
	"live.bilibili.com": "哔哩哔哩",
	"www.zhanqi.tv":     "战旗",
	"www.douyu.com":     "斗鱼",
	"star.longzhu.com":  "龙珠",
	"www.huomao.com":    "火猫",
	"www.yizhibo.com":   "一直播",
	"www.twitch.tv":     "twitch",
	"www.huya.com":      "虎牙",
	"cc.163.com":        "CC直播",
	"www.openrec.tv":    "openrec",
}

type Info struct {
	Live                        Live
	HostName, RoomName          string
	Status, Listening, Recoding bool
}

func (i *Info) MarshalJSON() ([]byte, error) {
	t := struct {
		Id                LiveId `json:"id"`
		LiveUrl           string `json:"live_url"`
		PlatformCNName    string `json:"platform_cn_name"`
		HostName          string `json:"host_name"`
		RoomName          string `json:"room_name"`
		Status            bool   `json:"status"`
		Listening         bool   `json:"listening"`
		Recoding          bool   `json:"recoding"`
		LastStartTime     string `json:"last_start_time,omitempty"`
		LastStartTimeUnix int64  `json:"last_start_time_unix,omitempty"`
	}{
		Id:             i.Live.GetLiveId(),
		LiveUrl:        i.Live.GetRawUrl(),
		PlatformCNName: i.Live.GetPlatformCNName(),
		HostName:       i.HostName,
		RoomName:       i.RoomName,
		Status:         i.Status,
		Listening:      i.Listening,
		Recoding:       i.Recoding,
	}
	if !i.Live.GetLastStartTime().IsZero() {
		t.LastStartTime = i.Live.GetLastStartTime().Format("2006-01-02 15:04:05")
		t.LastStartTimeUnix = i.Live.GetLastStartTime().Unix()
	}
	return json.Marshal(t)
}

type Live interface {
	GetLiveId() LiveId
	GetRawUrl() string
	GetInfo() (*Info, error)
	GetInfoMap() map[string]interface{}
	GetCachedInfo() *Info
	GetStreamUrls() ([]*url.URL, error)
	GetPlatformCNName() string
	GetLastStartTime() time.Time
	SetLastStartTime(time.Time)
}

func NewLive(url *url.URL) (Live, error) {
	baseLive := abstractLive{
		Url:    url,
		liveId: LiveId(utils.GetMd5String([]byte(fmt.Sprintf("%s%s", url.Host, url.Path)))),
	}
	var live Live
	switch url.Host {
	case "live.bilibili.com":
		live = &BiliBiliLive{abstractLive: baseLive}
	case "www.zhanqi.tv":
		live = &ZhanQiLive{abstractLive: baseLive}
	case "www.douyu.com":
		live = &DouyuLive{abstractLive: baseLive}
	case "star.longzhu.com":
		live = &LongzhuLive{abstractLive: baseLive}
	case "www.huomao.com":
		live = &HuoMaoLive{abstractLive: baseLive}
	case "www.yizhibo.com":
		live = &YiZhiBoLive{abstractLive: baseLive}
	case "www.twitch.tv":
		live = &TwitchLive{abstractLive: baseLive}
	case "www.huya.com":
		live = &HuYaLive{abstractLive: baseLive}
	case "cc.163.com":
		live = &CCLive{abstractLive: baseLive}
	case "www.openrec.tv":
		live = &OpenRecLive{abstractLive: baseLive}
	default:
		live = nil
	}
	if live != nil {
		for i := 0; i < 3; i++ {
			if _, err := live.GetInfo(); err != nil {
				if IsRoomNotExistsError(err) {
					return nil, err
				}
				time.Sleep(1 * time.Second)
			} else {
				return live, nil
			}
		}
	}
	return nil, errors.New("can not parse")
}
