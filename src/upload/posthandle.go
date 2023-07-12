package upload

import (
	"os"

	"go.uber.org/zap"
)

func RemoveFilesHandler(u *BiliUploads) {
	for file := range u.files.successVideo {
		u.log.Info("删除文件", zap.String("file", file))
		err := os.Remove(file)
		if err != nil {
			u.log.Error("删除文件失败", zap.String("file", file), zap.Error(err))
		}
	}
}
