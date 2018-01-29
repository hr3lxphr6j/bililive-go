package recorders

import (
	"net/url"
	"os/exec"
	"time"
)

type Recorder struct {
	LiveUrl    *url.URL
	OutPutFile string
	StartTime  time.Time
	cmd        *exec.Cmd
}

func (r *Recorder) Start() error {
	cmd := exec.Command(
		"ffmpeg",
		"-y", "-re",
		"-i", r.LiveUrl.String(),
		"-c", "copy",
		"-bsf:a", "aac_adtstoasc",
		r.OutPutFile,
	)
	r.StartTime = time.Now()
	return cmd.Start()
}

func (r *Recorder) Close() {
	stdIn, err := r.cmd.StdinPipe()
	if err != nil {
		return
	}
	stdIn.Write([]byte("q"))
}
