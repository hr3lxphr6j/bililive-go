package upload

import (
	"testing"
)

func TestBiliUPload(t *testing.T) {
	files, err := ScanVideoFiles("./")
	if err != nil {
		t.Error(err)
	}
	//上传文件

}

func ScanVideoFiles(path string) (files []string, err error) {
	//扫描文件夹中所有后缀为mp4,flv,avi的文件
}
