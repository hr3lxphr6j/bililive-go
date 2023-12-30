package huya

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/hr3lxphr6j/requests"
	uuid "github.com/satori/go.uuid"
	"github.com/tidwall/gjson"

	"github.com/hr3lxphr6j/bililive-go/src/live"
	"github.com/hr3lxphr6j/bililive-go/src/live/internal"
	"github.com/hr3lxphr6j/bililive-go/src/pkg/utils"
)

const (
	domain    = "www.huya.com"
	cnName    = "虎牙"
	userAgent = "Mozilla/5.0 (iPhone; CPU iPhone OS 16_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.6 Mobile/15E148 Safari/604.1"
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
	lastCdnIndex int
}

func (l *Live) GetInfo() (info *live.Info, err error) {
	resp, err := requests.Get(l.Url.String(), live.CommonUserAgent)
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

	var (
		strFilter = utils.NewStringFilterChain(utils.ParseUnicode, utils.UnescapeHTMLEntity)
		hostName  = strFilter.Do(utils.Match1(`"nick":"([^"]*)"`, body))
		roomName  = strFilter.Do(utils.Match1(`"introduction":"([^"]*)"`, body))
		status    = strFilter.Do(utils.Match1(`"isOn":([^,]*),`, body))
	)

	if hostName == "" || roomName == "" || status == "" {
		return nil, live.ErrInternalError
	}

	info = &live.Info{
		Live:     l,
		HostName: hostName,
		RoomName: roomName,
		Status:   status == "true",
	}
	return info, nil
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func parseAntiCode(anticode string, uid int64, streamName string) (string, error) {
	qr, err := url.ParseQuery(anticode)
	if err != nil {
		return "", err
	}
	qr.Set("ver", "1")
	qr.Set("sv", "2110211124")
	qr.Set("seqid", strconv.FormatInt(time.Now().Unix()*1000+uid, 10))
	qr.Set("uid", strconv.FormatInt(uid, 10))
	uuid, _ := uuid.NewV4()
	qr.Set("uuid", uuid.String())
	ss := GetMD5Hash(fmt.Sprintf("%s|%s|%s", qr.Get("seqid"), qr.Get("ctype"), qr.Get("t")))
	wsTime := strconv.FormatInt(time.Now().Add(6*time.Hour).Unix(), 16)

	decodeString, _ := base64.StdEncoding.DecodeString(qr.Get("fm"))
	fm := string(decodeString)
	fm = strings.ReplaceAll(fm, "$0", qr.Get("uid"))
	fm = strings.ReplaceAll(fm, "$1", streamName)
	fm = strings.ReplaceAll(fm, "$2", ss)
	fm = strings.ReplaceAll(fm, "$3", wsTime)

	qr.Set("wsSecret", GetMD5Hash(fm))
	qr.Set("ratio", "0")
	qr.Set("wsTime", wsTime)
	return qr.Encode(), nil
}

func (l *Live) GetStreamUrls() (us []*url.URL, err error) {
	roomId := strings.Split(strings.Split(l.Url.Path, "/")[1], "?")[0]
	mobileUrl := fmt.Sprintf("https://m.huya.com/%s", roomId)
	resp, err := requests.Get(mobileUrl, requests.UserAgent(userAgent))
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

	tmpStrings := strings.Split(body, `"tLiveInfo":`)
	if len(tmpStrings) < 2 {
		return nil, fmt.Errorf("tLiveInfo not found")
	}
	liveInfoJsonRawString := strings.Split(tmpStrings[1], `,"_classname":"LiveRoom.LiveInfo"}`)[0] + "}"
	if !gjson.Valid(liveInfoJsonRawString) {
		return nil, fmt.Errorf("liveInfoJsonRawString not valid")
	}
	liveInfoJson := gjson.Parse(liveInfoJsonRawString)

	streamInfoJsons := liveInfoJson.Get("tLiveStreamInfo.vStreamInfo.value").Array()
	if len(streamInfoJsons) == 0 {
		return nil, fmt.Errorf("streamInfoJsons not found")
	}

	index := l.lastCdnIndex + 1
	if index >= len(streamInfoJsons) {
		index = 0
	}
	l.lastCdnIndex = index
	gameStreamInfo := streamInfoJsons[index]
	// get streamName
	sStreamName := gameStreamInfo.Get("sStreamName").String()
	// get sFlvAntiCode
	sFlvAntiCode := gameStreamInfo.Get("sFlvAntiCode").String()
	// get sFlvUrl
	sFlvUrl := gameStreamInfo.Get("sFlvUrl").String()
	// get random uid
	uid := rand.Int63n(99999999999) + 1400000000000

	query, err := parseAntiCode(sFlvAntiCode, uid, sStreamName)
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(fmt.Sprintf("%s/%s.flv?%s", sFlvUrl, sStreamName, query))
	if err != nil {
		return nil, err
	}
	return []*url.URL{u}, nil
}

func (l *Live) GetPlatformCNName() string {
	return cnName
}

func (l *Live) GetHeadersForDownloader() map[string]string {
	return map[string]string{
		"User-Agent":      userAgent,
		"Accept":          `text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8`,
		"Accept-Encoding": `gzip, deflate`,
		"Accept-Language": `zh-CN,zh;q=0.8,en-US;q=0.5,en;q=0.3`,
	}
}
