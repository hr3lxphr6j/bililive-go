package recorders

import (
	"context"
	"fmt"
	"github.com/hr3lxphr6j/bililive-go/src/configs"
	"github.com/hr3lxphr6j/bililive-go/src/lib/parser/flv"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hr3lxphr6j/bililive-go/src/api"
	"github.com/hr3lxphr6j/bililive-go/src/instance"
	"github.com/hr3lxphr6j/bililive-go/src/interfaces"
	"github.com/hr3lxphr6j/bililive-go/src/lib/events"
	"github.com/hr3lxphr6j/bililive-go/src/lib/utils"
)

type Recorder struct {
	Live       api.Live
	OutPutPath string

	config               *configs.Config
	ed                   events.IEventDispatcher
	logger               *interfaces.Logger
	startOnce, closeOnce *sync.Once
	stop                 chan struct{}

	cmd       *exec.Cmd
	cmdStdIn  io.WriteCloser
	cmdStderr io.ReadCloser
}

func NewRecorder(ctx context.Context, live api.Live) (*Recorder, error) {
	inst := instance.GetInstance(ctx)
	return &Recorder{
		Live:       live,
		OutPutPath: instance.GetInstance(ctx).Config.OutPutPath,
		config:     inst.Config,
		ed:         inst.EventDispatcher.(events.IEventDispatcher),
		logger:     inst.Logger,
		startOnce:  new(sync.Once),
		closeOnce:  new(sync.Once),
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
			t := time.Now()
			outputPath := filepath.Join(
				r.OutPutPath,
				utils.ReplaceIllegalChar(r.Live.GetPlatformCNName()),
				utils.ReplaceIllegalChar(r.Live.GetCachedInfo().HostName),
			)
			os.MkdirAll(outputPath, os.ModePerm)
			outfile := filepath.Join(
				outputPath,
				fmt.Sprintf(
					"[%02d-%02d-%02d %02d-%02d-%02d][%s][%s].flv",
					t.Year(), t.Month(), t.Day(), t.Hour(),
					t.Minute(), t.Second(),
					utils.ReplaceIllegalChar(r.Live.GetCachedInfo().HostName),
					utils.ReplaceIllegalChar(r.Live.GetCachedInfo().RoomName),
				),
			)
			if strings.Contains(urls[0].Path, ".flv") && r.config.Feature.UseNativeFlvParser {
				r.callNativeFlvParser(urls, outputPath, outfile)
			} else {
				r.callFFmpeg(urls, outputPath, outfile)
			}
		}
	}
}

func (r *Recorder) callNativeFlvParser(urls []*url.URL, outputPath, outfile string) error {
	// init input
	resp, err := http.Get(urls[0].String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// init output
	f, err := os.OpenFile(outfile, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	parser, err := flv.NewParser(resp.Body, f)
	if err != nil {

	}
	return parser.ParseFlv()
}

func (r *Recorder) callFFmpeg(urls []*url.URL, outputPath, outfile string) error {
	r.cmd = exec.Command(
		"ffmpeg",
		"-loglevel", "warning",
		"-y", "-re",
		"-timeout", "60000000",
		"-i", urls[0].String(),
		"-c", "copy",
		"-bsf:a", "aac_adtstoasc",
		"-f", "flv",
		outfile,
	)
	r.cmdStdIn, _ = r.cmd.StdinPipe()
	if r.config.Debug {
		r.cmdStderr, _ = r.cmd.StderrPipe()
		go r.redirectTo(outputPath)
	}
	r.cmd.Start()
	return r.cmd.Wait()
}

func (r *Recorder) redirectTo(outputPath string) {
	f, err := os.OpenFile(path.Join(outputPath, "ffmpeg_debug.log"), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		r.logger.Warn(err)
		return
	}
	r.logger.Debugf("ffmped stderr -> %s", f.Name())
	fmt.Fprintf(f, "============= %s =============\n", time.Now().String())
	buf := make([]byte, 1024)
	io.CopyBuffer(f, r.cmdStderr, buf)
	f.Close()
}

func (r *Recorder) Start() error {
	r.startOnce.Do(func() {
		r.stop = make(chan struct{})
		go r.run()
		r.logger.WithFields(r.Live.GetInfoMap()).Info("Recorde Start")
		r.ed.DispatchEvent(events.NewEvent(RecorderStart, r.Live))
	})
	return nil
}

func (r *Recorder) Close() {
	r.closeOnce.Do(func() {
		close(r.stop)
		r.cmdStdIn.Write([]byte("q"))
		r.logger.WithFields(r.Live.GetInfoMap()).Info("Recorde End")
		r.ed.DispatchEvent(events.NewEvent(RecorderStop, r.Live))
	})
}
