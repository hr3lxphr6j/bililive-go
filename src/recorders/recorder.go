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

	"github.com/hr3lxphr6j/bililive-go/src/api"
	"github.com/hr3lxphr6j/bililive-go/src/configs"
	"github.com/hr3lxphr6j/bililive-go/src/instance"
	"github.com/hr3lxphr6j/bililive-go/src/interfaces"
	"github.com/hr3lxphr6j/bililive-go/src/lib/events"
	"github.com/hr3lxphr6j/bililive-go/src/lib/parser"
	"github.com/hr3lxphr6j/bililive-go/src/lib/parser/ffmpeg"
	"github.com/hr3lxphr6j/bililive-go/src/lib/parser/native/flv"
	"github.com/hr3lxphr6j/bililive-go/src/lib/utils"
)

const (
	begin uint32 = iota
	pending
	running
	stopped
)

type Recorder struct {
	Live       api.Live
	OutPutPath string

	config *configs.Config
	ed     events.IEventDispatcher
	logger *interfaces.Logger

	parser     parser.Parser
	parserLock *sync.RWMutex

	stop  chan struct{}
	state uint32
}

func NewRecorder(ctx context.Context, live api.Live) (*Recorder, error) {
	inst := instance.GetInstance(ctx)
	return &Recorder{
		Live:       live,
		OutPutPath: instance.GetInstance(ctx).Config.OutPutPath,
		config:     inst.Config,
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
			if err != nil {
				time.Sleep(5 * time.Second)
				continue
			}
			var (
				platformName = utils.ReplaceIllegalChar(r.Live.GetPlatformCNName())
				hostName     = utils.ReplaceIllegalChar(r.Live.GetCachedInfo().HostName)
				roomName     = utils.ReplaceIllegalChar(r.Live.GetCachedInfo().RoomName)
				fileName     = fmt.Sprintf("[%s][%s][%s].flv", time.Now().Format("2006-01-02 15-04-05"), hostName, roomName)
				outputPath   = filepath.Join(r.OutPutPath, platformName, hostName)
				file         = filepath.Join(outputPath, fileName)
				url          = urls[0]
			)
			os.MkdirAll(outputPath, os.ModePerm)
			if strings.Contains(url.Path, ".flv") && r.config.Feature.UseNativeFlvParser {
				r.setAndCloseParser(flv.NewParser())
			} else {
				r.setAndCloseParser(ffmpeg.New())
			}
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
	r.logger.WithFields(r.Live.GetInfoMap()).Info("Recorde Start")
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
	r.logger.WithFields(r.Live.GetInfoMap()).Info("Recorde End")
	r.ed.DispatchEvent(events.NewEvent(RecorderStop, r.Live))
}
