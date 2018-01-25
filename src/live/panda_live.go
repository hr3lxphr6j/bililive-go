package live

import (
	"net/url"
	"net/http"
	"github.com/mozillazg/request"
	"strings"
	"strconv"
	"time"
)

type PandaLive struct {
	Url *url.URL
}

func (p *PandaLive) GetRoom() (*Info, error) {
	c := new(http.Client)
	req := request.NewRequest(c)
	req.Headers = commonHeader
	req.Params = map[string]string{
		"roomid": strings.Split(p.Url.Path, "/")[1],
		"__plat": "pc_web",
		"_":      strconv.FormatInt(time.Now().Unix(), 10),
	}
	resp, err := req.Get("http://www.panda.tv/api_room_v2")
	if err != nil {
		return nil, err
	}
	respMap, err := resp.Json()
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (p *PandaLive) GetUrls() ([]url.URL, error) {
	return nil, nil
}
