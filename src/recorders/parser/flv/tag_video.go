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
	head := uint8(p.buf1[0])
	tag := new(VideoTagHeader)
	tag.FrameType = FrameType(head >> 4 & 15)
	tag.CodeID = CodeID(head & 15)

	offset := length - 1
	if tag.CodeID == AVCCode {
		offset -= 1
		if n, err := p.i.Read(p.buf1); err != nil || n != len(p.buf1) {
			if err == nil {
				err = io.EOF
			}
			return nil, err
		}
		tag.AVCPacketType = AVCPacketType(p.buf1[0])
		if tag.AVCPacketType == AVCNALU {
			offset -= 3
			if n, err := p.i.Read(p.buf3); err != nil || n != len(p.buf3) {
				if err == nil {
					err = io.EOF
				}
				return nil, err
			}
			tag.CompositionTime = uint32(p.buf3[0])<<16 | uint32(p.buf3[1])<<8 | uint32(p.buf3[2])
		}
	}

	// body
	buf := make([]byte, offset)
	if n, err := p.i.Read(buf); err != nil || n != len(buf) {
		if err == nil {
			err = io.EOF
		}
		return nil, err
	}

	return tag, nil
}
