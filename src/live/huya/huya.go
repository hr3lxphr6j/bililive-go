package huya

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/hr3lxphr6j/bililive-go/src/live"
	"github.com/hr3lxphr6j/bililive-go/src/live/internal"
	"github.com/hr3lxphr6j/bililive-go/src/pkg/utils"
	"github.com/hr3lxphr6j/requests"
	"github.com/tidwall/gjson"
)

const (
	domain = "www.huya.com"
	cnName = "虎牙"
	uaApp  = "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148 MicroMessenger/8.0.49(0x18003137) NetType/WIFI Language/zh_CN WeChat/8.0.49.33 CFNetwork/1474 Darwin/23.0.0"
)

func init() {
	live.Register(domain, new(builder))
}

type builder struct{}

func (b *builder) Build(url *url.URL, opt ...live.Option) (live.Live, error) {
	return &Live{
		BaseLive: internal.NewBaseLive(url, opt...),
	}, nil
}

type Live struct {
	internal.BaseLive
}

func (l *Live) GetHtmlBody() (htmlBody string, err error) {
	html, err := requests.Get(l.Url.String(), live.CommonUserAgent)
	if err != nil {
		return
	}
	if html.StatusCode != http.StatusOK {
		err = fmt.Errorf("status code: %d", html.StatusCode)
		return
	}
	htmlBody, err = html.Text()
	return
}

func (l *Live) getDate(htmlBody string) (result *gjson.Result, err error) {
	strFilter := utils.NewStringFilterChain(utils.ParseUnicode, utils.UnescapeHTMLEntity)
	rjson := strFilter.Do(utils.Match1(`stream: (\{"data".*?),"iWebDefaultBitRate"`, htmlBody)) + "}"
	gj := gjson.Parse(rjson)

	roomId := gj.Get("data.0.gameLiveInfo.profileRoom").String()
	params := make(map[string]string)
	params["m"] = "Live"
	params["do"] = "profileRoom"
	params["roomid"] = roomId
	params["showSecret"] = "1"

	headers := make(map[string]interface{})
	headers["User-Agent"] = uaApp
	headers["xweb_xhr"] = "1"
	headers["referer"] = "https://servicewechat.com/wx74767bf0b684f7d3/301/page-frame.html"
	headers["accept-language"] = "zh-CN,zh;q=0.9"
	resp, err := requests.Get("https://mp.huya.com/cache.php", requests.Headers(headers), requests.Queries(params), requests.UserAgent(uaApp))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, live.ErrRoomNotExist
	}
	body, err := resp.Text()
	if err != nil {
		return nil, err
	}
	res := gjson.Parse(body)
	return &res, nil
}

func (l *Live) GetInfo() (info *live.Info, err error) {
	body, err := l.GetHtmlBody()
	if err != nil {
		return nil, err
	}

	if res := utils.Match1("哎呀，虎牙君找不到这个主播，要不搜索看看？", body); res != "" {
		return nil, live.ErrRoomNotExist
	}

	if strings.Contains(body, "该主播涉嫌违规，正在整改中") {
		return &live.Info{
			Live:     l,
			HostName: "该主播涉嫌违规，正在整改中",
			RoomName: "该主播涉嫌违规，正在整改中",
			Status:   false,
		}, nil
	}

	res, err := l.getDate(body)
	if err != nil {
		return nil, err
	}

	if res := utils.Match1("该主播不存在！", res.String()); res != "" {
		return nil, live.ErrRoomNotExist
	}

	var (
		hostName = res.Get("data.liveData.nick").String()
		roomName = res.Get("data.liveData.introduction").String()
		status   = res.Get("data.realLiveStatus").String()
	)

	if hostName == "" || roomName == "" || status == "" {
		return nil, live.ErrInternalError
	}

	info = &live.Info{
		Live:     l,
		HostName: hostName,
		RoomName: roomName,
		Status:   status == "ON",
	}
	return info, nil
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func (l *Live) GetStreamUrls() (us []*url.URL, err error) {
	body, err := l.GetHtmlBody()
	if err != nil {
		return nil, err
	}

	data, err := l.getDate(body)
	if err != nil {
		return nil, err
	}
	sFlvUrl := data.Get("data.stream.baseSteamInfoList.0.sFlvUrl").String()
	sStreamName := data.Get("data.stream.baseSteamInfoList.0.sStreamName").String()
	sFlvUrlSuffix := data.Get("data.stream.baseSteamInfoList.0.sFlvUrlSuffix").String()
	sFlvAntiCode := data.Get("data.stream.baseSteamInfoList.0.sFlvAntiCode").String()
	streamUrl := fmt.Sprintf("%s/%s.%s?%s", sFlvUrl, sStreamName, sFlvUrlSuffix, sFlvAntiCode)

	res, err := utils.GenUrls(streamUrl)
	if err != nil {
		return nil, err
	}
	return res, nil

}

func (l *Live) GetPlatformCNName() string {
	return cnName
}

func (l *Live) GetHeadersForDownloader() map[string]string {
	return map[string]string{
		"User-Agent":      uaApp,
		"Accept":          `text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8`,
		"Accept-Encoding": `gzip, deflate`,
		"Accept-Language": `zh-CN,zh;q=0.8,en-US;q=0.5,en;q=0.3`,
	}
}
