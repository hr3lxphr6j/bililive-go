package api

import (
	"fmt"
	"github.com/hr3lxphr6j/bililive-go/src/lib/http"
	"github.com/hr3lxphr6j/bililive-go/src/lib/utils"
	"github.com/tidwall/gjson"
	"net/url"
	"strings"
	"time"
)

const (
	douyuLiveApiUrl = "http://www.douyutv.com/api/v1/"
	salt            = "zNzMV1y4EMxOHS6I5WKm"
)

var header = map[string]string{"user-agent": "Mozilla/5.0 (iPad; CPU OS 8_1_3 like Mac OS X) AppleWebKit/600.1.4 (KHTML, like Gecko) Version/8.0 Mobile/12B466 Safari/600.1.4"}

type DouyuLive struct {
	abstractLive
}

func (d *DouyuLive) requestRoomInfo() ([]byte, error) {
	args := fmt.Sprintf("room/%s?aid=wp&client_sys=wp&time=%d", strings.Split(d.Url.Path, "/")[1], time.Now().Unix())
	authMd5 := utils.GetMd5String([]byte(fmt.Sprintf("%s%s", args, salt)))
	body, err := http.Get(fmt.Sprintf("%s%s&auth=%s", douyuLiveApiUrl, args, authMd5), nil, header)
	if err != nil {
		return nil, err
	}
	if gjson.GetBytes(body, "error").Int() != 0 {
		return nil, &RoomNotExistsError{d.Url}
	}
	return body, nil
}

func (d *DouyuLive) GetInfo() (*Info, error) {
	data, err := d.requestRoomInfo()
	if err != nil {
		return nil, err
	}
	info := &Info{
		Live:     d,
		HostName: gjson.GetBytes(data, "data.nickname").String(),
		RoomName: gjson.GetBytes(data, "data.room_name").String(),
		Status:   gjson.GetBytes(data, "data.show_status").String() == "1",
	}
	d.cachedInfo = info
	return info, nil

}

func (d *DouyuLive) GetStreamUrls() ([]*url.URL, error) {
	data, err := d.requestRoomInfo()
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(fmt.Sprintf("%s/%s", gjson.GetBytes(data, "data.rtmp_url"), gjson.GetBytes(data, "data.rtmp_live")))
	if err != nil {
		return nil, err
	}
	return []*url.URL{u}, nil
}
