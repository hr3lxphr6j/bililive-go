package h264

import (
	"bytes"
	"errors"
	"io"
)

type NaluType uint8

const (
	SLICE    NaluType = 1
	DPA      NaluType = 2
	DPB      NaluType = 3
	DPC      NaluType = 4
	IDR      NaluType = 5
	SEI      NaluType = 6
	SPS      NaluType = 7
	PPS      NaluType = 8
	AUD      NaluType = 9
	EOSEQ    NaluType = 10
	EOSTREAM NaluType = 11
	FILL     NaluType = 12
	UNKNOWN  NaluType = 255
)

var (
	NALUStartCode        = []byte{0x00, 0x00, 0x01}
	NALUStartCode2       = []byte{0x00, 0x00, 0x00, 0x01}
	UnknownNALUStartCode = errors.New("unknown NALU start code")
)

type Parser struct {
	buf                    *bytes.Buffer
	buf1, buf2, buf3, buf4 []byte
}

func NewParser() *Parser {
	return &Parser{
		buf:  bytes.NewBuffer(make([]byte, 2<<10)),
		buf1: make([]byte, 1),
		buf2: make([]byte, 2),
		buf3: make([]byte, 3),
		buf4: make([]byte, 4),
	}
}

func (p *Parser) parseNALUStartCode(i io.Reader) ([]byte, error) {
	if n, err := i.Read(p.buf3); err != nil || n != len(p.buf3) {
		if err == nil {
			err = io.EOF
		}
		return nil, err
	}
	if !bytes.Equal(p.buf3, NALUStartCode) {
		if n, err := i.Read(p.buf1); err != nil || n != len(p.buf1) {
			if err == nil {
				err = io.EOF
			}
			return nil, err
		}
		if !bytes.Equal(append(p.buf3, p.buf1...), NALUStartCode2) {
			return nil, UnknownNALUStartCode
		} else {
			return NALUStartCode2, nil
		}
	} else {
		return NALUStartCode, nil
	}
}

func (p *Parser) ParseAVCSequenceHeader(i io.Reader, length uint32) {

}

func (p *Parser) ParseAnnexBNalu(i io.Reader, length uint32) (NaluType, error) {
	startCode, err := p.parseNALUStartCode(i)
	if err != nil {
		return UNKNOWN, err
	}
	if n, err := i.Read(p.buf1); err != nil || n != len(p.buf1) {
		if err == nil {
			err = io.EOF
		}
		return UNKNOWN, err
	}
	naluType := NaluType(uint8(p.buf1[0]) & 31)
	buf := make([]byte, length-1-uint32(len(startCode)))
	if n, err := i.Read(buf); err != nil || n != len(buf) {
		if err == nil {
			err = io.EOF
		}
		return UNKNOWN, err
	}
	return naluType, nil
}

func (p *Parser) ParseAVCCNalu(i io.Reader, o io.Writer, length uint32) {

}
