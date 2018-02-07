package api

import (
	"encoding/base64"
	"fmt"
	"github.com/hr3lxphr6j/bililive-go/src/lib/http"
	"github.com/tidwall/gjson"
	"net/url"
	"strings"
)

const (
	zhanQiApiUrl = "https://www.zhanqi.tv/api/static/v2.1/room/domain/%s.json"
)

type ZhanQiLive struct {
	Url *url.URL
}

func (z *ZhanQiLive) requestRoomInfo() ([]byte, error) {
	body, err := http.Get(fmt.Sprintf(zhanQiApiUrl, strings.Split(z.Url.Path, "/")[1]), nil)
	if err != nil {
		return nil, err
	}
	if gjson.GetBytes(body, "code").Int() == 0 {
		return body, nil
	} else {
		return nil, &RoomNotExistsError{z.Url}
	}
}
func (z *ZhanQiLive) GetRoom() (*Info, error) {
	body, err := z.requestRoomInfo()
	if err != nil {
		return nil, err
	}
	info := &Info{
		Live:     z,
		Url:      z.Url,
		HostName: gjson.GetBytes(body, "data.nickname").String(),
		RoomName: gjson.GetBytes(body, "data.title").String(),
		Status:   gjson.GetBytes(body, "data.status").String() == "4",
	}
	return info, nil
}

func (z *ZhanQiLive) GetUrls() ([]*url.URL, error) {
	body, err := z.requestRoomInfo()
	if err != nil {
		return nil, err
	}
	videoLevels := gjson.GetBytes(body, "data.flashvars.VideoLevels").String()
	data, err := base64.StdEncoding.DecodeString(videoLevels)
	if err != nil {
		return nil, err
	}
	m3u8, _ := url.Parse(gjson.GetBytes(data, "streamUrl").String())
	return []*url.URL{m3u8}, nil
}
