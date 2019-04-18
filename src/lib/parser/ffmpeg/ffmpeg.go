package ffmpeg

import (
	"net/url"
	"os/exec"
)

const (
	userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Safari/537.36"
)

type Parser struct {
	cmd *exec.Cmd
}

func New() *Parser {
	return new(Parser)
}

func (p *Parser) ParseLiveStream(url *url.URL, file string) error {
	p.cmd = exec.Command(
		"ffmpeg",
		"-loglevel", "warning",
		"-y", "-re",
		"-user_agent", userAgent,
		"-timeout", "60000000",
		"-i", url.String(),
		"-c", "copy",
		"-bsf:a", "aac_adtstoasc",
		"-f", "flv",
		file,
	)
	p.cmd.Start()
	return p.cmd.Wait()
}

func (p *Parser) Stop() error {
	if stdIn, err := p.cmd.StdinPipe(); err != nil {
		return err
	} else {
		stdIn.Write([]byte("q"))
	}
	return nil
}
