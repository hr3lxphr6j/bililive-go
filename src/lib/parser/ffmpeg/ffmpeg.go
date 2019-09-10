package ffmpeg

import (
	"io"
	"net/url"
	"os/exec"
	"sync"

	"github.com/hr3lxphr6j/bililive-go/src/lib/parser"
	"github.com/hr3lxphr6j/bililive-go/src/live"
)

const (
	Name = "ffmpeg"

	userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Safari/537.36"
)

func init() {
	parser.Register(Name, new(builder))
}

type builder struct{}

func (b *builder) Build() (parser.Parser, error) {
	return &Parser{
		closeOnce: new(sync.Once),
	}, nil
}

type Parser struct {
	cmd       *exec.Cmd
	cmdStdIn  io.WriteCloser
	closeOnce *sync.Once
}

func (p *Parser) ParseLiveStream(url *url.URL, live live.Live, file string) error {
	p.cmd = exec.Command(
		"ffmpeg",
		"-loglevel", "warning",
		"-y", "-re",
		"-user_agent", userAgent,
		"-referer", live.GetRawUrl(),
		"-timeout", "60000000",
		"-i", url.String(),
		"-c", "copy",
		"-bsf:a", "aac_adtstoasc",
		"-f", "flv",
		file,
	)
	stdIn, err := p.cmd.StdinPipe()
	if err != nil {
		return err
	}
	p.cmdStdIn = stdIn
	if err := p.cmd.Start(); err != nil {
		p.cmd.Process.Kill()
		return err
	}
	return p.cmd.Wait()
}

func (p *Parser) Stop() error {
	p.closeOnce.Do(func() {
		if p.cmd.ProcessState == nil {
			p.cmdStdIn.Write([]byte("q"))
		}
	})
	return nil
}
