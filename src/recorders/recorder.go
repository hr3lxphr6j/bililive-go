//go:generate mockgen -package recorders -destination mock_test.go github.com/hr3lxphr6j/bililive-go/src/recorders Recorder,Manager
package recorders

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bluele/gcache"
	"github.com/sirupsen/logrus"

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

// for test
var (
	newParser = func(u *url.URL, useNativeFlvParser bool) (parser.Parser, error) {
		parserName := ffmpeg.Name
		if strings.Contains(u.Path, ".flv") && useNativeFlvParser {
			parserName = flv.Name
		}
		return parser.New(parserName)
	}

	mkdir = func(path string) error {
		return os.MkdirAll(path, os.ModePerm)
	}

	removeEmptyFile = func(file string) {
		if stat, err := os.Stat(file); err == nil && stat.Size() == 0 {
			os.Remove(file)
		}
	}
)

type Recorder interface {
	Start() error
	Close()
}

type recorder struct {
	Live       live.Live
	OutPutPath string

	config *configs.Config
	ed     events.Dispatcher
	logger *interfaces.Logger
	cache  gcache.Cache

	parser     parser.Parser
	parserLock *sync.RWMutex

	stop  chan struct{}
	state uint32
}

func NewRecorder(ctx context.Context, live live.Live) (Recorder, error) {
	inst := instance.GetInstance(ctx)
	return &recorder{
		Live:       live,
		OutPutPath: instance.GetInstance(ctx).Config.OutPutPath,
		config:     inst.Config,
		cache:      inst.Cache,
		ed:         inst.EventDispatcher.(events.Dispatcher),
		logger:     inst.Logger,
		state:      begin,
		stop:       make(chan struct{}),
		parserLock: new(sync.RWMutex),
	}, nil
}

func (r *recorder) tryRecode() {
	urls, err := r.Live.GetStreamUrls()
	if err != nil || len(urls) == 0 {
		r.getLogger().WithError(err).Warn("failed to get stream url, will retry after 5s...")
		time.Sleep(5 * time.Second)
		return
	}

	obj, _ := r.cache.Get(r.Live)
	info := obj.(*live.Info)
	var (
		strFilter    = utils.NewStringFilterChain(utils.ReplaceIllegalChar, utils.UnescapeHTMLEntity)
		platformName = strFilter.Do(r.Live.GetPlatformCNName())
		hostName     = strFilter.Do(info.HostName)
		roomName     = strFilter.Do(info.RoomName)
		ts           = time.Now().Format("2006-01-02 15-04-05")
		fileName     = fmt.Sprintf("[%s][%s][%s].flv", ts, hostName, roomName)
		outputPath   = filepath.Join(r.OutPutPath, platformName, hostName)
		file         = filepath.Join(outputPath, fileName)
		url          = urls[0]
	)
	if err := mkdir(outputPath); err != nil {
		r.getLogger().WithError(err).Errorf("failed to create output path[%s]", outputPath)
		return
	}
	p, err := newParser(url, r.config.Feature.UseNativeFlvParser)
	if err != nil {
		r.getLogger().WithError(err).Error("failed to init parse")
		return
	}
	r.setAndCloseParser(p)
	r.getLogger().Debugln(r.parser.ParseLiveStream(url, r.Live, file))
	removeEmptyFile(file)
}

func (r *recorder) run() {
	for {
		select {
		case <-r.stop:
			return
		default:
			r.tryRecode()
		}
	}
}

func (r *recorder) getParser() parser.Parser {
	r.parserLock.RLock()
	defer r.parserLock.RUnlock()
	return r.parser
}

func (r *recorder) setAndCloseParser(p parser.Parser) {
	r.parserLock.Lock()
	defer r.parserLock.Unlock()
	if r.parser != nil {
		r.parser.Stop()
	}
	r.parser = p
}

func (r *recorder) Start() error {
	if !atomic.CompareAndSwapUint32(&r.state, begin, pending) {
		return nil
	}
	go r.run()
	r.getLogger().Info("Record Start")
	r.ed.DispatchEvent(events.NewEvent(RecorderStart, r.Live))
	atomic.CompareAndSwapUint32(&r.state, pending, running)
	return nil
}

func (r *recorder) Close() {
	if !atomic.CompareAndSwapUint32(&r.state, running, stopped) {
		return
	}
	close(r.stop)
	if p := r.getParser(); p != nil {
		p.Stop()
	}
	r.getLogger().Info("Record End")
	r.ed.DispatchEvent(events.NewEvent(RecorderStop, r.Live))
}

func (r *recorder) getLogger() *logrus.Entry {
	return r.logger.WithFields(r.getFields())
}

func (r *recorder) getFields() map[string]interface{} {
	obj, err := r.cache.Get(r.Live)
	if err != nil {
		return nil
	}
	info := obj.(*live.Info)
	return map[string]interface{}{
		"host": info.HostName,
		"room": info.RoomName,
	}
}
