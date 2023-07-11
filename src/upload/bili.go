package upload

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/imroc/req/v3"
	"github.com/matyle/bililive-go/src/configs"
	"github.com/panjf2000/ants/v2"
	"github.com/schollz/progressbar/v3"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

type BiliUpload struct {
	cookie string
	csrf   string
	client *req.Client

	title string

	threadNum int
	partChan  chan Part
	chunks    int64

	config *configs.BiliupConfig
}

type BiliUploads struct {
	BiliUploads []*BiliUpload
	Configs     []*configs.BiliupConfig
}

var wg sync.WaitGroup

// 支持上传到多个 bilibili 账号
func NewBiliUPLoads(configs []*configs.BiliupConfig, threadNum int) *BiliUploads {
	if len(configs) == 0 {
		panic("cookie文件不存在,请先登录")
	}
	var biliUploads []*BiliUpload
	for _, v := range configs {
		biliUploads = append(biliUploads, newBiliUPLoad(v, threadNum))
	}
	return &BiliUploads{
		BiliUploads: biliUploads,
		Configs:     configs,
	}
}

// 上传视频成功之后，可以删除本地视频
func (u *BiliUploads) Upload(postUploadHandler func()) {
	for i, v := range u.BiliUploads {
		wg.Add(1)
		go func(i int, v *BiliUpload) {
			defer wg.Done()
			log.Info("开始上传",
				zap.Int("第一个用户", i),
				zap.String("用户名", u.Configs[i].UserName))
			v.Upload()
		}(i, v)
	}
	wg.Wait()
	log.Info("全部上传完成，开始执行后续操作")
	if postUploadHandler != nil {
		postUploadHandler()
	}
}

func newBiliUPLoad(config *configs.BiliupConfig, threadNum int) *BiliUpload {
	if config.CookiePath == "" {
		panic("cookie文件不存在,请先登录")
	}
	var cookieinfo BiliCookie
	loginInfo, err := os.ReadFile(config.CookiePath)
	if err != nil || len(loginInfo) == 0 {
		panic("cookie文件不存在,请先登录")
	}
	_ = json.Unmarshal(loginInfo, &cookieinfo)
	var cookie string
	var csrf string
	for _, v := range cookieinfo.Data.CookieInfo.Cookies {
		cookie += v.Name + "=" + v.Value + ";"
		if v.Name == "bili_jct" {
			csrf = v.Value
		}
	}
	var client = req.C().SetCommonHeaders(map[string]string{
		"user-agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36 Edg/105.0.1343.53",
		"cookie":     cookie,
		"Connection": "keep-alive",
	})
	resp, _ := client.R().Get("https://api.bilibili.com/x/web-interface/nav")
	uname := gjson.ParseBytes(resp.Bytes()).Get("data.uname").String()
	if uname == "" {
		panic("cookie失效,请重新登录")
	}
	// log.Printf("%s 登录成功", uname)
	log.Info("登录成功", zap.String("uname", uname))
	return &BiliUpload{
		cookie:    cookie,
		csrf:      csrf,
		client:    client,
		upVideo:   &UpVideo{},
		threadNum: threadNum,
		config:    config,
	}
}

func (u *BiliUpload) SetVideos(videoPath string) *BiliUpload {
	u.upVideo.videoName = path.Base(videoPath)
	u.upVideo.videoSize = u.getVideoSize(videoPath)
	u.upVideo.coverUrl = u.uploadCover(u.config.CoverPath)
	return u
}

func (u *BiliUpload) getVideoSize(videoPath string) int64 {
	file, err := os.Open(videoPath)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		panic(err)
	}
	return fileInfo.Size()
}

func (u *BiliUpload) uploadCover(path string) string {
	if path == "" {
		return ""
	}
	bytes, err := os.ReadFile(path)
	if err != nil {
		log.Fatal("读取封面失败", zap.Error(err))
	}
	var base64Encoding string
	mimeType := http.DetectContentType(bytes)
	switch mimeType {
	case "image/jpeg", "image/jpg":
		base64Encoding = "data:image/jpeg;base64,"
	case "image/png":
		base64Encoding = "data:image/png;base64,"
	case "image/gif":
		base64Encoding = "data:image/gif;base64,"
	default:
		log.Fatal("不支持的图片格式")
	}
	base64Encoding += base64.StdEncoding.EncodeToString(bytes)
	var coverinfo CoverInfo
	u.client.R().SetFormDataFromValues(url.Values{
		"cover": {base64Encoding},
		"csrf":  {u.csrf},
	}).SetResult(&coverinfo).Post("https://member.bilibili.com/x/vu/web/cover/up")
	return coverinfo.Data.Url
}

func (u *BiliUpload) Upload(upVideo *UpVideo) error {
	var preupinfo PreUpInfo
	u.client.R().SetQueryParams(map[string]string{
		"probe_version": "20211012",
		"upcdn":         "bda2",
		"zone":          "cs",
		"name":          u.upVideo.videoName,
		"r":             "upos",
		"profile":       "ugcfx/bup",
		"ssl":           "0",
		"version":       "2.10.4.0",
		"build":         "2100400",
		"size":          strconv.FormatInt(u.upVideo.videoSize, 10),
		"webVersion":    "2.0.0",
	}).SetResult(&preupinfo).Get("https://member.bilibili.com/preupload")
	u.upVideo.uploadBaseUrl = fmt.Sprintf("https:%s/%s", preupinfo.Endpoint, strings.Split(preupinfo.UposUri, "//")[1])
	u.upVideo.biliFileName = strings.Split(strings.Split(strings.Split(preupinfo.UposUri, "//")[1], "/")[1], ".")[0]
	u.upVideo.chunkSize = preupinfo.ChunkSize
	u.upVideo.auth = preupinfo.Auth
	u.upVideo.bizId = preupinfo.BizId
	u.upload(upVideo)

	var addreq = BiliReq{
		Copyright:    u.config.UpType,
		Cover:        u.upVideo.coverUrl,
		Title:        u.title,
		Tid:          u.config.Tid,
		Tag:          u.config.Tag,
		DescFormatId: 16,
		Desc:         u.config.VideoDesc,
		Source:       u.config.Source,
		Dynamic:      "",
		Interactive:  0,
		Videos: []Video{
			{
				Filename: u.upVideo.biliFileName,
				Title:    u.upVideo.videoName,
				Desc:     "",
				Cid:      preupinfo.BizId,
			},
		},
		ActReserveCreate: 0,
		NoDisturbance:    0,
		NoReprint:        1,
		Subtitle: Subtitle{
			Open: 0,
			Lan:  "",
		},
		Dolby:         0,
		LosslessMusic: 0,
		Csrf:          u.csrf,
	}
	_ = addreq
	resp, err := u.client.R().SetQueryParams(map[string]string{
		"csrf": u.csrf,
	}).SetBodyJsonMarshal(addreq).Post("https://member.bilibili.com/x/vu/web/add/v3")
	log.Debug("resp", zap.String("resp", resp.String()))
	return err
}

func (u *BiliUpload) upload(upVideo *UpVideo) {
	defer ants.Release()
	var upinfo UpInfo
	u.client.SetCommonHeader(
		"X-Upos-Auth", u.upVideo.auth).R().
		SetQueryParams(map[string]string{
			"uploads":       "",
			"output":        "json",
			"profile":       "ugcfx/bup",
			"filesize":      strconv.FormatInt(u.upVideo.videoSize, 10),
			"partsize":      strconv.FormatInt(u.upVideo.chunkSize, 10),
			"biz_id":        strconv.FormatInt(u.upVideo.bizId, 10),
			"meta_upos_uri": u.getMetaUposUri(),
		}).SetResult(&upinfo).Post(u.upVideo.uploadBaseUrl)
	u.upVideo.uploadId = upinfo.UploadId
	u.chunks = int64(math.Ceil(float64(u.upVideo.videoSize) / float64(u.upVideo.chunkSize)))
	var reqjson = new(ReqJson)
	file, _ := os.Open(u.videosPath)
	defer file.Close()
	chunk := 0
	start := 0
	end := 0
	bar := progressbar.NewOptions(int(u.upVideo.videoSize/1024/1024),
		progressbar.OptionSetWriter(os.Stdout),
		progressbar.OptionSetItsString("MB"),
		progressbar.OptionSetDescription("视频上传中..."),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowIts(),
	)
	u.partChan = make(chan Part, u.chunks)
	go func() {
		for p := range u.partChan {
			reqjson.Parts = append(reqjson.Parts, p)
		}
	}()
	p, _ := ants.NewPool(u.threadNum)
	defer p.Release()
	for {
		buf := make([]byte, u.upVideo.chunkSize)
		size, err := file.Read(buf)
		if err != nil && err != io.EOF {
			break
		}
		buf = buf[:size]
		if size > 0 {
			wg.Add(1)
			end += size
			_ = p.Submit(u.uploadPartWrapper(chunk, start, end, size, buf, bar))
			buf = nil
			start += size
			chunk++
		}
		if err == io.EOF {
			break
		}
	}
	wg.Wait()
	close(u.partChan)
	jsonString, _ := json.Marshal(&reqjson)
	u.client.R().SetHeaders(map[string]string{
		"Content-Type": "application/json",
		"Origin":       "https://member.bilibili.com",
		"Referer":      "https://member.bilibili.com/",
	}).SetQueryParams(map[string]string{
		"output":   "json",
		"profile":  "ugcfx/bup",
		"name":     u.upVideo.videoName,
		"uploadId": u.upVideo.uploadId,
		"biz_id":   strconv.FormatInt(u.upVideo.bizId, 10),
	}).SetBodyString(string(jsonString)).SetResult(&upinfo).SetRetryCount(5).AddRetryHook(func(resp *req.Response, err error) {
		log.Debug("重试发送分片确认请求")
		return
	}).
		AddRetryCondition(func(resp *req.Response, err error) bool {
			return err != nil || resp.StatusCode != 200
		}).Post(u.upVideo.uploadBaseUrl)
}

type taskFunc func()

func (u *BiliUpload) uploadPartWrapper(chunk int, start, end, size int, buf []byte, bar *progressbar.ProgressBar) taskFunc {
	return func() {
		defer wg.Done()
		resp, _ := u.client.R().SetHeaders(map[string]string{
			"Content-Type":   "application/octet-stream",
			"Content-Length": strconv.Itoa(size),
		}).SetQueryParams(map[string]string{
			"partNumber": strconv.Itoa(chunk + 1),
			"uploadId":   u.upVideo.uploadId,
			"chunk":      strconv.Itoa(chunk),
			"chunks":     strconv.Itoa(int(u.chunks)),
			"size":       strconv.Itoa(size),
			"start":      strconv.Itoa(start),
			"end":        strconv.Itoa(end),
			"total":      strconv.FormatInt(u.upVideo.videoSize, 10),
		}).SetBodyBytes(buf).SetRetryCount(5).AddRetryHook(func(resp *req.Response, err error) {
			// log.Println("重试发送分片", chunk)
			log.Debug("uploadPartWrapper",
				zap.Int("重试发送分片", chunk))
			return
		}).
			AddRetryCondition(func(resp *req.Response, err error) bool {
				return err != nil || resp.StatusCode != 200
			}).Put(u.upVideo.uploadBaseUrl)
		bar.Add(len(buf) / 1024 / 1024)
		if resp.StatusCode != 200 {
			// log.Println("分片", chunk, "上传失败", resp.StatusCode, "size", size)
			log.Error("uploadPartWrapper",
				zap.Int("分片", chunk),
				zap.Int("StatusCode", resp.StatusCode),
				zap.Int("size", size),
				zap.Int("start", start),
				zap.Int("end", end))
		}
		u.partChan <- Part{
			PartNumber: int64(chunk + 1),
			ETag:       "etag",
		}
	}
}

func (u *BiliUpload) getMetaUposUri() string {
	var metaUposUri PreUpInfo
	u.client.R().SetQueryParams(map[string]string{
		"name":       "file_meta.txt",
		"size":       "2000",
		"r":          "upos",
		"profile":    "fxmeta/bup",
		"ssl":        "0",
		"version":    "2.10.4",
		"build":      "2100400",
		"webVersion": "2.0.0",
	}).SetResult(&metaUposUri).Get("https://member.bilibili.com/preupload")
	return metaUposUri.UposUri
}
