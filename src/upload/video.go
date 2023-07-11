package upload

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/matyle/bililive-go/src/configs"
	"github.com/matyle/bililive-go/src/pkg/zaplogger"
	"go.uber.org/zap"
)

type localVideoPool struct {
	pool *sync.Pool
}

func newLocalVideoPool() *localVideoPool {
	return &localVideoPool{
		pool: &sync.Pool{
			New: func() interface{} {
				return &localVideo{}
			},
		},
	}
}

func (l *localVideoPool) Get() *localVideo {
	return l.pool.Get().(*localVideo)
}

func (l *localVideoPool) Put(video *localVideo) {
	l.pool.Put(video)
}

type localVideo struct {
	videoPath     string
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

func getFileSize(path string) int64 {
	file, err := os.Open(path)
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

type MediaFiles struct {
	folderPath     string              // 文件夹路径
	uploadingVideo map[string]struct{} // 上传中的文件集合
	log            *zap.Logger
}

// 添加上传中的视频
func (v *MediaFiles) AddVideo(videoName string) {
	// v.uploadingVideo[videoName] = struct{}{}
}

func (v *MediaFiles) RemoveVideo(videoName string) {
	delete(v.uploadingVideo, videoName)
}

func (v *MediaFiles) IsUploading(videoName string) bool {
	_, ok := v.uploadingVideo[videoName]
	return ok
}

func (v *MediaFiles) IsEmpty() bool {
	return len(v.uploadingVideo) == 0
}

func (v *MediaFiles) Clear() {
	for k := range v.uploadingVideo {
		delete(v.uploadingVideo, k)
	}
}

func (v *MediaFiles) ScanBiludVideo() {
	// 获取当前文件夹下所有文件和文件夹
	files, err := ioutil.ReadDir(v.folderPath)
	if err != nil {
		fmt.Printf("Failed to read directory: %s\n", err)
		return
	}
	// 遍历文件和文件夹，并筛选后缀为 .mp4 或 .flv 的文件
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		// 检查文件扩展名是否为 .mp4 或 .flv
		extension := strings.ToLower(filepath.Ext(file.Name()))
		if extension == ".mp4" || extension == ".flv" {
			v.log.Info("Find video", zap.String("video", file.Name()))
			v.AddVideo(file.Name())
		}
	}
}

func NewMediaFiles(path string) *MediaFiles {
	logFile := configs.NewConfig().OutPutPath + "/vediofile.log"
	log := zaplogger.GetFileLogger(logFile).With(zap.String("pkg", "upload")).With(zap.String("file", "video.go"))
	return &MediaFiles{
		folderPath:     path,
		uploadingVideo: make(map[string]struct{}, 64),
		log:            log,
	}
}
