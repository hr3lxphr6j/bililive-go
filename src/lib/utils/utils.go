package utils

import "os/exec"

func IsFFmpegExist() bool {
	_, ok := exec.LookPath("ffmpeg")
	return ok == nil
}
