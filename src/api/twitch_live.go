package api

import (
	"fmt"
	"math/rand"
	"net/url"
	"strings"

	"github.com/tidwall/gjson"

	"github.com/hr3lxphr6j/bililive-go/src/lib/http"
)

const (
	twitchClientId      = "jzkbprff40iqj646a697cyrvl0zt2m6"
	twitchChannelApiUrl = "https://api.twitch.tv/kraken/channels/%s"
	twitchLiveBaseUrl   = "https://usher.ttvnw.net/api/channel/hls/%s.m3u8"
	twitchStreamApiUrl  = "https://api.twitch.tv/kraken/streams/%s"
	twitchTokenApiUrl   = "https://api.twitch.tv/api/channels/%s/access_token"
)

var twitchHeader = map[string]string{"client-id": twitchClientId}

type TwitchLive struct {
	abstractLive
	hostName, roomName string
}

// 在hostName, roomName为空执行，在live有效时再从steam api解析
func (t *TwitchLive) parseInfo() error {
	chanId := strings.Split(t.Url.Path, "/")[1]
	body, err := http.Get(fmt.Sprintf(twitchChannelApiUrl, chanId), nil, twitchHeader)
	if err != nil {
		return &RoomNotExistsError{t.Url}
	}
	t.hostName = gjson.GetBytes(body, "name").String()
	t.roomName = gjson.GetBytes(body, "status").String()
	return nil

}

func (t *TwitchLive) GetInfo() (info *Info, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()

	if t.hostName == "" || t.roomName == "" {
		if err := t.parseInfo(); err != nil {
			return nil, err
		}
	}
	body, err := http.Get(fmt.Sprintf(twitchStreamApiUrl, t.hostName), nil, twitchHeader)
	if err != nil {
		return nil, err
	}
	status := gjson.GetBytes(body, "stream").String() != ""
	if status {
		t.roomName = gjson.GetBytes(body, "stream.channel.status").String()
	}
	info = &Info{
		Live:     t,
		HostName: t.hostName,
		RoomName: t.roomName,
		Status:   status,
	}
	t.cachedInfo = info
	return info, nil
}

func (t *TwitchLive) GetStreamUrls() (us []*url.URL, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()

	if t.hostName == "" || t.roomName == "" {
		if err := t.parseInfo(); err != nil {
			return nil, err
		}
	}
	body, err := http.Get(fmt.Sprintf(twitchTokenApiUrl, t.hostName), nil, twitchHeader)
	if err != nil {
		return nil, err
	}
	token := gjson.GetBytes(body, "token").String()
	sig := gjson.GetBytes(body, "sig").String()
	p := fmt.Sprintf("%d", rand.Intn(9000000)+1000000)
	u, err := url.Parse(fmt.Sprintf(twitchLiveBaseUrl, t.hostName))
	v := &url.Values{}
	v.Add("allow_source", "true")
	v.Add("allow_audio_only", "true")
	v.Add("allow_spectre", "true")
	v.Add("p", p)
	v.Add("player", "twitchweb")
	v.Add("segment_preference", "4")
	v.Add("sig", sig)
	v.Add("token", token)
	u.RawQuery = v.Encode()
	return []*url.URL{u}, nil
}
