package upload

import (
	"testing"
)

func TestMediaFiles_ScanBiludVideo(t *testing.T) {
	mf := NewMediaFiles("./test")
	mf.ScanBiludVideo()
}
