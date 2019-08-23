package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/bluele/gcache"

	_ "github.com/hr3lxphr6j/bililive-go/src/cmd/bililive/internal" // load all live plugins
	"github.com/hr3lxphr6j/bililive-go/src/cmd/bililive/internal/flag"
	"github.com/hr3lxphr6j/bililive-go/src/configs"
	"github.com/hr3lxphr6j/bililive-go/src/consts"
	"github.com/hr3lxphr6j/bililive-go/src/instance"
	"github.com/hr3lxphr6j/bililive-go/src/lib/events"
	"github.com/hr3lxphr6j/bililive-go/src/lib/utils"
	"github.com/hr3lxphr6j/bililive-go/src/listeners"
	"github.com/hr3lxphr6j/bililive-go/src/live"
	"github.com/hr3lxphr6j/bililive-go/src/log"
	"github.com/hr3lxphr6j/bililive-go/src/recorders"
	"github.com/hr3lxphr6j/bililive-go/src/servers"
)

func getConfig() (*configs.Config, error) {
	var config *configs.Config
	if *flag.Conf != "" {
		if c, err := configs.NewConfigWithFile(*flag.Conf); err == nil {
			config = c
		} else {
			return nil, err
		}
	} else {
		config = &configs.Config{
			RPC: configs.RPC{
				Enable: *flag.Rpc,
				Port:   *flag.RpcAddr,
				Token:  *flag.RpcToken,
				TLS: configs.TLS{
					Enable:   *flag.RpcTLS,
					CertFile: *flag.CertFile,
					KeyFile:  *flag.KeyFile,
				},
			},
			Debug:      *flag.Debug,
			Interval:   *flag.Interval,
			OutPutPath: *flag.Output,
			LiveRooms:  *flag.Input,
		}
	}
	if err := configs.VerifyConfig(config); err != nil {
		return nil, err
	}
	return config, nil
}

func main() {
	if !utils.IsFFmpegExist() {
		fmt.Fprintf(os.Stderr, "FFmpeg binary not found, Please Check.\n")
		os.Exit(3)
	}

	config, err := getConfig()
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
	}

	inst := new(instance.Instance)
	inst.Config = config
	inst.Cache = gcache.New(128).LRU().Build()
	ctx := context.WithValue(context.Background(), instance.InstanceKey, inst)

	logger := log.New(ctx)
	logger.Infof("%s Version: %s Link Start", consts.AppName, consts.AppVersion)
	logger.Debugf("%+v", consts.AppInfo)
	logger.Debugf("%+v", inst.Config)

	events.NewIEventDispatcher(ctx)

	inst.Lives = make(map[live.ID]live.Live)
	for _, room := range inst.Config.LiveRooms {
		u, err := url.Parse(room)
		if err != nil {
			logger.WithField("url", room).Error(err)
		}
		if l, err := live.NewLive(u); err == nil {
			if _, ok := inst.Lives[l.GetLiveId()]; ok {
				logger.Errorf("%s is exist!", room)
			} else {
				inst.Lives[l.GetLiveId()] = l
			}
		} else {
			logger.WithField("url", room).Error(err.Error())
		}
	}

	if inst.Config.RPC.Enable {
		servers.NewServer(ctx).Start(ctx)
	}
	lm := listeners.NewIListenerManager(ctx)
	recorders.NewIRecorderManager(ctx)
	inst.ListenerManager.Start(ctx)
	inst.RecorderManager.Start(ctx)

	for _, _live := range inst.Lives {
		if err := lm.AddListener(ctx, _live); err != nil {
			logger.WithFields(map[string]interface{}{"url": _live.GetRawUrl()}).Error(err)
		}
	}

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-c
		if inst.Config.RPC.Enable {
			inst.Server.Close(ctx)
		}
		inst.ListenerManager.Close(ctx)
		inst.RecorderManager.Close(ctx)
	}()

	inst.WaitGroup.Wait()
	logger.Info("Bye~")
}
