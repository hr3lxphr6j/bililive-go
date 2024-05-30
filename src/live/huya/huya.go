package huya

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/hr3lxphr6j/requests"
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

type urlQueryParams struct {
	WsSecret string
	WsTime   string
	Seqid    string
	Ctype    string
	Ver      string
	Fs       string
	U        string
	T        string
	Sv       string
	Sdk_sid  string
	Codec    string
}

func parseAntiCode(anticode string, uid int64, streamName string) (string, error) {
	qr, err := url.ParseQuery(anticode)
	if err != nil {
		return "", err
	}
	resultTemplate := template.Must(template.New("urlQuery").Parse(
		"wsSecret={{.WsSecret}}" +
			"&wsTime={{.WsTime}}" +
			"&seqid={{.Seqid}}" +
			"&ctype={{.Ctype}}" +
			"&ver={{.Ver}}" +
			"&fs={{.Fs}}" +
			"&u={{.U}}" +
			"&t={{.T}}" +
			"&sv={{.Sv}}" +
			"&sdk_sid={{.Sdk_sid}}" +
			"&codec={{.Codec}}",
	))
	timeNow := time.Now().Unix() * 1000
	resultParams := urlQueryParams{
		WsSecret: "",
		WsTime:   qr.Get("wsTime"),
		Seqid:    strconv.FormatInt(timeNow+uid, 10),
		Ctype:    qr.Get("ctype"),
		Ver:      "1",
		Fs:       qr.Get("fs"),
		U:        strconv.FormatInt(uid, 10),
		T:        "100",
		Sv:       "2405220949",
		Sdk_sid:  strconv.FormatInt(uid, 10),
		Codec:    "264",
	}
	ss := GetMD5Hash(fmt.Sprintf("%s|%s|%s", resultParams.Seqid, resultParams.Ctype, resultParams.T))

	decodeString, _ := base64.StdEncoding.DecodeString(qr.Get("fm"))
	fm := string(decodeString)
	fm = strings.ReplaceAll(fm, "$0", resultParams.U)
	fm = strings.ReplaceAll(fm, "$1", streamName)
	fm = strings.ReplaceAll(fm, "$2", ss)
	fm = strings.ReplaceAll(fm, "$3", resultParams.WsTime)

	resultParams.WsSecret = GetMD5Hash(fm)
	var buf bytes.Buffer
	if err := resultTemplate.Execute(&buf, resultParams); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (l *Live) GetStreamUrls() (us []*url.URL, err error) {
	resp, err := requests.Get(l.Url.String(), requests.UserAgent(userAgent))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}
	body, err := resp.Text()
	if err != nil {
		return nil, err
	}

	tmpStrings := strings.Split(body, `stream: `)
	if len(tmpStrings) < 2 {
		return nil, fmt.Errorf("stream not found")
	}
	streamJsonRawString := strings.Split(tmpStrings[1], `};`)[0]
	if !gjson.Valid(streamJsonRawString) {
		return nil, fmt.Errorf("streamJsonRawString not valid")
	}
	streamJson := gjson.Parse(streamJsonRawString)
	vMultiStreamInfoJson := streamJson.Get("vMultiStreamInfo").Array()
	if len(vMultiStreamInfoJson) == 0 {
		return nil, fmt.Errorf("vMultiStreamInfo not found")
	}

	streamInfoJsons := streamJson.Get("data.0.gameStreamInfoList").Array()
	index := l.lastCdnIndex
	if index >= len(streamInfoJsons) {
		index = 0
	}
	l.lastCdnIndex = index + 1
	gameStreamInfoJson := streamInfoJsons[index]
	return getStreamUrlsFromGameStreamInfoJson(gameStreamInfoJson)
}

func getStreamUrlsFromGameStreamInfoJson(gameStreamInfoJson gjson.Result) (us []*url.URL, err error) {
	// get streamName
	sStreamName := gameStreamInfoJson.Get("sStreamName").String()
	// get sFlvAntiCode
	sFlvAntiCode := gameStreamInfoJson.Get("sFlvAntiCode").String()
	// get sFlvUrl
	sFlvUrl := gameStreamInfoJson.Get("sFlvUrl").String()
	// get random uid
	uid := rand.Int63n(99999999999) + 1200000000000

	query, err := parseAntiCode(sFlvAntiCode, uid, sStreamName)
	if err != nil {
		return nil, err
	}
	tmpUrlString := fmt.Sprintf("%s/%s.flv?%s", sFlvUrl, sStreamName, query)
	u, err := url.Parse(tmpUrlString)
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
