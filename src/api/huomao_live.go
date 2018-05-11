package api

import (
	"github.com/hr3lxphr6j/bililive-go/src/lib/http"
	"github.com/hr3lxphr6j/bililive-go/src/lib/utils"
	"github.com/tidwall/gjson"
	"net/url"
	"regexp"
)

const huomaoLiveApiUrl = "http://www.huomao.com/swf/live_data"

type HuoMaoLive struct {
	abstractLive
}

func (h *HuoMaoLive) GetInfo() (*Info, error) {
	dom, err := http.Get(h.Url.String(), nil, nil)
	if err != nil {
		return nil, err
	}
	info := &Info{
		Live:     h,
		HostName: utils.ParseUnicode(regexp.MustCompile(`"nickname":"([^"]*)"`).FindStringSubmatch(string(dom))[1]),
		RoomName: utils.ParseUnicode(regexp.MustCompile(`"channel":"([^"]*)"`).FindStringSubmatch(string(dom))[1]),
		Status:   utils.ParseUnicode(regexp.MustCompile(`"is_live":"?(\d*)"?,`).FindStringSubmatch(string(dom))[1]) == "1",
	}
	h.cachedInfo = info
	return info, nil
}

func (h *HuoMaoLive) GetStreamUrls() ([]*url.URL, error) {
	dom, err := http.Get(h.Url.String(), nil, nil)
	if err != nil {
		return nil, err
	}
	streamID := regexp.MustCompile(`"stream":"([^"]*)"`).FindStringSubmatch(string(dom))[1]
	body, err := http.Post(huomaoLiveApiUrl, map[string]string{
		"VideoIDS":   streamID,
		"streamtype": "live",
		"cdns":       "1",
		"from":       "huomaoh5room",
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
