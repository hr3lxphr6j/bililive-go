package api

import (
	"net/url"
)

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
