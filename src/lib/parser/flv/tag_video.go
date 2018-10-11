package flv

import (
	"io"
)

type FrameType uint8
type CodeID uint8
type AVCPacketType uint8
type VideoTagHeader struct {
	FrameType       FrameType
	CodeID          CodeID
	AVCPacketType   AVCPacketType
	CompositionTime uint32
}

const (
	// Frame Type
	KeyFrame             FrameType = 1 // for AVC, a seekable frame
	InterFrame           FrameType = 2 // for AVC, a non-seekable frame
	DisposableInterFrame FrameType = 3 // H.263 only
	GeneratedKeyFrame    FrameType = 4 // reserved for server use only
	VideoInfoFrame       FrameType = 5 // video info/command frame

	// CodeID
	H263Code          CodeID = 2 // Sorenson H.263
	ScreenVideoCode   CodeID = 3 // Screen video
	VP6Code           CodeID = 4 // On2 VP6
	VP6AlphaCode      CodeID = 5 // On2 VP6 with alpha channel
	ScreenVideoV2Code CodeID = 6 // Screen video version 2
	AVCCode           CodeID = 7 // AVC

	// AVCPacketType
	AVCSeqHeader AVCPacketType = 0 // AVC sequence header
	AVCNALU      AVCPacketType = 1 // NALU
	AVCEndSeq    AVCPacketType = 2 // AVC end of sequence (lower level NALU sequence ender is not required or supported)
)

func (p *Parser) parseVideoTag(length, timestamp uint32) (*VideoTagHeader, error) {
	// header
	if n, err := p.i.Read(p.buf1); err != nil || n != len(p.buf1) {
		if err == nil {
			err = io.EOF
		}
		return nil, err
	}
	p.buf.Write(p.buf1)
	head := uint8(p.buf1[0])
	tag := new(VideoTagHeader)
	tag.FrameType = FrameType(head >> 4 & 15)
	tag.CodeID = CodeID(head & 15)

	l := length - 1
	if tag.CodeID == AVCCode {
		// read AVCPacketType
		l -= 1
		if n, err := p.i.Read(p.buf1); err != nil || n != len(p.buf1) {
			if err == nil {
				err = io.EOF
			}
			return nil, err
		}
		p.buf.Write(p.buf1)
		tag.AVCPacketType = AVCPacketType(p.buf1[0])
		switch tag.AVCPacketType {
		case AVCNALU:
			// read CompositionTime
			l -= 3
			if n, err := p.i.Read(p.buf3); err != nil || n != len(p.buf3) {
				if err == nil {
					err = io.EOF
				}
				return nil, err
			}
			p.buf.Write(p.buf3)
			tag.CompositionTime = uint32(p.buf3[0])<<16 | uint32(p.buf3[1])<<8 | uint32(p.buf3[2])
		case AVCSeqHeader:
			p.avcHeaderCount++
			if p.avcHeaderCount > 1 {
				// new sps pps
				return nil, io.EOF
			}
		}
	}

	// write tag header
	if err := p.doWrite(p.bufTH); err != nil {
		return nil, err
	}
	// write video tag header & AVCPacketType & CompositionTime
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
