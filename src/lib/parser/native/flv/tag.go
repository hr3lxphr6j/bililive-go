package flv

import (
	"io"
)

func (p *Parser) parseTag() error {
	p.tagCount += 1

	if n, err := p.i.Read(p.bufTH); err != nil || n != len(p.bufTH) {
		if err == nil {
			err = io.EOF
		}
		return err
	}

	tagType := uint8(p.bufTH[4])
	length := uint32(p.bufTH[5])<<16 | uint32(p.bufTH[6])<<8 | uint32(p.bufTH[7])
	timeStamp := uint32(p.bufTH[8])<<16 | uint32(p.bufTH[9])<<8 | uint32(p.bufTH[10]) | uint32(p.bufTH[11])<<24

	switch tagType {
	case audioTag:
		if _, err := p.parseAudioTag(length, timeStamp); err != nil {
			return err
		}
	case videoTag:
		if _, err := p.parseVideoTag(length, timeStamp); err != nil {
			return err
		}
	case scriptTag:
		if err := p.parseScriptTag(length); err != nil {
			return err
		}
	default:
		return UnknownTag
	}

	return nil
}
