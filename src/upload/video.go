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

var videoPool = newLocalVideoPool()

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
	videoFilePath string
	videoSize     int64
	fileName      string
	coverUrl      string
	auth          string
	uploadBaseUrl string
	biliFileName  string
	uploadId      string
	chunkSize     int64
	bizId         int64
}

func newLocalVideo(videoFilePath string) *localVideo {
	video := videoPool.Get()
	video.videoFilePath = videoFilePath
	video.videoSize = getFileSize(videoFilePath)
	fileName := filepath.Base(videoFilePath)
	video.fileName = fileName[:strings.LastIndex(fileName, ".")]
	return video
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
	folderPath     string                 // 文件夹路径
	uploadingVideo map[string]*localVideo // 上传中的文件集合
	successVideo   map[string]struct{}    // 上传成功的文件集合
	log            *zap.Logger
}

// 添加上传中的视频
func (v *MediaFiles) AddVideo(videoName string) {
	video := newLocalVideo(videoName)
	v.uploadingVideo[videoName] = video
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
		videoPool.Put(v.uploadingVideo[k])
		delete(v.uploadingVideo, k)
		delete(v.successVideo, k)
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
			fileName := v.folderPath + file.Name()
			v.log.Info("Find video", zap.String("video", fileName))
			fmt.Println("Find video", fileName)
			v.AddVideo(fileName)
		}
	}
	v.log.Info("Scan video finished")
}

func NewMediaFiles(path string) *MediaFiles {
	logFile := configs.NewConfig().OutPutPath + "vediofile.log"
	fmt.Println("logFile:", logFile)
	log := zaplogger.GetFileLogger(logFile).With(zap.String("pkg", "upload")).With(zap.String("file", "video.go"))
	return &MediaFiles{
		folderPath:     path,
		uploadingVideo: make(map[string]*localVideo, 64),
		log:            log,
	}
}
