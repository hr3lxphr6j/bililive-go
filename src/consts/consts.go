package consts

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const (
	AppName = "BiliLive-go"
)

type Info struct {
	AppName          string `json:"app_name"`
	AppVersion       string `json:"app_version"`
	BuildTime        string `json:"build_time"`
	GitHash          string `json:"git_hash"`
	Pid              int    `json:"pid"`
	Platform         string `json:"platform"`
	GoVersion        string `json:"go_version"`
	CurrentDiskSpace string `json:"current_diskspace"`
}

var (
	BuildTime  string
	AppVersion string
	GitHash    string
	AppInfo    = Info{
		AppName:          AppName,
		AppVersion:       AppVersion,
		BuildTime:        BuildTime,
		GitHash:          GitHash,
		Pid:              os.Getpid(),
		Platform:         fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		GoVersion:        runtime.Version(),
		CurrentDiskSpace: getDir(),
	}
)

func getDir() string {
	dir, err := os.Getwd()
	if err != nil {
		// fmt.Println("Error:", err)
		return ("err")
	}

	// 获取磁盘驱动器（适用于 Windows）
	drive := filepath.VolumeName(dir)

	// 输出结果
	// fmt.Println("Current working directory:", dir)
	// fmt.Println("Current disk:", drive)
	return drive
}
