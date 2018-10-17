package reader

import (
	"errors"
	"io"
)

const (
	minBufferSize     = 16
	defaultBufferSize = 2048
)

var (
	OutOfBuffer = errors.New("n is bigger than len of buffer")
)

type BufferReader struct {
	io.Reader
	buf  []byte
	l, r int
}

func New(rd io.Reader) *BufferReader {
	return NewWithSize(rd, defaultBufferSize)
}

func NewWithSize(rd io.Reader, size int) *BufferReader {
	if size < minBufferSize {
		size = minBufferSize
	}
	return &BufferReader{
		Reader: rd,
		buf:    make([]byte, size),
	}
}

func (b *BufferReader) ReadN(n int) ([]byte, error) {
	if n > len(b.buf)-b.r {
		return nil, OutOfBuffer
	}
	b.l = b.r
	return b.readN(n, b.l)
}

func (b *BufferReader) readN(n, l int) ([]byte, error) {
	c, err := b.Read(b.buf[l : l+n])
	b.r += c
	if err != nil {
		return nil, err
	}
	if c < n {
		return b.readN(n-c, b.r)
	}
	return b.buf[b.l:b.r], nil
}

func (b *BufferReader) ReadByte() (byte, error) {
	buf, err := b.ReadN(1)
	return buf[0], err
}

func (b *BufferReader) Reset() {
	b.l = 0
	b.r = 0
}

func (b *BufferReader) Cap() int {
	return len(b.buf)
}

func (b *BufferReader) AllBytes() []byte {
	return b.buf[:b.r]
}

func (b *BufferReader) LastBytes() []byte {
	return b.buf[b.l:b.r]
}
