package api

import (
	"errors"
	"fmt"
	"github.com/hr3lxphr6j/bililive-go/src/lib/utils"
	"net/url"
)

var LivePlatformCNNameMap = map[string]string{
	"www.panda.tv":      "熊猫",
	"live.bilibili.com": "哔哩哔哩",
	"www.zhanqi.tv":     "战旗",
	"www.douyu.com":     "斗鱼",
	"star.longzhu.com":  "龙珠",
	"www.huomao.com":    "火猫",
	"www.yizhibo.com":   "一直播",
	"www.twitch.tv":     "twitch",
	"www.huya.com":      "虎牙",
	"www.quanmin.tv":    "全民",
	"cc.163.com":        "CC直播",
	"www.openrec.tv":    "openrec",
}

type Info struct {
	Live               Live
	HostName, RoomName string
	Status             bool
}

type LiveId string

type Live interface {
	GetLiveId() LiveId
	GetRawUrl() string
	GetInfo() (*Info, error)
	GetInfoMap() map[string]interface{}
	GetCachedInfo() *Info
	GetStreamUrls() ([]*url.URL, error)
	GetPlatformCNName() string
}

type abstractLive struct {
	Url        *url.URL
	cachedInfo *Info
	liveId     LiveId
}

func (a *abstractLive) GetLiveId() LiveId {
	return a.liveId
}

func (a *abstractLive) GetRawUrl() string {
	return a.Url.String()
}

func (a *abstractLive) GetCachedInfo() *Info {
	return a.cachedInfo
}

func (a *abstractLive) GetInfoMap() map[string]interface{} {
	return map[string]interface{}{
		"id":        a.GetLiveId(),
		"url":       a.GetRawUrl(),
		"host_name": a.GetCachedInfo().HostName,
		"room_name": a.GetCachedInfo().RoomName,
		"status":    a.GetCachedInfo().Status,
	}
}

func (a *abstractLive) GetPlatformCNName() string {
	return LivePlatformCNNameMap[a.Url.Host]
}

type RoomNotExistsError struct {
	Url *url.URL
}

func (e *RoomNotExistsError) Error() string {
	return "room not exists"
}

func IsRoomNotExistsError(err error) bool {
	_, ok := err.(*RoomNotExistsError)
	return ok
}

func NewLive(url *url.URL) (Live, error) {
	baseLive := abstractLive{
		Url:    url,
		liveId: LiveId(utils.GetMd5String([]byte(fmt.Sprintf("%s%s", url.Host, url.Path)))),
	}
	var live Live
	switch url.Host {
	case "www.panda.tv":
		live = &PandaLive{abstractLive: baseLive}
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
	case "www.quanmin.tv":
		live = &QuanMinLive{abstractLive: baseLive}
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
			} else {
				return live, nil
			}
		}
	}
	return nil, errors.New("cat not parse")
}
