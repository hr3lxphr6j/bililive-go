package flv

import (
	"encoding/json"
	"fmt"
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
		fmt.Printf("[%04d]: Type: Audio, Length: %d, TimeStamp: %d\n", p.tagCount, length, timeStamp)
		if header, err := p.parseAudioTag(length, timeStamp); err == nil {
			printTagInfo(header)
		}
	case videoTag:
		fmt.Printf("[%04d]: Type: Video, Length: %d, TimeStamp: %d\n", p.tagCount, length, timeStamp)
		if header, err := p.parseVideoTag(length, timeStamp); err == nil {
			printTagInfo(header)
		}
	case scriptTag:
		fmt.Printf("[%04d]: Type: Script, Length: %d, TimeStamp: %d\n", p.tagCount, length, timeStamp)
		p.parseScriptTag(length)
	default:
		return UnknownTag
	}

	return nil
}

// for test
func printTagInfo(data interface{}) {
	b, _ := json.Marshal(data)
	fmt.Printf("\t%s\n\n", string(b))
}
