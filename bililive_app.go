package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/hr3lxphr6j/bililive-go/src/configs"
	"github.com/hr3lxphr6j/bililive-go/src/instance"
	"github.com/hr3lxphr6j/bililive-go/src/lib/events"
	"github.com/hr3lxphr6j/bililive-go/src/lib/utils"
	"github.com/hr3lxphr6j/bililive-go/src/listeners"
	"github.com/hr3lxphr6j/bililive-go/src/log"
	"github.com/hr3lxphr6j/bililive-go/src/recorders"
	"os"
	"strings"
)

const (
	AppName     = "BiliLive-go"
	AppVersion  = "0.10"
	CommandName = "bililive-go"
)

func version() {
	fmt.Fprintf(os.Stderr, "%s Version: %s\n", AppName, AppVersion)
}

func help() {
	version()
	fmt.Fprintf(os.Stderr,
		"Usage: %s [-hv] [-i urls] [-o path] [-t seconds] [-c filename]\n\n"+
			"Options:\n"+
			"  -h:\tthis help\n"+
			"  -v:\tshow version and exit\n"+
			"  -i:\tlive room urls, if have many urls, split with \"|\"\n"+
			"  -o:\toutput file path (default: ./)\n"+
			"  -t:\tinterval of query live status (default: 30)\n"+
			"  -c:\tset configuration file, command line options with override this (default: ./config.yml)\n", CommandName)
}

var (
	h bool   // 帮助
	v bool   // 版本信息
	c string // 配置文件
	i string // 直播间urls
	o string // 输出路径
	t int    // 轮训间隔

)

func parse(inst *instance.Instance) {
	flag.BoolVar(&h, "h", false, "show help info")
	flag.BoolVar(&v, "v", false, "show version")
	flag.StringVar(&c, "c", "", "config file")
	flag.StringVar(&i, "i", "", "live room urls, if have many urls, split with \"|\"")
	flag.StringVar(&o, "o", "", "output file path (default: ./)")
	flag.IntVar(&t, "t", -1, "interval of query live status")
	flag.Parse()

	if h {
		help()
		os.Exit(0)
	}
	if v {
		version()
		os.Exit(0)
	}

	if c == "" {
		// 未定义配置文件，尝试解析默认位置
		config, err := configs.NewConfigWithFile("./config.yml")
		if err != nil {
			config = configs.NewConfig()
		}
		inst.Config = config
	} else {
		// 已定义配置文件，若解析出错则报错退出
		config, err := configs.NewConfigWithFile(c)
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error()+"\n")
			os.Exit(1)
		}
		inst.Config = config
	}

	if i != "" {
		for _, u := range strings.Split(i, "|") {
			if u != "" {
				inst.Config.LiveRooms = append(inst.Config.LiveRooms, u)
			}
		}
	}

	if o != "" {
		inst.Config.OutPutPath = o
	}

	if t != -1 {
		inst.Config.Interval = t
	}

	if err := configs.VerifyConfig(inst.Config); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(2)
	}

}

func main() {
	// 判断FFmpeg
	if !utils.IsFFmpegExist() {
		fmt.Fprintf(os.Stderr, "FFmpeg binary not found, Please Check.\n")
		os.Exit(3)
	}

	// 初始化实例
	inst := new(instance.Instance)
	ctx := context.WithValue(context.Background(), instance.InstanceKey, inst)

	// 解析参数和配置
	parse(inst)

	// 初始化组件
	events.NewIEventDispatcher(ctx)
	logger := log.NewLogger(ctx)
	logger.Infof("%s Version: %s Link Start", AppName, AppVersion)
	logger.Debug(inst.Config)

	listeners.NewIListenerManager(ctx)
	recorders.NewIRecorderManager(ctx)

	inst.ListenerManager.Start(ctx)
	inst.RecorderManager.Start(ctx)

	inst.WaitGroup.Wait()
}
