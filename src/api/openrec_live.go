package api

import (
	"github.com/hr3lxphr6j/bililive-go/src/lib/http"
	"github.com/hr3lxphr6j/bililive-go/src/lib/utils"
	"net/url"
	"regexp"
	"strings"
)

type OpenRecLive struct {
	abstractLive
}

func (o *OpenRecLive) GetInfo() (*Info, error) {
	dom, err := http.Get(o.Url.String(), nil, nil)
	if err != nil {
		return nil, err
	}
	info := &Info{
		Live:     o,
		RoomName: strings.TrimSpace(regexp.MustCompile(`<div class="p-playbox__content__info__title">[\n\r]*([^\n\r]*)[\n\r]*</div>`).FindStringSubmatch(string(dom))[1]),
		HostName: utils.ParseUnicode(regexp.MustCompile(`"nickname":"([^:]*)",`).FindStringSubmatch(string(dom))[1]),
		Status:   regexp.MustCompile(`gbl_onair_status\s*=\s(\d*);`).FindStringSubmatch(string(dom))[1] == "1",
	}
	o.cachedInfo = info
	return info, nil
}

func (o *OpenRecLive) GetStreamUrls() ([]*url.URL, error) {
	dom, err := http.Get(o.Url.String(), nil, nil)
	if err != nil {
		return nil, err
	}
	u1, _ := url.Parse(regexp.MustCompile(`data-file="(.*)"`).FindStringSubmatch(string(dom))[1])
	us := []*url.URL{u1}
	return us, nil
}
