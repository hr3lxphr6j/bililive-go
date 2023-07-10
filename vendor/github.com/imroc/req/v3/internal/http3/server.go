package http3

import (
	"github.com/quic-go/quic-go"
)

const (
	nextProtoH3Draft29 = "h3-29"
	nextProtoH3        = "h3"
)

// StreamType is the stream type of a unidirectional stream.
type StreamType uint64

const (
	streamTypeControlStream      = 0
	streamTypePushStream         = 1
	streamTypeQPACKEncoderStream = 2
	streamTypeQPACKDecoderStream = 3
)

func versionToALPN(v quic.VersionNumber) string {
	switch v {
	case Version1, Version2:
		return nextProtoH3
	case VersionDraft29:
		return nextProtoH3Draft29
	}
	return ""
}

type requestError struct {
	err       error
	streamErr errorCode
	connErr   errorCode
}

func newStreamError(code errorCode, err error) requestError {
	return requestError{err: err, streamErr: code}
}

func newConnError(code errorCode, err error) requestError {
	return requestError{err: err, connErr: code}
}
