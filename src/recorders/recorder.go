package recorders

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
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

	cmd       *exec.Cmd
	cmdStdIn  io.WriteCloser
	cmdStdErr io.ReadCloser
	ed        events.IEventDispatcher
	logger    *interfaces.Logger
	stop      chan struct{}
}

func NewRecorder(ctx context.Context, live api.Live) (*Recorder, error) {
	inst := instance.GetInstance(ctx)
	return &Recorder{
		Live:       live,
		OutPutPath: instance.GetInstance(ctx).Config.OutPutPath,
		ed:         inst.EventDispatcher.(events.IEventDispatcher),
		logger:     inst.Logger,
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
			outputPath := filepath.Join(r.OutPutPath, utils.ReplaceIllegalChar(r.Live.GetPlatformCNName()), utils.ReplaceIllegalChar(r.Live.GetCachedInfo().HostName))
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
			r.cmdStdErr, _ = r.cmd.StderrPipe()
			r.cmd.Start()
			r.logger.WithFields(r.Live.GetInfoMap()).WithField("stream_url", urls[0].String()).Debug("ffmpeg start")
			if b, err := ioutil.ReadAll(r.cmdStdErr); err == nil {
				r.logger.WithFields(r.Live.GetInfoMap()).WithField("std_err", string(b)).Debug("ffmpeg log info")
			}
			r.cmd.Wait()
			r.logger.WithFields(r.Live.GetInfoMap()).WithField("stream_url", urls[0].String()).Debug("ffmpeg stop")

		}
	}
}

func (r *Recorder) Start() error {
	r.stop = make(chan struct{})
	go r.run()
	r.logger.WithFields(r.Live.GetInfoMap()).Info("Recorde Start")
	r.ed.DispatchEvent(events.NewEvent(RecorderStart, r.Live))
	return nil
}

func (r *Recorder) Close() {
	close(r.stop)
	r.cmdStdIn.Write([]byte("q"))
	r.logger.WithFields(r.Live.GetInfoMap()).Info("Recorde End")
	r.ed.DispatchEvent(events.NewEvent(RecorderStop, r.Live))
}
