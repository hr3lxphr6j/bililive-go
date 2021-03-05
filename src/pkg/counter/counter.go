package counter

import (
	"io"
)

type Counter interface {
	Count() uint
}

type CountReader interface {
	Counter
	io.Reader
}

type CountWriter interface {
	Counter
	io.Writer
}

type countReader struct {
	r     io.Reader
	total uint
}

func NewCountReader(r io.Reader) CountReader {
	return &countReader{r: r}
}

func (r *countReader) Count() uint {
	return r.total
}

func (r *countReader) Read(p []byte) (int, error) {
	n, err := r.r.Read(p)
	r.total += uint(n)
	return n, err
}

type countWriter struct {
	w     io.Writer
	total uint
}

func NewCountWriter(w io.Writer) CountWriter {
	return &countWriter{w: w}
}

func (w *countWriter) Count() uint {
	return w.total
}

func (w *countWriter) Write(p []byte) (int, error) {
	n, err := w.w.Write(p)
	w.total += uint(n)
	return n, err
}
