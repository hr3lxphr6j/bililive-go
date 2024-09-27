package stripchat

import (
	"fmt"
	"net"
	"net/url"
	"reflect"
	"regexp"
	"strings"

	"github.com/hr3lxphr6j/bililive-go/src/cmd/bililive/readconfig"
	"github.com/hr3lxphr6j/bililive-go/src/configs"
	"github.com/hr3lxphr6j/bililive-go/src/live"
	"github.com/hr3lxphr6j/bililive-go/src/live/internal"
	"github.com/hr3lxphr6j/bililive-go/src/pkg/utils"
	"github.com/parnurzeal/gorequest"
	"github.com/tidwall/gjson"
)

func get_modelId(modleName string, daili string) string {
	if modleName == "" {
		return "false"
	}
	request := gorequest.New()
	if daili != "" {
		request = request.Proxy(daili) //代理
	}

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
	if errs != nil {
		fmt.Println("get_modeId出错详情:")
		for _, err := range errs {
			if err1, ok := err.(*url.Error); ok {
				// urlErr 是 *url.Error 类型的错误
				// fmt.Println("*url.Error 类型的错误")
				if err2, ok := err1.Err.(*net.OpError); ok {
					// netErr 是 *net.OpError 类型的错误
					// 可以进一步判断 netErr.Err 的类型
					fmt.Println("*net.OpError 类型的错误", err.Error(), err2.Op)
				}
				return "url.Error"
			} else {
				fmt.Println(reflect.TypeOf(err), "错误详情:", err)
			}
		}
		return "false"
	} else {
		// 解析 JSON 响应
		if len(gjson.Get(body, "messages").String()) > 2 {
			modelId := gjson.Get(body, "messages.0.modelId").String()
			return modelId
		} else if len(gjson.Get(body, "messages").String()) == 2 {
			return "OffLine"
		} else if len(gjson.Get(body, "messages").String()) == 0 {
			// fmt.Println("error name")
			return "false"
		}
		return "false"
	}
}

func get_M3u8(modelId string, daili string) string {
	if modelId == "false" || modelId == "OffLine" || modelId == "url.Error" {
		return "false"
	}
	// url := "https://edge-hls.doppiocdn.com/hls/" + modelId + "/master/" + modelId + "_auto.m3u8?playlistType=lowLatency"
	url := "https://edge-hls.doppiocdn.com/hls/" + modelId + "/master/" + modelId + "_auto.m3u8?playlistType=standard"
	// url := "https://edge-hls.doppiocdn.com/hls/" + modelId + "/master/" + modelId + ".m3u8"
	request := gorequest.New()
	if daili != "" {
		request = request.Proxy(daili) //代理
	}
	resp, body, errs := request.Get(url).End()

	if len(errs) > 0 || resp.StatusCode != 200 {
		return "false"
	} else {
		// fmt.Println((body))
		// re := regexp.MustCompile(`(https:\/\/[\w\-\.]+\/hls\/[\d]+\/[\d\_p]+\.m3u8\?playlistType=lowLatency)`)
		re := regexp.MustCompile(`(https:\/\/[\w\-\.]+\/hls\/[\d]+\/[\d\_p]+\.m3u8\?playlistType=standard)`) //等价于\?playlistType=standard
		matches := re.FindString(body)
		return matches
	}
}
func test_m3u8(url string, daili string) bool {
	if url == "false" || url == "" {
		return false
	} else {
		request := gorequest.New()
		if daili != "" {
			request = request.Proxy(daili) //代理
		}
		resp, body, errs := request.Get(url).End()
		if len(errs) > 0 || resp.StatusCode != 200 {
			return false
		}
		if resp.StatusCode == 200 { //403代表开票，普通用户无法查看，只能看大厅表演
			_ = body
			return true
		}
		return false
	}
}

const (
	domain = "zh.stripchat.com"
	cnName = "stripchat"
)

type Live struct {
	internal.BaseLive
	liveroom *configs.LiveRoom
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
	fmt.Println(l.liveroom)
	fmt.Println(l.GetStreamUrls())

	modeName := strings.Split(l.Url.String(), "/")
	modelName := modeName[len(modeName)-1]

	daili := ""
	config, config_err := readconfig.Get_config()
	if config_err != nil {
		daili = ""
	} else {
		daili = config.Proxy
	}
	modelID := get_modelId(modelName, daili)
	m3u8 := get_M3u8(modelID, daili)
	m3u8_status := test_m3u8(m3u8, daili)
	if modelID == "false" {
		return nil, live.ErrRoomUrlIncorrect
	} else if modelID == "url.Error" {
		return nil, live.ErrInternalError
	} else if modelID == "OffLine" {
		info = &live.Info{
			Live:     l,
			RoomName: modelID,
			HostName: modelName,
			Status:   false,
		}
		return info, nil
	} else if m3u8 != "" {
		info = &live.Info{
			Live:     l,
			RoomName: modelID,
			HostName: modelName,
			Status:   m3u8_status,
		}
		return info, nil
	}
	return nil, live.ErrRoomNotExist //live.ErrInternalError
}

func (l *Live) GetStreamUrls() (us []*url.URL, err error) {
	// modeName := regexp.MustCompile(`stripchat.com\/(\w|-)+`).FindString(l.Url.String())
	modeName := strings.Split(l.Url.String(), "/")
	modelName := modeName[len(modeName)-1]
	daili := ""
	config, config_err := readconfig.Get_config()
	if config_err != nil {
		daili = ""
	} else {
		daili = config.Proxy
	}
	modelID := get_modelId(modelName, daili)
	m3u8 := get_M3u8(modelID, daili)
	m3u8_status := test_m3u8(m3u8, daili)
	if m3u8_status {
		return utils.GenUrls(m3u8)
	}
	if modelID == "url.Error" {
		return nil, live.ErrInternalError
	}
	if modelID == "false" || modelID == "OffLine" || m3u8 == "false" || !m3u8_status {
		return nil, err //live.ErrRoomNotExist
	}
	return nil, live.ErrInternalError
}

func (l *Live) GetPlatformCNName() string {
	return cnName
}
