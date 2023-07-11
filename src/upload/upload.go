package upload

import "github.com/matyle/bililive-go/src/pkg/zaplogger"

var (
	log = zaplogger.GetLogger().With("pkg", "upload")
)

type Uploader interface {
	Upload() error
}

type UpVideo struct {
	videoSize     int64
	videoName     string
	coverUrl      string
	auth          string
	uploadBaseUrl string
	biliFileName  string
	uploadId      string
	chunkSize     int64
	bizId         int64
}
