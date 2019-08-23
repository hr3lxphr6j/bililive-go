package recorders

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bluele/gcache"

	"github.com/hr3lxphr6j/bililive-go/src/configs"
	"github.com/hr3lxphr6j/bililive-go/src/instance"
	"github.com/hr3lxphr6j/bililive-go/src/interfaces"
	"github.com/hr3lxphr6j/bililive-go/src/lib/events"
	"github.com/hr3lxphr6j/bililive-go/src/lib/parser"
	"github.com/hr3lxphr6j/bililive-go/src/lib/parser/ffmpeg"
	"github.com/hr3lxphr6j/bililive-go/src/lib/parser/native/flv"
	"github.com/hr3lxphr6j/bililive-go/src/lib/utils"
	"github.com/hr3lxphr6j/bililive-go/src/live"
)

const (
	begin uint32 = iota
	pending
	running
	stopped
)

type Recorder struct {
	Live       live.Live
	OutPutPath string

	config *configs.Config
	ed     events.IEventDispatcher
	logger *interfaces.Logger
	cache  gcache.Cache

	parser     parser.Parser
	parserLock *sync.RWMutex

	stop  chan struct{}
	state uint32
}

func NewRecorder(ctx context.Context, live live.Live) (*Recorder, error) {
	inst := instance.GetInstance(ctx)
	return &Recorder{
		Live:       live,
		OutPutPath: instance.GetInstance(ctx).Config.OutPutPath,
		config:     inst.Config,
		cache:      inst.Cache,
		ed:         inst.EventDispatcher.(events.IEventDispatcher),
		logger:     inst.Logger,
		state:      begin,
		stop:       make(chan struct{}),
		parserLock: new(sync.RWMutex),
	}, nil
}

func (r *Recorder) run() {
	for {
		select {
		case <-r.stop:
			return
		default:
			urls, err := r.Live.GetStreamUrls()
			if err != nil || len(urls) == 0 {
				time.Sleep(5 * time.Second)
				continue
			}

			obj, _ := r.cache.Get(r.Live)
			info := obj.(*live.Info)
			var (
				platformName = utils.ReplaceIllegalChar(r.Live.GetPlatformCNName())
				hostName     = utils.ReplaceIllegalChar(info.HostName)
				roomName     = utils.ReplaceIllegalChar(info.RoomName)
				fileName     = fmt.Sprintf("[%s][%s][%s].flv", time.Now().Format("2006-01-02 15-04-05"), hostName, roomName)
				outputPath   = filepath.Join(r.OutPutPath, platformName, hostName)
				file         = filepath.Join(outputPath, fileName)
				url          = urls[0]
			)
			os.MkdirAll(outputPath, os.ModePerm)
			parserName := ffmpeg.Name
			if strings.Contains(url.Path, ".flv") && r.config.Feature.UseNativeFlvParser {
				parserName = flv.Name
			}
			p, err := parser.New(parserName)
			if err != nil {
				continue
			}
			r.setAndCloseParser(p)
			r.logger.Debugln(r.parser.ParseLiveStream(url, file))
			if stat, err := os.Stat(file); err == nil && stat.Size() == 0 {
				os.Remove(file)
			}
		}
	}
}

func (r *Recorder) getParser() parser.Parser {
	r.parserLock.RLock()
	defer r.parserLock.RUnlock()
	return r.parser
}

func (r *Recorder) setAndCloseParser(p parser.Parser) {
	r.parserLock.Lock()
	defer r.parserLock.Unlock()
	if r.parser != nil {
		r.parser.Stop()
	}
	r.parser = p
}

func (r *Recorder) Start() error {
	if !atomic.CompareAndSwapUint32(&r.state, begin, pending) {
		return fmt.Errorf("recorder in error state")
	}
	go r.run()
	r.logger.WithFields(r.getFields()).Info("Record Start")
	r.ed.DispatchEvent(events.NewEvent(RecorderStart, r.Live))
	atomic.CompareAndSwapUint32(&r.state, pending, running)
	return nil
}

func (r *Recorder) Close() {
	if !atomic.CompareAndSwapUint32(&r.state, running, stopped) {
		return
	}
	close(r.stop)
	if p := r.getParser(); p != nil {
		p.Stop()
	}
	r.logger.WithFields(r.getFields()).Info("Record End")
	r.ed.DispatchEvent(events.NewEvent(RecorderStop, r.Live))
}

func (r *Recorder) getFields() map[string]interface{} {
	obj, err := r.cache.Get(r.Live)
	info := obj.(*live.Info)
	if err != nil {
		return nil
	}
	return map[string]interface{}{
		"host": info.HostName,
		"room": info.RoomName,
	}
}
