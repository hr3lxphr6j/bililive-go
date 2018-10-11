package flv

import (
	"bytes"
	"encoding/binary"
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
}

func NewParser(i io.Reader, o io.Writer) (*Parser, error) {
	if i == nil || o == nil {
		return nil, IOError
	}
	return &Parser{
		Metadata: Metadata{},
		i:        i,
		o:        o,
		buf:      bytes.NewBuffer(make([]byte, 2<<10)),
		buf1:     make([]byte, 1),
		buf2:     make([]byte, 2),
		buf3:     make([]byte, 3),
		buf4:     make([]byte, 4),
		bufTH:    make([]byte, 15),
	}, nil
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
		if err := p.parseTag(); err != nil {
			return err
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
