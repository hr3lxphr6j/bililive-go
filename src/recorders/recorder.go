package recorders

import (
	"context"
	"fmt"
	"github.com/hr3lxphr6j/bililive-go/src/api"
	"github.com/hr3lxphr6j/bililive-go/src/instance"
	"github.com/hr3lxphr6j/bililive-go/src/lib/events"
	"os/exec"
	"path/filepath"
	"time"
)

type Recorder struct {
	Info       *api.Info
	Live       api.Live
	OutPutFile string
	StartTime  time.Time

	cmd  *exec.Cmd
	ed   events.IEventDispatcher
	stop chan struct{}
}

func NewRecorder(ctx context.Context, info *api.Info) (*Recorder, error) {
	inst := instance.GetInstance(ctx)
	t := time.Now()
	return &Recorder{
		Info: info,
		Live: info.Live,
		OutPutFile: filepath.Join(
			instance.GetInstance(ctx).Config.OutPutPath,
			fmt.Sprintf("[%02d-%02d-%02d %02d-%02d-%02d][%s]%s.flv",
				t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), info.HostName, info.RoomName)),
		ed: inst.EventDispatcher.(events.IEventDispatcher),
	}, nil
}

func (r *Recorder) run() {
	for {
		select {
		case <-r.stop:
			return
		default:
			urls, err := r.Live.GetUrls()
			if err != nil {
				time.Sleep(5 * time.Second)
				continue
			}
			r.cmd = exec.Command(
				"ffmpeg",
				"-y", "-re",
				"-i", urls[0].String(),
				"-c", "copy",
				"-bsf:a", "aac_adtstoasc",
				"-f", "flv",
				r.OutPutFile,
			)
			r.cmd.Run()

		}
	}
}

func (r *Recorder) Start() error {
	r.StartTime = time.Now()
	r.stop = make(chan struct{})
	go r.run()
	r.ed.DispatchEvent(events.NewEvent(RecordeStart, r.Info))
	return nil
}

func (r *Recorder) Close() {
	close(r.stop)
	stdIn, err := r.cmd.StdinPipe()
	if err == nil {
		stdIn.Write([]byte("q"))
	}
	r.ed.DispatchEvent(events.NewEvent(RecordeStop, r.Info))
}
