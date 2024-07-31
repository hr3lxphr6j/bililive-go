package stripchat

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/hr3lxphr6j/bililive-go/src/live"
	"github.com/hr3lxphr6j/bililive-go/src/live/internal"
	"github.com/hr3lxphr6j/bililive-go/src/pkg/utils"
	"github.com/parnurzeal/gorequest"
	"github.com/tidwall/gjson"
)

func get_modelId(modleName string) string {

	// modleName := "S-wan"
	// 创建一个新的 Request 对象
	request := gorequest.New()

	// 添加头部信息
	request.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	request.Set("Accept-Language", "zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2")
	request.Set("Accept-Encoding", "gzip, deflate")
	request.Set("Upgrade-Insecure-Requests", "1")
	request.Set("Sec-Fetch-Dest", "document")
	request.Set("Sec-Fetch-Mode", "navigate")
	request.Set("Sec-Fetch-Site", "none")
	request.Set("Sec-Fetch-User", "?1")
	request.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:124.0) Gecko/20100101 Firefox/124.0")
	// request.Set("If-Modified-Since", "Mon, 29 Jul 2024 08:41:12 GMT")
	request.Set("Te", "trailers")
	request.Set("Connection", "close")

	// 发起 GET 请求
	_, body, errs := request.Get("https://zh.stripchat.com/api/front/v2/models/username/" + modleName + "/chat").End()

	// 处理响应
	if len(errs) > 0 {
		fmt.Println("请求出错:", errs)
		return "false"
	} else {
		// 解析 JSON 响应
		if (len(gjson.Get(body, "messages").String())) > 2 {
			modelId := gjson.Get(body, "messages.0.modelId").String()
			return modelId
		} else {
			return "OffLine"
		}
	}
}

func get_M3u8(modelId string) string {
	// fmt.Println(modelId)
	url := "https://edge-hls.doppiocdn.com/hls/" + modelId + "/master/" + modelId + "_auto.m3u8?playlistType=lowLatency"
	request := gorequest.New()
	resp, body, errs := request.Get(url).End()

	if len(errs) > 0 || resp.StatusCode == 404 || modelId == "false" {
		return "false"
	} else {
		// fmt.Println((body))
		// re := regexp.MustCompile(`(https:\/\/[\w\-\.]+\/hls\/[\d]+\/[\d]+\.m3u8\?playlistType=lowLatency)`)
		re := regexp.MustCompile(`(https:\/\/[\w\-\.]+\/hls\/[\d]+\/[\d\_p]+\.m3u8\?playlistType=lowLatency)`)

		matches := re.FindString(body)
		return matches
	}
}

// func main() {
// 	// m3u8 := get_M3u8(get_modelId("Sakura_Anne"))
// 	m3u8 := get_M3u8(get_modelId("Lucky-uu"))
// 	fmt.Println(m3u8)
// }

const (
	domain = "zh.stripchat.com"
	cnName = "stripchat"
)

type Live struct {
	internal.BaseLive
}

func init() {
	live.Register(domain, new(builder))
}

type builder struct{}

func (b *builder) Build(url *url.URL, opt ...live.Option) (live.Live, error) {
	return &Live{
		BaseLive: internal.NewBaseLive(url, opt...),
	}, nil
}

func (l *Live) GetInfo() (info *live.Info, err error) {
	modeName := strings.Split(l.Url.String(), "/")
	modelName := modeName[len(modeName)-1]
	modelID := get_modelId(modelName)
	m3u8 := get_M3u8(modelID)

	if modelID == "false" {
		return nil, live.ErrRoomNotExist
	}
	if (modelID == "OffLine") && (m3u8 == "false") {
		info = &live.Info{
			Live:     l,
			HostName: modelID,
			RoomName: modelName,
			Status:   false,
		}
		return info, nil
	}
	if m3u8 != "false" {
		info = &live.Info{
			Live:     l,
			HostName: modelID,
			RoomName: modelName,
			Status:   true,
		}
		return info, nil
	}
	return info, live.ErrRoomUrlIncorrect
}

func (l *Live) GetStreamUrls() (us []*url.URL, err error) {
	// modeName := regexp.MustCompile(`stripchat.com\/(\w|-)+`).FindString(l.Url.String())
	modeName := strings.Split(l.Url.String(), "/")
	modelName := modeName[len(modeName)-1]
	m3u8 := get_M3u8(get_modelId(modelName))
	if m3u8 == "false" {
		return nil, err //live.ErrRoomNotExist
	}
	return utils.GenUrls(m3u8)
}

func (l *Live) GetPlatformCNName() string {
	return cnName
}
