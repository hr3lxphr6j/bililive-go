package api

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/hr3lxphr6j/bililive-go/src/lib/http"
	"github.com/hr3lxphr6j/bililive-go/src/lib/utils"
)

type OpenRecLive struct {
	abstractLive
}

func (o *OpenRecLive) GetInfo() (info *Info, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()

	dom, err := http.Get(o.Url.String(), nil, nil)
	if err != nil {
		return nil, err
	}
	info = &Info{
		Live:     o,
		RoomName: strings.TrimSpace(regexp.MustCompile(`"title":"([^:]*)",`).FindStringSubmatch(string(dom))[1]),
		HostName: utils.ParseUnicode(regexp.MustCompile(`"name":"([^:]*)",`).FindStringSubmatch(string(dom))[1]),
		Status:   regexp.MustCompile(`"onairStatus":(\d),`).FindStringSubmatch(string(dom))[1] == "1",
	}
	o.cachedInfo = info
	return info, nil
}

func (o *OpenRecLive) GetStreamUrls() (us []*url.URL, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()

	dom, err := http.Get(o.Url.String(), nil, nil)
	if err != nil {
		return nil, err
	}
	u1, _ := url.Parse(regexp.MustCompile(`{"url":"(\S*m3u8)",`).FindStringSubmatch(string(dom))[1])
	us = []*url.URL{u1}
	return us, nil
}
