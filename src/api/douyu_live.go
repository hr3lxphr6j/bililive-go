package api

import (
	"fmt"
	"github.com/robertkrimen/otto"
	"github.com/satori/go.uuid"
	"github.com/tidwall/gjson"
	"net/url"
	"strings"
	"time"

	"github.com/hr3lxphr6j/bililive-go/src/lib/http"
)

const (
	douyuLiveInfoUrl = "https://open.douyucdn.cn/api/RoomApi/room"
	douyuLiveEncUrl  = "https://www.douyu.com/swf_api/homeH5Enc"
	douyuLiveAPIUrl  = "https://www.douyu.com/lapi/live/getH5Play"
)

var (
	cryptoJS []byte
	header   = map[string]string{
		"Referer":      "https://www.douyu.com",
		"content-type": "application/x-www-form-urlencoded",
	}
)

func loadCryptoJS() {
	body, err := http.Get("https://cdnjs.cloudflare.com/ajax/libs/crypto-js/3.1.9-1/crypto-js.min.js", nil, nil)
	if err != nil {
		// TODO: not panic
		panic(err)
	}
	cryptoJS = body
}

func getEngineWithCryptoJS() (*otto.Otto, error) {
	if cryptoJS == nil {
		loadCryptoJS()
	}
	engine := otto.New()
	if _, err := engine.Eval(cryptoJS); err != nil {
		return nil, err
	}
	return engine, nil
}

type DouyuLive struct {
	abstractLive
}

func (d *DouyuLive) GetInfo() (info *Info, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()

	body, err := http.Get(fmt.Sprintf("%s/%s", douyuLiveInfoUrl, strings.Split(d.Url.Path, "/")[1]), nil, nil)
	if err != nil {
		return nil, err
	}
	if gjson.GetBytes(body, "error").Int() != 0 {
		return nil, &RoomNotExistsError{d.Url}
	}
	info = &Info{
		Live:     d,
		HostName: gjson.GetBytes(body, "data.owner_name").String(),
		RoomName: gjson.GetBytes(body, "data.room_name").String(),
		Status:   gjson.GetBytes(body, "data.room_status").String() == "1",
	}
	d.cachedInfo = info
	return info, nil

}

func (d *DouyuLive) getSignParams() (url.Values, error) {
	roomID := strings.Split(d.Url.Path, "/")[1]
	body, err := http.Get(douyuLiveEncUrl, map[string]string{
		"rids": roomID,
	}, nil)
	if err != nil {
		return nil, err
	}
	jsEnc := gjson.GetBytes(body, fmt.Sprintf("data.room%s", roomID)).String()
	engine, err := getEngineWithCryptoJS()
	if err != nil {
		return nil, err
	}
	if _, err := engine.Eval(jsEnc); err != nil {
		return nil, err
	}
	did := strings.ReplaceAll(uuid.Must(uuid.NewV4()).String(), "-", "")
	ts := time.Now()
	res, err := engine.Call("ub98484234", nil, roomID, did, ts.Unix())
	if err != nil {
		return nil, err
	}

	values := url.Values{
		"cdn":  {""},
		"iar":  {"0"},
		"ive":  {"0"},
		"rate": {"0"},
	}
	for _, entry := range strings.Split(res.String(), "&") {
		if entry == "" {
			continue
		}
		strs := strings.SplitN(entry, "=", 2)
		values.Set(strs[0], strs[1])
	}
	return values, nil
}

func (d *DouyuLive) GetStreamUrls() (us []*url.URL, err error) {
	roomID := strings.Split(d.Url.Path, "/")[1]
	params, err := d.getSignParams()
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(fmt.Sprintf("%s/%s", douyuLiveAPIUrl, roomID), nil, []byte(params.Encode()), header)
	if gjson.GetBytes(resp, "error").Int() != 0 {
		return nil, fmt.Errorf("get stream error")
	}
	u, err := url.Parse(gjson.GetBytes(resp, "data.rtmp_url").String() + "/" + gjson.GetBytes(resp, "data.rtmp_live").String())
	if err != nil {
		return nil, err
	}
	return []*url.URL{u}, nil
}
