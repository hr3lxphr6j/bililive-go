package api

import (
	"errors"
	"net/url"
	"regexp"
	"strings"

	"github.com/tidwall/gjson"

	"github.com/hr3lxphr6j/bililive-go/src/lib/http"
)

var (
	qqLiveMobileUrl   = "https://m.egame.qq.com/live"
	qqLiveStreamsReg  = regexp.MustCompile(`"urlArray":(\[[^\]]+\])`)
	qqLiveRoomNameReg = regexp.MustCompile(`title:"([^"]*)"`)
	qqLiveHostNameReg = regexp.MustCompile(`nickName:"([^"]+)"`)
	qqLiveIsLiveReg   = regexp.MustCompile(`"isLive":(\d+)`)
)

type QQLive struct {
	abstractLive
}

func (q *QQLive) GetInfo() (info *Info, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()

	anchorID := strings.Split(q.Url.Path, "/")[1]
	dom, err := http.Get(q.Url.String(), nil, nil)
	if err != nil {
		return nil, err
	}
	roomName := qqLiveRoomNameReg.FindSubmatch(dom)[1]
	hostName := qqLiveHostNameReg.FindSubmatch(dom)[1]
	dom2, err := http.Get(qqLiveMobileUrl, map[string]string{
		"anchorid": anchorID,
	}, nil)
	isLive := string(qqLiveIsLiveReg.FindSubmatch(dom2)[1]) == "1"
	info = &Info{
		Live:     q,
		RoomName: string(roomName),
		HostName: string(hostName),
		Status:   isLive,
	}
	q.cachedInfo = info
	return info, nil
}

func (q *QQLive) GetStreamUrls() (us []*url.URL, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()

	dom, err := http.Get(q.Url.String(), nil, nil)
	if err != nil {
		return nil, err
	}
	result := qqLiveStreamsReg.FindSubmatch(dom)
	if len(result) != 2 {
		return nil, errors.New("failed to get streams")
	}
	u, err := url.Parse(gjson.GetBytes(result[1], "#[bitrate==0].playUrl").String())
	if err != nil {
		return nil, err
	}
	return []*url.URL{u}, nil
}
