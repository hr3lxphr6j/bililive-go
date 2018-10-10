package flv

import (
	"bytes"
	"errors"
	"io"
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
)

type Metadata struct {
	HasVideo, HasAudio bool
}

type Parser struct {
	Metadata Metadata
	i        io.Reader
	tagCount uint32

	buf1, buf2, buf3, buf4, bufTH []byte
}

func NewParser(i io.Reader) *Parser {
	return &Parser{
		Metadata: Metadata{},
		i:        i,
		tagCount: 0,
		buf1:     make([]byte, 1),
		buf2:     make([]byte, 2),
		buf3:     make([]byte, 3),
		buf4:     make([]byte, 4),
		bufTH:    make([]byte, 15),
	}
}

func (p *Parser) ParseFlv() error {
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
	p.Metadata.HasVideo = uint8(buf9[5])&(1<<2) != 0
	p.Metadata.HasAudio = uint8(buf9[5])&1 != 0

	for {
		if err := p.parseTag(); err != nil {
			return err
		}
	}
}
