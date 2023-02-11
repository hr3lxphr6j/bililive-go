package flv

import "context"

type (
	SoundFormat   uint8
	SoundRate     uint8
	SoundSize     uint8
	SoundType     uint8
	AACPacketType uint8

	AudioTagHeader struct {
		SoundFormat   SoundFormat
		SoundRate     SoundRate
		SoundSize     SoundSize
		SoundType     SoundType
		AACPacketType AACPacketType
	}
)

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

func (p *Parser) parseAudioTag(ctx context.Context, length, timestamp uint32) (*AudioTagHeader, error) {
	b, err := p.i.ReadByte()
	l := length - 1
	if err != nil {
		return nil, err
	}
	tag := new(AudioTagHeader)

	tag.SoundFormat = SoundFormat(b >> 4 & 15)
	tag.SoundRate = SoundRate(b >> 2 & 3)
	tag.SoundSize = SoundSize(b >> 1 & 1)
	tag.SoundType = SoundType(b & 1)

	if tag.SoundFormat == AAC {
		b, err := p.i.ReadByte()
		l -= 1
		if err != nil {
			return nil, err
		}
		tag.AACPacketType = AACPacketType(b)
	}

	// write tag header && audio tag header & AACPacketType
	if err := p.doWrite(ctx, p.i.AllBytes()); err != nil {
		return nil, err
	}
	p.i.Reset()
	// write body
	if err := p.doCopy(ctx, l); err != nil {
		return nil, err
	}

	return tag, nil
}
