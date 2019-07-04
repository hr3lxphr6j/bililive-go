package api

import (
	"fmt"
	"math/rand"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/hr3lxphr6j/bililive-go/src/lib/http"
	"github.com/hr3lxphr6j/bililive-go/src/lib/utils"
)

type HuYaLive struct {
	abstractLive
}

func (h *HuYaLive) GetInfo() (info *Info, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()

	dom, err := http.Get(h.Url.String(), nil, nil)
	if err != nil {
		return nil, err
	}
	if res := regexp.MustCompile("哎呀，虎牙君找不到这个主播，要不搜索看看？").FindStringSubmatch(string(dom)); res != nil {
		return nil, &RoomNotExistsError{h.Url}
	}
	info = &Info{
		Live:     h,
		HostName: utils.ParseUnicode(regexp.MustCompile(`"nick":"([^"]*)"`).FindStringSubmatch(string(dom))[1]),
		RoomName: utils.ParseUnicode(regexp.MustCompile(`"introduction":"([^"]*)"`).FindStringSubmatch(string(dom))[1]),
		Status:   utils.ParseUnicode(regexp.MustCompile(`"isOn":([^,]*),`).FindStringSubmatch(string(dom))[1]) == "true",
	}
	h.cachedInfo = info
	return info, nil
}

func (h *HuYaLive) GetStreamUrls() (us []*url.URL, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()

	dom, err := http.Get(h.Url.String(), nil, nil)
	if err != nil {
		return nil, err
	}
	sStreamName := regexp.MustCompile(`"sStreamName":"([^"]*)"`).FindStringSubmatch(string(dom))[1]
	sFlvUrl := strings.Replace(regexp.MustCompile(`"sFlvUrl":"([^"]*)"`).FindStringSubmatch(string(dom))[1], `\/`, `/`, -1)
	sFlvAntiCode := regexp.MustCompile(`"sFlvAntiCode":"([^"]*)"`).FindStringSubmatch(string(dom))[1]
	iLineIndex := regexp.MustCompile(`"iLineIndex":(\d*),`).FindStringSubmatch(string(dom))[1]
	uid := (time.Now().Unix()%1e7*1e6 + int64(1e3*rand.Float64())) % 4294967295
	u, err := url.Parse(fmt.Sprintf("%s/%s.flv", sFlvUrl, sStreamName))
	if err != nil {
		return nil, err
	}
	value := &url.Values{}
	value.Add("line", iLineIndex)
	value.Add("p2p", "0")
	value.Add("type", "web")
	value.Add("ver", "1805071653")
	value.Add("uid", fmt.Sprintf("%d", uid))
	u.RawQuery = fmt.Sprintf("%s&%s", value.Encode(), strings.ReplaceAll(sFlvAntiCode, "&amp;", "&"))
	return []*url.URL{u}, nil
}
