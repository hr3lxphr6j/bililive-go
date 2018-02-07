package api

import (
	"github.com/hr3lxphr6j/bililive-go/src/lib/http"
	"github.com/tidwall/gjson"
	"net/url"
	"strings"
)

const (
	biliBiliRoomInitUrl = "https://api.live.bilibili.com/room/v1/Room/room_init"
	biliBiliRoomApiUrl  = "https://api.live.bilibili.com/room/v1/Room/get_info"
	biliBiliUserApiUrl  = "https://api.live.bilibili.com/live_user/v1/UserInfo/get_anchor_in_room"
	biliBiliLiveApiUrl  = "https://api.live.bilibili.com/room/v1/Room/playUrl"
)

type BiliBiliLive struct {
	Url             *url.URL
	shortId, fullId string
}

func (b *BiliBiliLive) GetRoom() (*Info, error) {
	// Parse the short id from URL to full id
	if b.shortId == "" || b.fullId == "" {
		b.shortId = strings.Split(b.Url.Path, "/")[1]
		if body, err := http.Get(biliBiliRoomInitUrl, map[string]string{"id": b.shortId}); err != nil {
			return nil, err
		} else {
			if gjson.GetBytes(body, "code").Int() != 0 {
				return nil, &RoomNotExistsError{b.Url}
			} else {
				b.fullId = gjson.GetBytes(body, "data.room_id").String()
			}
		}
	}

	body, err := http.Get(biliBiliRoomApiUrl, map[string]string{
		"room_id": b.fullId,
		"from":    "room",
	})
	if err != nil {
		return nil, err
	}
	if gjson.GetBytes(body, "code").Int() != 0 {
		return nil, &RoomNotExistsError{b.Url}
	}

	info := &Info{
		Live:     b,
		Url:      b.Url,
		RoomName: gjson.GetBytes(body, "data.title").String(),
		Status:   gjson.GetBytes(body, "data.live_status").Int() == 1,
	}

	body2, err := http.Get(biliBiliUserApiUrl, map[string]string{
		"roomid": strings.Split(b.Url.Path, "/")[1],
	})
	if err != nil {
		return nil, err
	}

	info.HostName = gjson.GetBytes(body2, "data.info.uname").String()
	return info, nil
}

func (b *BiliBiliLive) GetUrls() ([]*url.URL, error) {
	body, err := http.Get(biliBiliLiveApiUrl, map[string]string{
		"cid":      b.fullId,
		"quality":  "0",
		"platform": "web",
	})
	if err != nil {
		return nil, err
	}

	urls := gjson.GetBytes(body, "data.durl.#.url").Array()

	us := make([]*url.URL, 0, 4)
	for _, u := range urls {
		url_, _ := url.Parse(u.String())
		us = append(us, url_)
	}
	return us, nil
}
