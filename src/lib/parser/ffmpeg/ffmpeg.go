package ffmpeg

import (
	"io"
	"net/url"
	"os/exec"
)

type Parser struct {
	cmd      *exec.Cmd
	cmdStdIn io.WriteCloser
}

func New() *Parser {
	return new(Parser)
}

func (p *Parser) ParseLiveStream(url *url.URL, file string) error {
	p.cmd = exec.Command(
		"ffmpeg",
		"-loglevel", "warning",
		"-y", "-re",
		"-timeout", "60000000",
		"-i", url.String(),
		"-c", "copy",
		"-bsf:a", "aac_adtstoasc",
		"-f", "flv",
		file,
	)
	p.cmdStdIn, _ = p.cmd.StdinPipe()
	p.cmd.Start()
	return p.cmd.Wait()
}

func (p *Parser) Stop() error {
	_, err := p.cmdStdIn.Write([]byte("q"))
	return err
}
