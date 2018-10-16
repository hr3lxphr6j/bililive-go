package flv

import "io"

type SoundFormat uint8
type SoundRate uint8
type SoundSize uint8
type SoundType uint8
type AACPacketType uint8
type AudioTagHeader struct {
	SoundFormat   SoundFormat
	SoundRate     SoundRate
	SoundSize     SoundSize
	SoundType     SoundType
	AACPacketType AACPacketType
}

const (
	// SoundFormat
	LPCM_PE  SoundFormat = 0 // Linear PCM, platform endian
	ADPCM    SoundFormat = 1
	MP3      SoundFormat = 2
	LPCM_LE  SoundFormat = 3 // Linear PCM, little endian
	AAC      SoundFormat = 10
	Speex    SoundFormat = 11
	MP3_8kHz SoundFormat = 14 // MP3 8 kHz

	// SoundRate
	Rate5kHz  SoundRate = 0 // 5.5kHz
	Rate11kHz SoundRate = 1 // 11 kHz
	Rate22kHz SoundRate = 2 // 22 kHz
	Rate44kHz SoundRate = 3 // 44 kHz

	// SoundSize
	Sample8  uint8 = 0 // 8-bit samples
	Sample16 uint8 = 1 // 16-bit samples

	// SoundType
	Mono   SoundType = 0 // Mono sound
	Stereo SoundType = 1 // Stereo sound

	// AACPacketType
	AACSeqHeader AACPacketType = 0
	AACRaw       AACPacketType = 1
)

func (p *Parser) parseAudioTag(length, timestamp uint32) (*AudioTagHeader, error) {
	if n, err := p.i.Read(p.buf1); err != nil || n != len(p.buf1) {
		if err == nil {
			err = io.EOF
		}
		return nil, err
	}
	p.buf.Write(p.buf1)
	head := uint8(p.buf1[0])
	tag := new(AudioTagHeader)

	tag.SoundFormat = SoundFormat(head >> 4 & 15)
	tag.SoundRate = SoundRate(head >> 2 & 3)
	tag.SoundSize = SoundSize(head >> 1 & 1)
	tag.SoundType = SoundType(head & 1)

	l := length - 1
	if tag.SoundFormat == AAC {
		l -= 1
		if n, err := p.i.Read(p.buf1); err != nil || n != len(p.buf1) {
			if err == nil {
				err = io.EOF
			}
			return nil, err
		}
		p.buf.Write(p.buf1)
		tag.AACPacketType = AACPacketType(p.buf1[0])
	}

	// write tag header
	if err := p.doWrite(p.bufTH); err != nil {
		return nil, err
	}
	// write audio tag header & AACPacketType
	if err := p.doWrite(p.buf.Bytes()); err != nil {
		return nil, err
	}
	p.buf.Reset()
	// write body
	if err := p.doCopy(l); err != nil {
		return nil, err
	}

	return tag, nil
}
