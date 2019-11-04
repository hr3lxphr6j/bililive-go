package twitch

import (
	"fmt"
	"math/rand"
	"net/url"
	"strings"

	"github.com/tidwall/gjson"

	"github.com/hr3lxphr6j/bililive-go/src/lib/http"
	"github.com/hr3lxphr6j/bililive-go/src/live"
	"github.com/hr3lxphr6j/bililive-go/src/live/internal"
)

const (
	domain = "www.twitch.tv"
	cnName = "twitch"

	clientId      = "jzkbprff40iqj646a697cyrvl0zt2m6"
	channelApiUrl = "https://api.twitch.tv/kraken/channels/%s"
	liveBaseUrl   = "https://usher.ttvnw.net/api/channel/hls/%s.m3u8"
	streamApiUrl  = "https://api.twitch.tv/kraken/streams/%s"
	tokenApiUrl   = "https://api.twitch.tv/api/channels/%s/access_token"
)

func init() {
	live.Register(domain, new(builder))
}

type builder struct{}

func (b *builder) Build(url *url.URL) (live.Live, error) {
	return &Live{
		BaseLive: internal.NewBaseLive(url),
	}, nil
}

var headers = map[string]string{"client-id": clientId}

type Live struct {
	internal.BaseLive
	hostName, roomName string
}

// 在hostName, roomName为空执行，在live有效时再从steam api解析
func (l *Live) parseInfo() error {
	paths := strings.Split(l.Url.Path, "/")
	if len(paths) < 2 {
		return live.ErrRoomUrlIncorrect
	}
	chanId := paths[1]
	body, err := http.Get(fmt.Sprintf(channelApiUrl, chanId), headers, nil)
	if err != nil {
		return live.ErrRoomNotExist
	}
	l.hostName = gjson.GetBytes(body, "name").String()
	l.roomName = gjson.GetBytes(body, "status").String()
	return nil
}

func (l *Live) GetInfo() (info *live.Info, err error) {
	if l.hostName == "" || l.roomName == "" {
		if err := l.parseInfo(); err != nil {
			return nil, err
		}
	}
	body, err := http.Get(fmt.Sprintf(streamApiUrl, l.hostName), headers, nil)
	if err != nil {
		return nil, err
	}
	status := gjson.GetBytes(body, "stream").String() != ""
	if status {
		l.roomName = gjson.GetBytes(body, "stream.channel.status").String()
	}
	info = &live.Info{
		Live:     l,
		HostName: l.hostName,
		RoomName: l.roomName,
		Status:   status,
	}
	return info, nil
}

func (l *Live) GetStreamUrls() (us []*url.URL, err error) {
	if l.hostName == "" || l.roomName == "" {
		if err := l.parseInfo(); err != nil {
			return nil, err
		}
	}
	body, err := http.Get(fmt.Sprintf(tokenApiUrl, l.hostName), headers, nil)
	if err != nil {
		return nil, err
	}
	var (
		token = gjson.GetBytes(body, "token").String()
		sig   = gjson.GetBytes(body, "sig").String()
		p     = fmt.Sprintf("%d", rand.Intn(9000000)+1000000)
	)
	u, err := url.Parse(fmt.Sprintf(liveBaseUrl, l.hostName))
	if err != nil {
		return nil, err
	}
	v := url.Values{}
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

func (l *Live) GetPlatformCNName() string {
	return cnName
}
