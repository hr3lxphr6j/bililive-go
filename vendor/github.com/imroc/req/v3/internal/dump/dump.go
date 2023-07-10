package dump

import (
	"context"
	"io"
	"net/http"
)

// Options controls the dump behavior.
type Options interface {
	Output() io.Writer
	RequestHeaderOutput() io.Writer
	RequestBodyOutput() io.Writer
	ResponseHeaderOutput() io.Writer
	ResponseBodyOutput() io.Writer
	RequestHeader() bool
	RequestBody() bool
	ResponseHeader() bool
	ResponseBody() bool
	Async() bool
	Clone() Options
}

func (d *Dumper) WrapResponseBodyReadCloser(rc io.ReadCloser) io.ReadCloser {
	return &dumpReponseBodyReadCloser{rc, d}
}

type dumpReponseBodyReadCloser struct {
	io.ReadCloser
	dump *Dumper
}

func (r *dumpReponseBodyReadCloser) Read(p []byte) (n int, err error) {
	n, err = r.ReadCloser.Read(p)
	r.dump.DumpResponseBody(p[:n])
	if err == io.EOF {
		r.dump.DumpDefault([]byte("\r\n"))
	}
	return
}

func (d *Dumper) WrapRequestBodyWriteCloser(rc io.WriteCloser) io.WriteCloser {
	return &dumpRequestBodyWriteCloser{rc, d}
}

type dumpRequestBodyWriteCloser struct {
	io.WriteCloser
	dump *Dumper
}

func (w *dumpRequestBodyWriteCloser) Write(p []byte) (n int, err error) {
	n, err = w.WriteCloser.Write(p)
	w.dump.DumpRequestBody(p[:n])
	return
}

type dumpRequestHeaderWriter struct {
	w    io.Writer
	dump *Dumper
}

func (w *dumpRequestHeaderWriter) Write(p []byte) (n int, err error) {
	n, err = w.w.Write(p)
	w.dump.DumpRequestHeader(p[:n])
	return
}

func (d *Dumper) WrapRequestHeaderWriter(w io.Writer) io.Writer {
	return &dumpRequestHeaderWriter{
		w:    w,
		dump: d,
	}
}

type dumpRequestBodyWriter struct {
	w    io.Writer
	dump *Dumper
}

func (w *dumpRequestBodyWriter) Write(p []byte) (n int, err error) {
	n, err = w.w.Write(p)
	w.dump.DumpRequestBody(p[:n])
	return
}

func (d *Dumper) WrapRequestBodyWriter(w io.Writer) io.Writer {
	return &dumpRequestBodyWriter{
		w:    w,
		dump: d,
	}
}

// GetResponseHeaderDumpers return Dumpers which need dump response header.
func GetResponseHeaderDumpers(ctx context.Context, dump *Dumper) Dumpers {
	dumpers := GetDumpers(ctx, dump)
	var ds []*Dumper
	for _, d := range dumpers {
		if d.ResponseHeader() {
			ds = append(ds, d)
		}
	}
	return Dumpers(ds)
}

// Dumpers is an array of Dumpper
type Dumpers []*Dumper

// ShouldDump is true if Dumper is not empty.
func (ds Dumpers) ShouldDump() bool {
	return len(ds) > 0
}

func (ds Dumpers) DumpResponseHeader(p []byte) {
	for _, d := range ds {
		d.DumpResponseHeader(p)
	}
}

// Dumper is the dump tool.
type Dumper struct {
	Options
	ch chan *dumpTask
}

type dumpTask struct {
	Data   []byte
	Output io.Writer
}

// NewDumper create a new Dumper.
func NewDumper(opt Options) *Dumper {
	d := &Dumper{
		Options: opt,
		ch:      make(chan *dumpTask, 20),
	}
	return d
}

func (d *Dumper) SetOptions(opt Options) {
	d.Options = opt
	return
}

func (d *Dumper) Clone() *Dumper {
	if d == nil {
		return nil
	}
	return &Dumper{
		Options: d.Options.Clone(),
		ch:      make(chan *dumpTask, 20),
	}
}

func (d *Dumper) DumpTo(p []byte, output io.Writer) {
	if len(p) == 0 || output == nil {
		return
	}
	if d.Async() {
		b := make([]byte, len(p))
		copy(b, p)
		d.ch <- &dumpTask{Data: b, Output: output}
		return
	}
	output.Write(p)
}

func (d *Dumper) DumpDefault(p []byte) {
	d.DumpTo(p, d.Output())
}

func (d *Dumper) DumpRequestHeader(p []byte) {
	d.DumpTo(p, d.RequestHeaderOutput())
}

func (d *Dumper) DumpRequestBody(p []byte) {
	d.DumpTo(p, d.RequestBodyOutput())
}

func (d *Dumper) DumpResponseHeader(p []byte) {
	d.DumpTo(p, d.ResponseHeaderOutput())
}

func (d *Dumper) DumpResponseBody(p []byte) {
	d.DumpTo(p, d.ResponseBodyOutput())
}

func (d *Dumper) Stop() {
	d.ch <- nil
}

func (d *Dumper) Start() {
	for t := range d.ch {
		if t == nil {
			return
		}
		t.Output.Write(t.Data)
	}
}

type dumperKeyType int

const DumperKey dumperKeyType = iota

func GetDumpers(ctx context.Context, dump *Dumper) []*Dumper {
	dumps := []*Dumper{}
	if dump != nil {
		dumps = append(dumps, dump)
	}
	if ctx == nil {
		return dumps
	}
	if d, ok := ctx.Value(DumperKey).(*Dumper); ok {
		dumps = append(dumps, d)
	}
	return dumps
}

func WrapResponseBodyIfNeeded(res *http.Response, req *http.Request, dump *Dumper) {
	dumps := GetDumpers(req.Context(), dump)
	for _, d := range dumps {
		if d.ResponseBody() {
			res.Body = d.WrapResponseBodyReadCloser(res.Body)
		}
	}
}
