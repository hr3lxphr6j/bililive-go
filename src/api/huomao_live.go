package api

import (
	"fmt"
	"net/url"
	"regexp"
	"time"

	"github.com/tidwall/gjson"

	"github.com/hr3lxphr6j/bililive-go/src/lib/http"
	"github.com/hr3lxphr6j/bililive-go/src/lib/utils"
)

const (
	huomaoLiveApiUrl = "http://www.huomao.com/swf/live_data"
	huomaoSalt       = "6FE26D855E1AEAE090E243EB1AF73685"
)

type HuoMaoLive struct {
	abstractLive
	isDuanbo bool
}

func (h *HuoMaoLive) GetInfo() (info *Info, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()

	dom, err := http.Get(h.Url.String(), nil, nil)
	if err != nil {
		return nil, err
	}
	h.isDuanbo = regexp.MustCompile(`face_label\s?=\s?(\d*);`).FindStringSubmatch(string(dom))[1] == "1"
	var hostNameReg string
	var roomNameReg string
	var statusReg string
	if h.isDuanbo {
		hostNameReg = `live_yz_h_nickName\s?=\s?"([^"]*)";`
		roomNameReg = `live_yz_h_channelName\s?=\s?"([^"]*)";`
		statusReg = `is_live\s?=\s?"?(\d*)"?;`
	} else {
		hostNameReg = `"nickname":"([^"]*)",`
		roomNameReg = `"channel":"([^"]*)"`
		statusReg = `"is_live":"?(\d*)"?,`
	}
	info = &Info{
		Live:     h,
		HostName: utils.ParseUnicode(regexp.MustCompile(hostNameReg).FindStringSubmatch(string(dom))[1]),
		RoomName: utils.ParseUnicode(regexp.MustCompile(roomNameReg).FindStringSubmatch(string(dom))[1]),
		Status:   utils.ParseUnicode(regexp.MustCompile(statusReg).FindStringSubmatch(string(dom))[1]) == "1",
	}
	h.cachedInfo = info
	return info, nil
}

func (h *HuoMaoLive) GetStreamUrls() ([]*url.URL, error) {
	dom, err := http.Get(h.Url.String(), nil, nil)
	if err != nil {
		return nil, err
	}
	var streamReg string
	if !h.isDuanbo {
		streamReg = `"stream":"([^"]*)"`
	} else {
		streamReg = `getFlash\("\d*","([^"]*)","\d*"\);`
	}
	from := "huomaoh5room"
	t := fmt.Sprintf("%d", time.Now().Unix())
	streamID := regexp.MustCompile(streamReg).FindStringSubmatch(string(dom))[1]
	token := utils.GetMd5String([]byte(fmt.Sprintf("%s%s%s%s", streamID, from, t, huomaoSalt)))
	body, err := http.Post(huomaoLiveApiUrl, map[string]string{
		"VideoIDS":   streamID,
		"streamtype": "live",
		"cdns":       "1",
		"from":       from,
		"time":       t,
		"token":      token,
	}, nil, nil)
	us := make([]*url.URL, 0, 4)
	gjson.GetBytes(body, "streamList.#.list.#.url").ForEach(func(key, value gjson.Result) bool {
		for _, u := range value.Array() {
			url_, _ := url.Parse(u.String())
			us = append(us, url_)
		}
		return true
	})

	return us, nil
}
