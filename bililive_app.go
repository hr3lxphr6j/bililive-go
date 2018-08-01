package main

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/alecthomas/kingpin"
	"github.com/hr3lxphr6j/bililive-go/src/api"
	"github.com/hr3lxphr6j/bililive-go/src/configs"
	"github.com/hr3lxphr6j/bililive-go/src/consts"
	"github.com/hr3lxphr6j/bililive-go/src/instance"
	"github.com/hr3lxphr6j/bililive-go/src/lib/events"
	"github.com/hr3lxphr6j/bililive-go/src/lib/utils"
	"github.com/hr3lxphr6j/bililive-go/src/listeners"
	"github.com/hr3lxphr6j/bililive-go/src/log"
	"github.com/hr3lxphr6j/bililive-go/src/recorders"
	"github.com/hr3lxphr6j/bililive-go/src/servers"
)

var (
	app      = kingpin.New(consts.AppName, "A command-line live stream save tools.").Version(consts.AppVersion)
	debug    = app.Flag("debug", "Enable debug mode.").Default("false").Bool()
	interval = app.Flag("interval", "Interval of query live status").Default("20").Short('t').Int()
	output   = app.Flag("output", "Output file path.").Short('o').Default("./").String()
	input    = app.Flag("input", "Live room urls").Short('i').Strings()
	conf     = app.Flag("config", "Config file.").Short('c').String()
	rpc      = app.Flag("enable-rpc", "Enable RPC server.").Default("false").Bool()
	rpcAddr  = app.Flag("rpc-addr", "RPC server listen port").Default(":8080").String()
	rpcToken = app.Flag("rpc-token", "RPC server token.").String()
	rpcTLS   = app.Flag("enable-rpc-tls", "Enable TLS for RPC server").Bool()
	certFile = app.Flag("rpc-tls-cert-file", "Cert file for TLS on RPC").String()
	keyFile  = app.Flag("rpc-tls-key-file", "Key file for TLS on RPC").String()
)

func getConfig() (*configs.Config, error) {
	kingpin.MustParse(app.Parse(os.Args[1:]))
	var config *configs.Config
	if *conf != "" {
		if c, err := configs.NewConfigWithFile(*conf); err == nil {
			config = c
		} else {
			return nil, err
		}
	} else {
		config = &configs.Config{
			RPC: configs.RPC{
				Enable: *rpc,
				Port:   *rpcAddr,
				Token:  *rpcToken,
				TLS: configs.TLS{
					Enable:   *rpcTLS,
					CertFile: *certFile,
					KeyFile:  *keyFile,
				},
			},
			Debug:      *debug,
			Interval:   *interval,
			OutPutPath: *output,
			LiveRooms:  *input,
		}
	}
	if err := configs.VerifyConfig(config); err != nil {
		return nil, err
	}
	return config, nil
}

func main() {
	// 判断FFmpeg
	if !utils.IsFFmpegExist() {
		fmt.Fprintf(os.Stderr, "FFmpeg binary not found, Please Check.\n")
		os.Exit(3)
	}
	config, err := getConfig()
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
	}
	// 初始化实例
	inst := new(instance.Instance)
	inst.Config = config
	ctx := context.WithValue(context.Background(), instance.InstanceKey, inst)
	logger := log.NewLogger(ctx)
	logger.Infof("%s Version: %s Link Start", consts.AppName, consts.AppVersion)
	logger.Debugf("%+v", consts.AppInfo)
	logger.Debugf("%+v", inst.Config)

	// 初始化事件分发模块
	events.NewIEventDispatcher(ctx)
	// 初始化直播间记录
	inst.Lives = make(map[api.LiveId]api.Live)
	// 从配置添加直播间
	for _, room := range inst.Config.LiveRooms {
		u, err := url.Parse(room)
		if err != nil {
			logger.WithField("url", room).Error(err)
		}
		if l, err := api.NewLive(u); err == nil {
			if _, ok := inst.Lives[l.GetLiveId()]; ok {
				logger.Errorf("%s is exist!", room)
			} else {
				inst.Lives[l.GetLiveId()] = l
			}
		} else {
			logger.WithField("url", room).Error(err.Error())
		}
	}
	// 初始化RPC
	if inst.Config.RPC.Enable {
		servers.NewServer(ctx).Start(ctx)
	}
	// 初始化监听、录制
	lm := listeners.NewIListenerManager(ctx)
	recorders.NewIRecorderManager(ctx)
	// 启动监听模块
	inst.ListenerManager.Start(ctx)
	inst.RecorderManager.Start(ctx)

	// 添加现有直播间监听
	for _, live := range inst.Lives {
		if err := lm.AddListener(ctx, live); err != nil {
			logger.WithFields(map[string]interface{}{"url": live.GetRawUrl()}).Error(err)
		}
	}
	inst.WaitGroup.Wait()
	logger.Info("Bye~")
}
