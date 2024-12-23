package consts

import (
	"fmt"
	"os"
	"runtime"

	"github.com/shirou/gopsutil/disk"
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
	currentPath, _ := os.Getwd()
	usage, err := disk.Usage(currentPath)
	if err != nil {
		// fmt.Println("Error:", err)
		return ("err")
	}
	// 输出硬盘使用情况
	// 将字节转换为 GB
	// totalGB := float64(usage.Total) / (1024 * 1024 * 1024)
	// usedGB := float64(usage.Used) / (1024 * 1024 * 1024)
	freeGB := float64(usage.Free) / (1024 * 1024 * 1024)

	// 输出硬盘使用情况
	// fmt.Printf("Total: %.2f GB\n", totalGB)
	// fmt.Printf("Used: %.2f GB\n", usedGB)
	result := fmt.Sprintf("剩余空间: %.2f GB\n", freeGB)
	// fmt.Printf("Used Percent: %.2f%%\n", usage.UsedPercent)
	return result
}
