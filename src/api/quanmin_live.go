package api

import (
	"fmt"
	"github.com/hr3lxphr6j/bililive-go/src/lib/http"
	"github.com/tidwall/gjson"
	"net/url"
	"regexp"
)

type QuanMinLive struct {
	abstractLive
}

func (q *QuanMinLive) requestRoomInfo() (string, error) {
	dom, err := http.Get(q.Url.String(), nil, nil)
	if err != nil {
		return "", err
	}
	if res := regexp.MustCompile("你想要的页面不存在噢！").FindStringSubmatch(string(dom)); res != nil {
		return "", &RoomNotExistsError{q.Url}
	}
	return regexp.MustCompile(`var roomModel = (.*)`).FindStringSubmatch(string(dom))[1], nil

}

func (q *QuanMinLive) GetInfo() (*Info, error) {
	roomModel, err := q.requestRoomInfo()
	if err != nil {
		return nil, err
	}
	info := &Info{
		Live:     q,
		HostName: gjson.Get(roomModel, "nick").String(),
		RoomName: gjson.Get(roomModel, "title").String(),
		Status:   gjson.Get(roomModel, "status").String() == "2",
	}
	q.cachedInfo = info
	return info, nil
}

func (q *QuanMinLive) GetStreamUrls() ([]*url.URL, error) {
	roomModel, err := q.requestRoomInfo()
	if err != nil {
		return nil, err
	}
	us := make([]*url.URL, 0)
	gjson.Get(roomModel, "room_lines").ForEach(func(key, value gjson.Result) bool {
		level := value.Get("flv.main_pc").String()
		src := value.Get(fmt.Sprintf("flv.%s.src", level)).String()
		u, err := url.Parse(src)
		if err != nil {
			return true
		}
		us = append(us, u)
		return true
	})
	return us, nil
}
