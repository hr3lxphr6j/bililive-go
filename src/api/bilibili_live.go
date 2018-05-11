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
	abstractLive
	realId string
}

func (b *BiliBiliLive) parseRealId() error {
	if body, err := http.Get(biliBiliRoomInitUrl, map[string]string{"id": strings.Split(b.Url.Path, "/")[1]}, nil); err != nil {
		return err
	} else {
		if gjson.GetBytes(body, "code").Int() != 0 {
			return &RoomNotExistsError{b.Url}
		} else {
			b.realId = gjson.GetBytes(body, "data.room_id").String()
		}
		return nil
	}
}

func (b *BiliBiliLive) GetInfo() (*Info, error) {
	// Parse the short id from URL to full id
	if b.realId == "" {
		if err := b.parseRealId(); err != nil {
			return nil, err
		}
	}
	body, err := http.Get(biliBiliRoomApiUrl, map[string]string{
		"room_id": b.realId,
		"from":    "room",
	}, nil)
	if err != nil {
		return nil, err
	}
	if gjson.GetBytes(body, "code").Int() != 0 {
		return nil, &RoomNotExistsError{b.Url}
	}

	info := &Info{
		Live:     b,
		RoomName: gjson.GetBytes(body, "data.title").String(),
		Status:   gjson.GetBytes(body, "data.live_status").Int() == 1,
	}

	body2, err := http.Get(biliBiliUserApiUrl, map[string]string{
		"roomid": b.realId,
	}, nil)
	if err != nil {
		return nil, err
	}

	info.HostName = gjson.GetBytes(body2, "data.info.uname").String()
	b.cachedInfo = info
	return info, nil
}

func (b *BiliBiliLive) GetStreamUrls() ([]*url.URL, error) {
	if b.realId == "" {
		if err := b.parseRealId(); err != nil {
			return nil, err
		}
	}
	body, err := http.Get(biliBiliLiveApiUrl, map[string]string{
		"cid":      b.realId,
		"quality":  "0",
		"platform": "web",
	}, nil)
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
