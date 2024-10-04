package stripchat

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"reflect"
	"regexp"
	"strings"

	"github.com/hr3lxphr6j/bililive-go/src/cmd/bililive/readconfig"
	"github.com/hr3lxphr6j/bililive-go/src/live"
	"github.com/hr3lxphr6j/bililive-go/src/live/internal"
	"github.com/hr3lxphr6j/bililive-go/src/pkg/utils"
	"github.com/parnurzeal/gorequest"
	"github.com/tidwall/gjson"
)

var (
	ErrFalse                     = errors.New("false")
	ErrModelName                 = errors.New("err model name")
	Err_GetInfo_Unexpected       = errors.New("GetInfo未知错误")
	Err_GetStreamUrls_Unexpected = errors.New("GetStreamUrls未知错误")
	Err_TestUrl_Unexpected       = errors.New("testUrl未知错误")
	ErrOffline                   = errors.New("OffLine")
	// ErrNullUrl                   = errors.New("no url")
)

func get_modelId(modleName string, daili string) (string, error) {
	if modleName == "" {
		return "", ErrFalse
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
				return "", live.ErrInternalError
			} else {
				fmt.Println(reflect.TypeOf(err), "错误详情:", err)
			}
		}
		return "", ErrFalse
	} else {
		// 解析 JSON 响应
		if len(gjson.Get(body, "messages").String()) > 2 {
			modelId := gjson.Get(body, "messages.0.modelId").String()
			return modelId, nil
		} else if len(gjson.Get(body, "messages").String()) == 2 {
			return "", ErrOffline
		} else if len(gjson.Get(body, "messages").String()) == 0 {
			return "", ErrModelName
		}
		return "", ErrFalse
	}
}

func get_M3u8(modelId string, daili string) (string, error) {
	if modelId == "" { // || modelId == "false" || modelId == "OffLine" || modelId == "url.Error" {
		return "", ErrFalse
	}
	// url := "https://edge-hls.doppiocdn.com/hls/" + modelId + "/master/" + modelId + "_auto.m3u8?playlistType=lowLatency"
	urlinput := "https://edge-hls.doppiocdn.com/hls/" + modelId + "/master/" + modelId + "_auto.m3u8?playlistType=standard"
	// url := "https://edge-hls.doppiocdn.com/hls/" + modelId + "/master/" + modelId + ".m3u8"
	request := gorequest.New()
	if daili != "" {
		request = request.Proxy(daili) //代理
	}
	resp, body, errs := request.Get(urlinput).End()
	if errs != nil {
		for _, err := range errs {
			if _, ok := err.(*url.Error); ok {
				return "", live.ErrInternalError
			}
		}
		return "", ErrFalse
	}
	if resp.StatusCode == 404 || resp.StatusCode == 403 {
		return "", ErrOffline
	}
	if resp.StatusCode == 200 {
		// re := regexp.MustCompile(`(https:\/\/[\w\-\.]+\/hls\/[\d]+\/[\d\_p]+\.m3u8\?playlistType=lowLatency)`)
		re := regexp.MustCompile(`(https:\/\/[\w\-\.]+\/hls\/[\d]+\/[\d\_p]+\.m3u8\?playlistType=standard)`) //等价于\?playlistType=standard
		matches := re.FindString(body)
		return matches, nil
	} else {
		return "", ErrFalse
	}
}
func test_m3u8(urlinput string, daili string) (bool, error) {
	if urlinput == "" {
		return false, ErrFalse
	} else {
		request := gorequest.New()
		if daili != "" {
			request = request.Proxy(daili) //代理
		}
		resp, body, errs := request.Get(urlinput).End()
		if errs != nil {
			for _, err := range errs {
				if _, ok := err.(*url.Error); ok {
					return false, live.ErrInternalError
				}
			}
			return false, ErrFalse
		}
		if resp.StatusCode == 200 {
			_ = body
			return true, nil
		}
		if resp.StatusCode == 403 || resp.StatusCode == 404 { //403代表开票，普通用户无法查看，只能看大厅表演
			_ = body
			return false, ErrOffline
		}
		if resp.StatusCode != 200 {
			return false, ErrFalse
		}
		return false, Err_TestUrl_Unexpected
	}
}

const (
	domain = "zh.stripchat.com"
	cnName = "stripchat"
)

type Live struct {
	internal.BaseLive
	// liveroom map[string]*configs.LiveRoom
	m3u8Url string
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
	daili := ""
	config, config_err := readconfig.Get_config()
	if config_err != nil {
		daili = ""
	} else {
		daili = config.Proxy
	}

	modelID, err_getid := get_modelId(modelName, daili)
	m3u8, err_getm3u8 := get_M3u8(modelID, daili)

	if m3u8 == "" && l.m3u8Url != "" { //url
		m3u8 = l.m3u8Url
	}
	m3u8_status, err_testm3u8 := test_m3u8(m3u8, daili)

	if m3u8_status { //strings.Contains(m3u8, ".m3u8")
		l.m3u8Url = m3u8
		info = &live.Info{
			Live:         l,
			RoomName:     modelID,
			HostName:     modelName,
			Status:       true,
			CustomLiveId: m3u8, //l.GetLiveId()可获取持久化数据
		}
		return info, nil
	}
	if errors.Is(err_getid, ErrOffline) || errors.Is(err_getm3u8, ErrOffline) || errors.Is(err_testm3u8, ErrOffline) {
		info = &live.Info{
			Live:     l,
			RoomName: "OffLine",
			HostName: modelName,
			Status:   m3u8_status, //false,
		}
		return info, nil
	}
	// if m3u8 == "" {
	// 	if strings.Contains(l.m3u8Url, ".m3u8") {
	// 		m3u8 = l.m3u8Url
	// 		m3u8_status = test_m3u8(m3u8, daili)
	// 		if m3u8_status != false { //strings.Contains(m3u8, ".m3u8")
	// 			l.m3u8Url = m3u8
	// 			info = &live.Info{
	// 				Live:         l,
	// 				RoomName:     modelID,
	// 				HostName:     modelName,
	// 				Status:       true,
	// 				CustomLiveId: m3u8, //l.GetLiveId()可获取持久化数据
	// 			}
	// 			return info, nil
	// 		}
	// 	} else {
	// 		return nil, err_getid
	// 	}
	// }
	return nil, Err_GetInfo_Unexpected
}

func (l *Live) GetStreamUrls() (us []*url.URL, err error) {
	modeName := strings.Split(l.Url.String(), "/")
	modelName := modeName[len(modeName)-1]
	daili := ""
	m3u8 := ""
	config, config_err := readconfig.Get_config()
	if config_err != nil {
		daili = ""
	} else {
		daili = config.Proxy
	}
	modelID, err := get_modelId(modelName, daili)
	if l.m3u8Url == "" {
		m3u8, err = get_M3u8(modelID, daili)
	} else {
		m3u8 = l.m3u8Url
	}
	if errors.Is(err, live.ErrInternalError) || errors.Is(err, ErrOffline) {
		return nil, live.ErrInternalError
	}
	// fmt.Println("\n l.m3u8Url=", l.m3u8Url, " l.GetLiveId()", string(l.GetLiveId()))
	m3u8_status, err_testm3u8 := test_m3u8(m3u8, daili)
	if m3u8_status {
		return utils.GenUrls(m3u8)
	}

	if !m3u8_status {
		return nil, err_testm3u8
	}

	return nil, Err_GetStreamUrls_Unexpected
}

func (l *Live) GetPlatformCNName() string {
	return cnName
}
