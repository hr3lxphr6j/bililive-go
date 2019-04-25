package api

import (
	"fmt"
	"net/url"
	"regexp"

	"github.com/tidwall/gjson"

	"github.com/hr3lxphr6j/bililive-go/src/lib/http"
)

const (
	ccLiveApiUrl = "http://cgi.v.cc.163.com/video_play_url/"
)

type CCLive struct {
	abstractLive
	ccid string
}

func (c *CCLive) GetInfo() (info *Info, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()

	dom, err := http.Get(c.Url.String(), nil, nil)
	if err != nil {
		return nil, err
	}
	c.ccid = regexp.MustCompile(`anchorCcId\s*:\s*'(\d*)'`).FindStringSubmatch(string(dom))[1]
	info = &Info{
		Live:     c,
		HostName: regexp.MustCompile(`anchorName\s*:\s*'([^']*)',`).FindStringSubmatch(string(dom))[1],
		RoomName: regexp.MustCompile(`js-live-title nick" title\s*=\s*"([^"]*)"`).FindStringSubmatch(string(dom))[1],
		Status:   len(regexp.MustCompile(`isLive\s*:\s*\d+,`).FindAll(dom, -1)) > 0,
	}
	c.cachedInfo = info
	return info, nil
}

func (c *CCLive) GetStreamUrls() (us []*url.URL, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()
	data, err := http.Get(fmt.Sprintf("%s%s", ccLiveApiUrl, c.ccid), nil, nil)
	if err != nil {
		return nil, err
	}
	us = make([]*url.URL, 0)
	u0, _ := url.Parse(gjson.GetBytes(data, "videourl").String())
	u1, _ := url.Parse(gjson.GetBytes(data, "bakvideourl").String())
	us = append(us, u0, u1)
	return us, nil
}
