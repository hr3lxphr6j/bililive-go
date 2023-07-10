// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package req

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/imroc/req/v3/internal/dump"
	"github.com/imroc/req/v3/internal/util"
	"net/textproto"
	"strings"
	"sync"
)

func isASCIILetter(b byte) bool {
	b |= 0x20 // make lower case
	return 'a' <= b && b <= 'z'
}

// A textprotoReader implements convenience methods for reading requests
// or responses from a text protocol network connection.
type textprotoReader struct {
	R        *bufio.Reader
	buf      []byte // a re-usable buffer for readContinuedLineSlice
	readLine func() (line []byte, isPrefix bool, err error)
}

// NewReader returns a new textprotoReader reading from r.
//
// To avoid denial of service attacks, the provided bufio.Reader
// should be reading from an io.LimitReader or similar textprotoReader to bound
// the size of responses.
func newTextprotoReader(r *bufio.Reader, ds dump.Dumpers) *textprotoReader {
	commonHeaderOnce.Do(initCommonHeader)
	t := &textprotoReader{R: r}

	if ds.ShouldDump() {
		t.readLine = func() (line []byte, isPrefix bool, err error) {
			line, err = t.R.ReadSlice('\n')
			if len(line) == 0 {
				if err != nil {
					line = nil
				}
				return
			}
			err = nil
			ds.DumpResponseHeader(line)
			if line[len(line)-1] == '\n' {
				drop := 1
				if len(line) > 1 && line[len(line)-2] == '\r' {
					drop = 2
				}
				line = line[:len(line)-drop]
			}
			return
		}
	} else {
		t.readLine = t.R.ReadLine
	}

	return t
}

// ReadLine reads a single line from r,
// eliding the final \n or \r\n from the returned string.
func (r *textprotoReader) ReadLine() (string, error) {
	line, err := r.readLineSlice()
	return string(line), err
}

func (r *textprotoReader) readLineSlice() ([]byte, error) {
	var line []byte

	for {
		l, more, err := r.readLine()
		if err != nil {
			return nil, err
		}
		// Avoid the copy if the first call produced a full line.
		if line == nil && !more {
			return l, nil
		}
		line = append(line, l...)
		if !more {
			break
		}
	}
	return line, nil
}

// trim returns s with leading and trailing spaces and tabs removed.
// It does not assume Unicode or UTF-8.
func trim(s []byte) []byte {
	i := 0
	for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
		i++
	}
	n := len(s)
	for n > i && (s[n-1] == ' ' || s[n-1] == '\t') {
		n--
	}
	return s[i:n]
}

// readContinuedLineSlice reads continued lines from the reader buffer,
// returning a byte slice with all lines. The validateFirstLine function
// is run on the first read line, and if it returns an error then this
// error is returned from readContinuedLineSlice.
func (r *textprotoReader) readContinuedLineSlice(validateFirstLine func([]byte) error) ([]byte, error) {
	if validateFirstLine == nil {
		return nil, fmt.Errorf("missing validateFirstLine func")
	}

	// Read the first line.
	line, err := r.readLineSlice()
	if err != nil {
		return nil, err
	}
	if len(line) == 0 { // blank line - no continuation
		return line, nil
	}

	if err := validateFirstLine(line); err != nil {
		return nil, err
	}

	// Optimistically assume that we have started to buffer the next line
	// and it starts with an ASCII letter (the next header key), or a blank
	// line, so we can avoid copying that buffered data around in memory
	// and skipping over non-existent whitespace.
	if r.R.Buffered() > 1 {
		peek, _ := r.R.Peek(2)
		if len(peek) > 0 && (isASCIILetter(peek[0]) || peek[0] == '\n') ||
			len(peek) == 2 && peek[0] == '\r' && peek[1] == '\n' {
			return trim(line), nil
		}
	}

	// ReadByte or the next readLineSlice will flush the read buffer;
	// copy the slice into buf.
	r.buf = append(r.buf[:0], trim(line)...)

	// Read continuation lines.
	for r.skipSpace() > 0 {
		line, err := r.readLineSlice()
		if err != nil {
			break
		}
		r.buf = append(r.buf, ' ')
		r.buf = append(r.buf, trim(line)...)
	}
	return r.buf, nil
}

// skipSpace skips R over all spaces and returns the number of bytes skipped.
func (r *textprotoReader) skipSpace() int {
	n := 0
	for {
		c, err := r.R.ReadByte()
		if err != nil {
			// Bufio will keep err until next read.
			break
		}
		if c != ' ' && c != '\t' {
			r.R.UnreadByte()
			break
		}
		n++
	}
	return n
}

// A protocolError describes a protocol violation such
// as an invalid response or a hung-up connection.
type protocolError string

func (p protocolError) Error() string {
	return string(p)
}

var colon = []byte(":")

// ReadMIMEHeader reads a MIME-style header from r.
// The header is a sequence of possibly continued Key: Value lines
// ending in a blank line.
// The returned map m maps CanonicalMIMEHeaderKey(key) to a
// sequence of values in the same order encountered in the input.
//
// For example, consider this input:
//
//	My-Key: Value 1
//	Long-Key: Even
//	       Longer Value
//	My-Key: Value 2
//
// Given that input, ReadMIMEHeader returns the map:
//
//	map[string][]string{
//		"My-Key": {"Value 1", "Value 2"},
//		"Long-Key": {"Even Longer Value"},
//	}
func (r *textprotoReader) ReadMIMEHeader() (textproto.MIMEHeader, error) {
	// Avoid lots of small slice allocations later by allocating one
	// large one ahead of time which we'll cut up into smaller
	// slices. If this isn't big enough later, we allocate small ones.
	var strs []string
	hint := r.upcomingHeaderNewlines()
	if hint > 0 {
		strs = make([]string, hint)
	}

	m := make(textproto.MIMEHeader, hint)

	// The first line cannot start with a leading space.
	if buf, err := r.R.Peek(1); err == nil && (buf[0] == ' ' || buf[0] == '\t') {
		line, err := r.readLineSlice()
		if err != nil {
			return m, err
		}
		return m, protocolError("malformed MIME header initial line: " + string(line))
	}

	for {
		kv, err := r.readContinuedLineSlice(mustHaveFieldNameColon)
		if len(kv) == 0 {
			return m, err
		}

		// Key ends at first colon.
		k, v, ok := util.CutBytes(kv, colon)
		if !ok {
			return m, protocolError("malformed MIME header line: " + string(kv))
		}
		key := canonicalMIMEHeaderKey(k)

		// As per RFC 7230 field-name is a token, tokens consist of one or more chars.
		// We could return a protocolError here, but better to be liberal in what we
		// accept, so if we get an empty key, skip it.
		if key == "" {
			continue
		}

		// Skip initial spaces in value.
		value := strings.TrimLeft(string(v), " \t")

		vv := m[key]
		if vv == nil && len(strs) > 0 {
			// More than likely this will be a single-element key.
			// Most headers aren't multi-valued.
			// Set the capacity on strs[0] to 1, so any future append
			// won't extend the slice into the other strings.
			vv, strs = strs[:1:1], strs[1:]
			vv[0] = value
			m[key] = vv
		} else {
			m[key] = append(vv, value)
		}

		if err != nil {
			return m, err
		}
	}
}

// mustHaveFieldNameColon ensures that, per RFC 7230, the
// field-name is on a single line, so the first line must
// contain a colon.
func mustHaveFieldNameColon(line []byte) error {
	if bytes.IndexByte(line, ':') < 0 {
		return protocolError(fmt.Sprintf("malformed MIME header: missing colon: %q", line))
	}
	return nil
}

var nl = []byte("\n")

// upcomingHeaderNewlines returns an approximation of the number of newlines
// that will be in this header. If it gets confused, it returns 0.
func (r *textprotoReader) upcomingHeaderNewlines() (n int) {
	// Try to determine the 'hint' size.
	r.R.Peek(1) // force a buffer load if empty
	s := r.R.Buffered()
	if s == 0 {
		return
	}
	peek, _ := r.R.Peek(s)
	return bytes.Count(peek, nl)
}

const toLower = 'a' - 'A'

// validHeaderFieldByte reports whether b is a valid byte in a header
// field name. RFC 7230 says:
//
//	header-field   = field-name ":" OWS field-value OWS
//	field-name     = token
//	tchar = "!" / "#" / "$" / "%" / "&" / "'" / "*" / "+" / "-" / "." /
//	        "^" / "_" / "`" / "|" / "~" / DIGIT / ALPHA
//	token = 1*tchar
func validHeaderFieldByte(b byte) bool {
	return int(b) < len(isTokenTable) && isTokenTable[b]
}

// canonicalMIMEHeaderKey is like CanonicalMIMEHeaderKey but is
// allowed to mutate the provided byte slice before returning the
// string.
//
// For invalid inputs (if a contains spaces or non-token bytes), a
// is unchanged and a string copy is returned.
func canonicalMIMEHeaderKey(a []byte) string {
	// See if a looks like a header key. If not, return it unchanged.
	for _, c := range a {
		if validHeaderFieldByte(c) {
			continue
		}
		// Don't canonicalize.
		return string(a)
	}

	upper := true
	for i, c := range a {
		// Canonicalize: first letter upper case
		// and upper case after each dash.
		// (Host, User-Agent, If-Modified-Since).
		// MIME headers are ASCII only, so no Unicode issues.
		if upper && 'a' <= c && c <= 'z' {
			c -= toLower
		} else if !upper && 'A' <= c && c <= 'Z' {
			c += toLower
		}
		a[i] = c
		upper = c == '-' // for next time
	}
	// The compiler recognizes m[string(byteSlice)] as a special
	// case, so a copy of a's bytes into a new string does not
	// happen in this map lookup:
	if v := commonHeader[string(a)]; v != "" {
		return v
	}
	return string(a)
}

// commonHeader interns common header strings.
var commonHeader map[string]string

var commonHeaderOnce sync.Once

func initCommonHeader() {
	commonHeader = make(map[string]string)
	for _, v := range []string{
		"Accept",
		"Accept-Charset",
		"Accept-Encoding",
		"Accept-Language",
		"Accept-Ranges",
		"Cache-Control",
		"Cc",
		"Connection",
		"Content-Id",
		"Content-Language",
		"Content-Length",
		"Content-Transfer-Encoding",
		"Content-Type",
		"Cookie",
		"Date",
		"Dkim-Signature",
		"Etag",
		"Expires",
		"From",
		"Host",
		"If-Modified-Since",
		"If-None-Match",
		"In-Reply-To",
		"Last-Modified",
		"Location",
		"Message-Id",
		"Mime-Version",
		"Pragma",
		"Received",
		"Return-Path",
		"Server",
		"Set-Cookie",
		"Subject",
		"To",
		"User-Agent",
		"Via",
		"X-Forwarded-For",
		"X-Imforwards",
		"X-Powered-By",
	} {
		commonHeader[v] = v
	}
}

// isTokenTable is a copy of net/http/lex.go's isTokenTable.
// See https://httpwg.github.io/specs/rfc7230.html#rule.token.separators
var isTokenTable = [127]bool{
	'!':  true,
	'#':  true,
	'$':  true,
	'%':  true,
	'&':  true,
	'\'': true,
	'*':  true,
	'+':  true,
	'-':  true,
	'.':  true,
	'0':  true,
	'1':  true,
	'2':  true,
	'3':  true,
	'4':  true,
	'5':  true,
	'6':  true,
	'7':  true,
	'8':  true,
	'9':  true,
	'A':  true,
	'B':  true,
	'C':  true,
	'D':  true,
	'E':  true,
	'F':  true,
	'G':  true,
	'H':  true,
	'I':  true,
	'J':  true,
	'K':  true,
	'L':  true,
	'M':  true,
	'N':  true,
	'O':  true,
	'P':  true,
	'Q':  true,
	'R':  true,
	'S':  true,
	'T':  true,
	'U':  true,
	'W':  true,
	'V':  true,
	'X':  true,
	'Y':  true,
	'Z':  true,
	'^':  true,
	'_':  true,
	'`':  true,
	'a':  true,
	'b':  true,
	'c':  true,
	'd':  true,
	'e':  true,
	'f':  true,
	'g':  true,
	'h':  true,
	'i':  true,
	'j':  true,
	'k':  true,
	'l':  true,
	'm':  true,
	'n':  true,
	'o':  true,
	'p':  true,
	'q':  true,
	'r':  true,
	's':  true,
	't':  true,
	'u':  true,
	'v':  true,
	'w':  true,
	'x':  true,
	'y':  true,
	'z':  true,
	'|':  true,
	'~':  true,
}
