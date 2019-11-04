package flv

func (p *Parser) parseTag() error {
	p.tagCount += 1

	b, err := p.i.ReadN(15)
	if err != nil {
		return err
	}

	tagType := uint8(b[4])
	length := uint32(b[5])<<16 | uint32(b[6])<<8 | uint32(b[7])
	timeStamp := uint32(b[8])<<16 | uint32(b[9])<<8 | uint32(b[10]) | uint32(b[11])<<24

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
		return p.parseScriptTag(length)
	default:
		return ErrUnknownTag
	}

	return nil
}
