package flv

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	audioTag  uint8 = 8
	videoTag  uint8 = 9
	scriptTag uint8 = 18
)

var (
	flvSign      = []byte{0x46, 0x4c, 0x56, 0x01} // flv version01
	NotFlvStream = errors.New("not flv stream")
	UnknownTag   = errors.New("unknown tag")
	IOError      = errors.New("io error")
)

type Metadata struct {
	HasVideo, HasAudio bool
}

type Parser struct {
	Metadata Metadata

	i              io.Reader
	o              io.Writer
	avcHeaderCount uint8
	tagCount       uint32
	buf            *bytes.Buffer

	buf1, buf2, buf3, buf4, bufTH []byte

	hc     *http.Client
	stopCh chan struct{}
}

func NewParser() *Parser {
	return &Parser{
		Metadata: Metadata{},
		buf:      bytes.NewBuffer(make([]byte, 2<<10)),
		buf1:     make([]byte, 1),
		buf2:     make([]byte, 2),
		buf3:     make([]byte, 3),
		buf4:     make([]byte, 4),
		bufTH:    make([]byte, 15),
		hc: &http.Client{
			Timeout: time.Minute,
		},
		stopCh: make(chan struct{}),
	}
}

func (p *Parser) ParseLiveStream(url *url.URL, file string) error {
	// init input
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return err
	}
	resp, err := p.hc.Do(req)
	if err != nil {
		return err
	}
	p.i = resp.Body
	defer resp.Body.Close()

	// init output
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	p.o = f
	defer f.Close()

	// start parse
	return p.doParse()
}

func (p *Parser) Stop() error {
	close(p.stopCh)
	return nil
}

func (p *Parser) doParse() error {
	// header of flv
	buf9 := make([]byte, 9)
	if n, err := p.i.Read(buf9); err != nil || n != len(buf9) {
		if err == nil {
			err = io.EOF
		}
		return err
	}
	// signature
	if !bytes.Equal(buf9[:4], flvSign) {
		return NotFlvStream
	}
	// flag
	p.Metadata.HasVideo = uint8(buf9[4])&(1<<2) != 0
	p.Metadata.HasAudio = uint8(buf9[4])&1 != 0

	// offset must be 9
	if binary.BigEndian.Uint32(buf9[5:]) != 9 {
		return NotFlvStream
	}

	// write flv header
	if err := p.doWrite(buf9); err != nil {
		return err
	}

	for {
		select {
		case <-p.stopCh:
			return nil
		default:
			if err := p.parseTag(); err != nil {
				return err
			}
		}
	}
}

func (p *Parser) doCopy(n uint32) error {
	if n, err := io.CopyN(p.o, p.i, int64(n)); err != nil || n != int64(n) {
		if err == nil {
			err = io.EOF
		}
		return err
	}
	return nil
}

func (p *Parser) doWrite(b []byte) error {
	if n, err := p.o.Write(b); err != nil || n != len(b) {
		if err == nil {
			err = io.EOF
		}
		return err
	}
	return nil
}
