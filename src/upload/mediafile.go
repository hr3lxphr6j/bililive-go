package upload

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

type Videos struct {
	folderPath   string              // 文件夹路径
	uploadingSet map[string]struct{} // 上传中的文件集合
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

func (v *Videos) AddVideo(videoName string) {
	v.uploadingSet[videoName] = struct{}{}
}

func (v *Videos) RemoveVideo(videoName string) {
	delete(v.uploadingSet, videoName)
}

func (v *Videos) IsUploading(videoName string) bool {
	_, ok := v.uploadingSet[videoName]
	return ok
}

func (v *Videos) IsEmpty() bool {
	return len(v.uploadingSet) == 0
}

func (v *Videos) Clear() {
	for k := range v.uploadingSet {
		delete(v.uploadingSet, k)
	}
}

func (v *Videos) ScanBiludVideo() {
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
			log.Info("Find video", zap.String("video", file.Name()))
			v.AddVideo(file.Name())
		}
	}
}

func NewVideos(path string) *Videos {
	return &Videos{
		folderPath:   path,
		uploadingSet: make(map[string]struct{}),
	}
}

func StartUpload(videos *Videos) {
	// TODO
}
