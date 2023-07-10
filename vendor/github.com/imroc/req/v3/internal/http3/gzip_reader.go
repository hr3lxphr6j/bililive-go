package http3

// copied from net/transport.go

// GzipReader wraps a response body so it can lazily
// call gzip.NewReader on the first call to Read
import (
	"compress/gzip"
	"io"
)

// call gzip.NewReader on the first call to Read
type GzipReader struct {
	Body io.ReadCloser // underlying Response.Body
	zr   *gzip.Reader  // lazily-initialized gzip reader
	zerr error         // sticky error
}

func newGzipReader(body io.ReadCloser) io.ReadCloser {
	return &GzipReader{Body: body}
}

func (gz *GzipReader) Read(p []byte) (n int, err error) {
	if gz.zerr != nil {
		return 0, gz.zerr
	}
	if gz.zr == nil {
		gz.zr, err = gzip.NewReader(gz.Body)
		if err != nil {
			gz.zerr = err
			return 0, err
		}
	}
	return gz.zr.Read(p)
}

func (gz *GzipReader) Close() error {
	return gz.Body.Close()
}
