package api

import (
	"errors"
	"fmt"
	"github.com/hr3lxphr6j/bililive-go/src/lib/utils"
	"net/url"
)

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
