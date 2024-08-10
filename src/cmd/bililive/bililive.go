package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/bluele/gcache"

	_ "github.com/hr3lxphr6j/bililive-go/src/cmd/bililive/internal"
	"github.com/hr3lxphr6j/bililive-go/src/cmd/bililive/internal/flag"
	"github.com/hr3lxphr6j/bililive-go/src/configs"
	"github.com/hr3lxphr6j/bililive-go/src/consts"
	"github.com/hr3lxphr6j/bililive-go/src/instance"
	"github.com/hr3lxphr6j/bililive-go/src/listeners"
	"github.com/hr3lxphr6j/bililive-go/src/live"
	"github.com/hr3lxphr6j/bililive-go/src/log"
	"github.com/hr3lxphr6j/bililive-go/src/metrics"
	"github.com/hr3lxphr6j/bililive-go/src/pkg/events"
	"github.com/hr3lxphr6j/bililive-go/src/pkg/utils"
	"github.com/hr3lxphr6j/bililive-go/src/recorders"
	"github.com/hr3lxphr6j/bililive-go/src/servers"
)

func getConfig() (*configs.Config, error) {
	var config *configs.Config
	if *flag.Conf != "" {
		c, err := configs.NewConfigWithFile(*flag.Conf)
		if err != nil {
			return nil, err
		}
		config = c
	} else {
		config = flag.GenConfigFromFlags()
	}
	if !config.RPC.Enable && len(config.LiveRooms) == 0 {
		// if config is invalid, try using the config.yml file besides the executable file.
		config, err := getConfigBesidesExecutable()
		if err == nil {
			return config, config.Verify()
		}
	}
	return config, config.Verify()
}

func getConfigBesidesExecutable() (*configs.Config, error) {
	exePath, err := os.Executable()
	if err != nil {
		return nil, err
	}
	configPath := filepath.Join(filepath.Dir(exePath), "config.yml")
	config, err := configs.NewConfigWithFile(configPath)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func main() {
	config, err := getConfig()
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
	}

	inst := new(instance.Instance)
	inst.Config = config
	// TODO: Replace gcache with hashmap.
	// LRU seems not necessary here.
	inst.Cache = gcache.New(1024).LRU().Build()
	ctx := context.WithValue(context.Background(), instance.Key, inst)

	logger := log.New(ctx)
	logger.Infof("%s Version: %s Link Start", consts.AppName, consts.AppVersion)
	if config.File != "" {
		logger.Debugf("config path: %s.", config.File)
		logger.Debugf("other flags have been ignored.")
	} else {
		logger.Debugf("config file is not used.")
		logger.Debugf("flag: %s used.", os.Args)
	}
	logger.Debugf("%+v", consts.AppInfo)
	logger.Debugf("%+v", inst.Config)

	if !utils.IsFFmpegExist(ctx) {
		logger.Fatalln("FFmpeg binary not found, Please Check.")
	}

	events.NewDispatcher(ctx)

	inst.Lives = make(map[live.ID]live.Live)
	for index, _ := range inst.Config.LiveRooms {
		room := &inst.Config.LiveRooms[index]
		u, err := url.Parse(room.Url)
		if err != nil {
			logger.WithField("url", room).Error(err)
			continue
		}
		opts := make([]live.Option, 0)
		if v, ok := inst.Config.Cookies[u.Host]; ok {
			opts = append(opts, live.WithKVStringCookies(u, v))
		}
		opts = append(opts, live.WithQuality(room.Quality))
		opts = append(opts, live.WithAudioOnly(room.AudioOnly))

		l, err := live.New(u, inst.Cache, opts...)
		if err != nil {
			logger.WithField("url", room).Error(err.Error())
			continue
		}
		if _, ok := inst.Lives[l.GetLiveId()]; ok {
			logger.Errorf("%s is exist!", room)
			continue
		}
		inst.Lives[l.GetLiveId()] = l
		room.LiveId = l.GetLiveId()
	}

	if inst.Config.RPC.Enable {
		if err := servers.NewServer(ctx).Start(ctx); err != nil {
			logger.WithError(err).Fatalf("failed to init server")
		}
	}
	lm := listeners.NewManager(ctx)
	rm := recorders.NewManager(ctx)
	if err := lm.Start(ctx); err != nil {
		logger.Fatalf("failed to init listener manager, error: %s", err)
	}
	if err := rm.Start(ctx); err != nil {
		logger.Fatalf("failed to init recorder manager, error: %s", err)
	}

	if err = metrics.NewCollector(ctx).Start(ctx); err != nil {
		logger.Fatalf("failed to init metrics collector, error: %s", err)
	}

	for _, _live := range inst.Lives {
		room, err := inst.Config.GetLiveRoomByUrl(_live.GetRawUrl())
		if err != nil {
			logger.WithFields(map[string]interface{}{"room": _live.GetRawUrl()}).Error(err)
			panic(err)
		}
		if room.IsListening {
			if err := lm.AddListener(ctx, _live); err != nil {
				logger.WithFields(map[string]interface{}{"url": _live.GetRawUrl()}).Error(err)
			}
		}
		time.Sleep(time.Second * 5)
	}

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
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
