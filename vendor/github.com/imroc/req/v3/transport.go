// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// HTTP client implementation. See RFC 7230 through 7235.
//
// This is the low-level Transport implementation of http.RoundTripper.
// The high-level interface is in client.go.

package req

import (
	"bufio"
	"compress/gzip"
	"container/list"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/imroc/req/v3/internal/altsvcutil"
	"github.com/imroc/req/v3/internal/ascii"
	"github.com/imroc/req/v3/internal/common"
	"github.com/imroc/req/v3/internal/dump"
	"github.com/imroc/req/v3/internal/header"
	"github.com/imroc/req/v3/internal/http2"
	"github.com/imroc/req/v3/internal/http3"
	"github.com/imroc/req/v3/internal/netutil"
	"github.com/imroc/req/v3/internal/socks"
	"github.com/imroc/req/v3/internal/transport"
	"github.com/imroc/req/v3/internal/util"
	"github.com/imroc/req/v3/pkg/altsvc"
	reqtls "github.com/imroc/req/v3/pkg/tls"
	htmlcharset "golang.org/x/net/html/charset"
	"golang.org/x/text/encoding/ianaindex"
	"io"
	"log"
	"mime"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/textproto"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/http/httpguts"
)

// httpVersion represents http version.
type httpVersion string

const (
	// h1 represents "HTTP/1.1"
	h1 httpVersion = "1.1"
	// h2 represents "HTTP/2.0"
	h2 httpVersion = "2"
	// h3 represents "HTTP/3.0"
	h3 httpVersion = "3"
)

// defaultMaxIdleConnsPerHost is the default value of Transport's
// MaxIdleConnsPerHost.
const defaultMaxIdleConnsPerHost = 2

// Transport is an implementation of http.RoundTripper that supports HTTP,
// HTTPS, and HTTP proxies (for either HTTP or HTTPS with CONNECT).
//
// By default, Transport caches connections for future re-use.
// This may leave many open connections when accessing many hosts.
// This behavior can be managed using Transport's CloseIdleConnections method
// and the MaxIdleConnsPerHost and DisableKeepAlives fields.
//
// Transports should be reused instead of created as needed.
// Transports are safe for concurrent use by multiple goroutines.
//
// A Transport is a low-level primitive for making HTTP and HTTPS requests.
// For high-level functionality, such as cookies and redirects, see Client.
//
// Transport uses HTTP/1.1 for HTTP URLs and either HTTP/1.1 or HTTP/2
// for HTTPS URLs, depending on whether the server supports HTTP/2,
// and how the Transport is configured. The DefaultTransport supports HTTP/2.
// To explicitly enable HTTP/2 on a transport, use golang.org/x/net/http2
// and call ConfigureTransport. See the package docs for more about HTTP/2.
//
// Responses with status codes in the 1xx range are either handled
// automatically (100 expect-continue) or ignored. The one
// exception is HTTP status code 101 (Switching Protocols), which is
// considered a terminal status and returned by RoundTrip. To see the
// ignored 1xx responses, use the httptrace trace package's
// ClientTrace.Got1xxResponse.
//
// Transport only retries a request upon encountering a network error
// if the request is idempotent and either has no body or has its
// Request.GetBody defined. HTTP requests are considered idempotent if
// they have HTTP methods GET, HEAD, OPTIONS, or TRACE; or if their
// Header map contains an "Idempotency-Key" or "X-Idempotency-Key"
// entry. If the idempotency key value is a zero-length slice, the
// request is treated as idempotent but the header is not sent on the
// wire.
type Transport struct {
	idleMu       sync.Mutex
	closeIdle    bool                                // user has requested to close all idle conns
	idleConn     map[connectMethodKey][]*persistConn // most recently used at end
	idleConnWait map[connectMethodKey]wantConnQueue  // waiting getConns
	idleLRU      connLRU

	reqMu       sync.Mutex
	reqCanceler map[cancelKey]func(error)

	connsPerHostMu   sync.Mutex
	connsPerHost     map[connectMethodKey]int
	connsPerHostWait map[connectMethodKey]wantConnQueue // waiting getConns
	altSvcJar        altsvc.Jar
	pendingAltSvcs   map[string]*pendingAltSvc
	pendingAltSvcsMu sync.Mutex

	// Force using specific http version
	forceHttpVersion httpVersion

	transport.Options

	t2 *http2.Transport // non-nil if http2 wired up
	t3 *http3.RoundTripper

	// disableAutoDecode, if true, prevents auto detect response
	// body's charset and decode it to utf-8
	disableAutoDecode bool

	// autoDecodeContentType specifies an optional function for determine
	// whether the response body should been auto decode to utf-8.
	// Only valid when DisableAutoDecode is true.
	autoDecodeContentType func(contentType string) bool
	wrappedRoundTrip      http.RoundTripper
	httpRoundTripWrappers []HttpRoundTripWrapper
}

// NewTransport is an alias of T
func NewTransport() *Transport {
	return T()
}

// T create a Transport.
func T() *Transport {
	t := &Transport{
		Options: transport.Options{
			Proxy:                 http.ProxyFromEnvironment,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig:       &tls.Config{NextProtos: []string{"http/1.1", "h2"}},
		},
	}
	t.t2 = &http2.Transport{Options: &t.Options}
	return t
}

// HttpRoundTripFunc is a http.RoundTripper implementation, which is a simple function.
type HttpRoundTripFunc func(req *http.Request) (resp *http.Response, err error)

// RoundTrip implements http.RoundTripper.
func (fn HttpRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

// HttpRoundTripWrapper is transport middleware function.
type HttpRoundTripWrapper func(rt http.RoundTripper) http.RoundTripper

// HttpRoundTripWrapperFunc is transport middleware function, more convenient than HttpRoundTripWrapper.
type HttpRoundTripWrapperFunc func(rt http.RoundTripper) HttpRoundTripFunc

func (f HttpRoundTripWrapperFunc) wrapper() HttpRoundTripWrapper {
	return func(rt http.RoundTripper) http.RoundTripper {
		return f(rt)
	}
}

// WrapRoundTripFunc adds a transport middleware function that will give the caller
// an opportunity to wrap the underlying http.RoundTripper.
func (t *Transport) WrapRoundTripFunc(funcs ...HttpRoundTripWrapperFunc) *Transport {
	var wrappers []HttpRoundTripWrapper
	for _, fn := range funcs {
		wrappers = append(wrappers, fn.wrapper())
	}
	return t.WrapRoundTrip(wrappers...)
}

// WrapRoundTrip adds a transport middleware function that will give the caller
// an opportunity to wrap the underlying http.RoundTripper.
func (t *Transport) WrapRoundTrip(wrappers ...HttpRoundTripWrapper) *Transport {
	if len(wrappers) == 0 {
		return t
	}
	if t.wrappedRoundTrip == nil {
		t.httpRoundTripWrappers = wrappers
		fn := func(req *http.Request) (*http.Response, error) {
			return t.roundTrip(req)
		}
		t.wrappedRoundTrip = HttpRoundTripFunc(fn)
	} else {
		t.httpRoundTripWrappers = append(t.httpRoundTripWrappers, wrappers...)
	}

	for _, w := range wrappers {
		t.wrappedRoundTrip = w(t.wrappedRoundTrip)
	}
	return t
}

// DisableAutoDecode disable auto-detect charset and decode to utf-8
// (enabled by default).
func (t *Transport) DisableAutoDecode() *Transport {
	t.disableAutoDecode = true
	return t
}

// EnableAutoDecode enable auto-detect charset and decode to utf-8
// (enabled by default).
func (t *Transport) EnableAutoDecode() *Transport {
	t.disableAutoDecode = false
	return t
}

// SetAutoDecodeContentTypeFunc set the function that determines whether the
// specified `Content-Type` should be auto-detected and decode to utf-8.
func (t *Transport) SetAutoDecodeContentTypeFunc(fn func(contentType string) bool) *Transport {
	t.autoDecodeContentType = fn
	return t
}

// SetAutoDecodeAllContentType enable try auto-detect charset and decode all
// content type to utf-8.
func (t *Transport) SetAutoDecodeAllContentType() *Transport {
	t.autoDecodeContentType = func(contentType string) bool {
		return true
	}
	return t
}

// SetAutoDecodeContentType set the content types that will be auto-detected and decode
// to utf-8 (e.g. "json", "xml", "html", "text").
func (t *Transport) SetAutoDecodeContentType(contentTypes ...string) {
	t.autoDecodeContentType = autoDecodeContentTypeFunc(contentTypes...)
}

// GetMaxIdleConns returns MaxIdleConns.
func (t *Transport) GetMaxIdleConns() int {
	return t.MaxIdleConns
}

// SetMaxIdleConns set the MaxIdleConns, which controls the maximum number of idle (keep-alive)
// connections across all hosts. Zero means no limit.
func (t *Transport) SetMaxIdleConns(max int) *Transport {
	t.MaxIdleConns = max
	return t
}

// SetMaxConnsPerHost set the MaxConnsPerHost, optionally limits the
// total number of connections per host, including connections in the
// dialing, active, and idle states. On limit violation, dials will block.
//
// Zero means no limit.
func (t *Transport) SetMaxConnsPerHost(max int) *Transport {
	t.MaxConnsPerHost = max
	return t
}

// SetIdleConnTimeout set the IdleConnTimeout, which  is the maximum
// amount of time an idle (keep-alive) connection will remain idle before
// closing itself.
//
// Zero means no limit.
func (t *Transport) SetIdleConnTimeout(timeout time.Duration) *Transport {
	t.IdleConnTimeout = timeout
	return t
}

// SetTLSHandshakeTimeout set the TLSHandshakeTimeout, which specifies the
// maximum amount of time waiting to wait for a TLS handshake.
//
// Zero means no timeout.
func (t *Transport) SetTLSHandshakeTimeout(timeout time.Duration) *Transport {
	t.TLSHandshakeTimeout = timeout
	return t
}

// SetResponseHeaderTimeout set the ResponseHeaderTimeout, if non-zero, specifies
// the amount of time to wait for a server's response headers after fully writing
// the request (including its body, if any). This time does not include the time
// to read the response body.
func (t *Transport) SetResponseHeaderTimeout(timeout time.Duration) *Transport {
	t.ResponseHeaderTimeout = timeout
	return t
}

// SetExpectContinueTimeout set the ExpectContinueTimeout, if non-zero, specifies
// the amount of time to wait for a server's first response headers after fully
// writing the request headers if the request has an "Expect: 100-continue" header.
// Zero means no timeout and causes the body to be sent immediately, without waiting
// for the server to approve.
// This time does not include the time to send the request header.
func (t *Transport) SetExpectContinueTimeout(timeout time.Duration) *Transport {
	t.ExpectContinueTimeout = timeout
	return t
}

// SetGetProxyConnectHeader set the GetProxyConnectHeader, which optionally specifies a func
// to return headers to send to proxyURL during a CONNECT request to the ip:port target.
// If it returns an error, the Transport's RoundTrip fails with that error. It can
// return (nil, nil) to not add headers.
// If GetProxyConnectHeader is non-nil, ProxyConnectHeader is ignored.
func (t *Transport) SetGetProxyConnectHeader(fn func(ctx context.Context, proxyURL *url.URL, target string) (http.Header, error)) *Transport {
	t.GetProxyConnectHeader = fn
	return t
}

// SetProxyConnectHeader set the ProxyConnectHeader, which optionally specifies headers to
// send to proxies during CONNECT requests.
// To set the header dynamically, see SetGetProxyConnectHeader.
func (t *Transport) SetProxyConnectHeader(header http.Header) *Transport {
	t.ProxyConnectHeader = header
	return t
}

// SetReadBufferSize set the ReadBufferSize, which specifies the size of the read buffer used
// when reading from the transport.
// If zero, a default (currently 4KB) is used.
func (t *Transport) SetReadBufferSize(size int) *Transport {
	t.ReadBufferSize = size
	return t
}

// SetWriteBufferSize set the WriteBufferSize, which specifies the size of the write buffer used
// when writing to the transport.
// If zero, a default (currently 4KB) is used.
func (t *Transport) SetWriteBufferSize(size int) *Transport {
	t.WriteBufferSize = size
	return t
}

// SetMaxResponseHeaderBytes set the MaxResponseHeaderBytes, which specifies a limit on how many
// response bytes are allowed in the server's response header.
//
// Zero means to use a default limit.
func (t *Transport) SetMaxResponseHeaderBytes(max int64) *Transport {
	t.MaxResponseHeaderBytes = max
	return t
}

// SetHTTP2MaxHeaderListSize set the http2 MaxHeaderListSize,
// which is the http2 SETTINGS_MAX_HEADER_LIST_SIZE to
// send in the initial settings frame. It is how many bytes
// of response headers are allowed. Unlike the http2 spec, zero here
// means to use a default limit (currently 10MB). If you actually
// want to advertise an unlimited value to the peer, Transport
// interprets the highest possible value here (0xffffffff or 1<<32-1)
// to mean no limit.
func (t *Transport) SetHTTP2MaxHeaderListSize(max uint32) *Transport {
	t.t2.MaxHeaderListSize = max
	return t
}

// SetHTTP2StrictMaxConcurrentStreams set the http2
// StrictMaxConcurrentStreams, which controls whether the
// server's SETTINGS_MAX_CONCURRENT_STREAMS should be respected
// globally. If false, new TCP connections are created to the
// server as needed to keep each under the per-connection
// SETTINGS_MAX_CONCURRENT_STREAMS limit. If true, the
// server's SETTINGS_MAX_CONCURRENT_STREAMS is interpreted as
// a global limit and callers of RoundTrip block when needed,
// waiting for their turn.
func (t *Transport) SetHTTP2StrictMaxConcurrentStreams(strict bool) *Transport {
	t.t2.StrictMaxConcurrentStreams = strict
	return t
}

// SetHTTP2ReadIdleTimeout set the http2 ReadIdleTimeout,
// which is the timeout after which a health check using ping
// frame will be carried out if no frame is received on the connection.
// Note that a ping response will is considered a received frame, so if
// there is no other traffic on the connection, the health check will
// be performed every ReadIdleTimeout interval.
// If zero, no health check is performed.
func (t *Transport) SetHTTP2ReadIdleTimeout(timeout time.Duration) *Transport {
	t.t2.ReadIdleTimeout = timeout
	return t
}

// SetHTTP2PingTimeout set the http2 PingTimeout, which is the timeout
// after which the connection will be closed if a response to Ping is
// not received.
// Defaults to 15s
func (t *Transport) SetHTTP2PingTimeout(timeout time.Duration) *Transport {
	t.t2.PingTimeout = timeout
	return t
}

// SetHTTP2WriteByteTimeout set the http2 WriteByteTimeout, which is the
// timeout after which the connection will be closed no data can be written
// to it. The timeout begins when data is available to write, and is
// extended whenever any bytes are written.
func (t *Transport) SetHTTP2WriteByteTimeout(timeout time.Duration) *Transport {
	t.t2.WriteByteTimeout = timeout
	return t
}

// SetTLSClientConfig set the custom TLSClientConfig, which specifies the TLS configuration to
// use with tls.Client.
// If nil, the default configuration is used.
// If non-nil, HTTP/2 support may not be enabled by default.
func (t *Transport) SetTLSClientConfig(cfg *tls.Config) *Transport {
	t.TLSClientConfig = cfg
	return t
}

// SetDebug set the optional debug function.
func (t *Transport) SetDebug(debugf func(format string, v ...interface{})) *Transport {
	t.Debugf = debugf
	return t
}

// SetProxy set the http proxy, only valid for HTTP1 and HTTP2, which specifies a function
// to return a proxy for a given Request. If the function returns a non-nil error, the request
// is aborted with the provided error.
//
// The proxy type is determined by the URL scheme. "http",
// "https", and "socks5" are supported. If the scheme is empty,
// "http" is assumed.
//
// If Proxy is nil or returns a nil *URL, no proxy is used.
func (t *Transport) SetProxy(proxy func(*http.Request) (*url.URL, error)) *Transport {
	t.Proxy = proxy
	return t
}

// SetDial set the custom DialContext function, only valid for HTTP1 and HTTP2, which specifies the
// dial function for creating unencrypted TCP connections.
// If it is nil, then the transport dials using package net.
//
// The dial function runs concurrently with calls to RoundTrip.
// A RoundTrip call that initiates a dial may end up using a connection dialed previously when the
// earlier connection becomes idle before the later dial function completes.
func (t *Transport) SetDial(fn func(ctx context.Context, network, addr string) (net.Conn, error)) *Transport {
	t.DialContext = fn
	return t
}

// SetDialTLS set the custom DialTLSContext function, only valid for HTTP1 and HTTP2, which specifies
// an optional dial function for creating TLS connections for non-proxied HTTPS requests (proxy will
// not work if set).
//
// If it is nil, DialContext and TLSClientConfig are used.
//
// If it is set, the function that set in SetDial is not used for HTTPS requests and the TLSClientConfig
// and TLSHandshakeTimeout are ignored. The returned net.Conn is assumed to already be past the TLS handshake.
func (t *Transport) SetDialTLS(fn func(ctx context.Context, network, addr string) (net.Conn, error)) *Transport {
	t.DialTLSContext = fn
	return t
}

// SetTLSHandshake set the custom tls handshake function, only valid for HTTP1 and HTTP2, not HTTP3,
// it specifies an optional dial function for tls handshake, it works even if a proxy is set, can be
// used to customize the tls fingerprint.
func (t *Transport) SetTLSHandshake(fn func(ctx context.Context, addr string, plainConn net.Conn) (conn net.Conn, tlsState *tls.ConnectionState, err error)) *Transport {
	t.TLSHandshakeContext = fn
	return t
}

type pendingAltSvc struct {
	CurrentIndex int
	Entries      []*altsvc.AltSvc
	Mu           sync.Mutex
	LastTime     time.Time
	Transport    http.RoundTripper
}

// EnableForceHTTP1 enable force using HTTP1 (disabled by default).
func (t *Transport) EnableForceHTTP1() *Transport {
	t.forceHttpVersion = h1
	return t
}

// EnableForceHTTP2 enable force using HTTP2 for https requests
// (disabled by default).
func (t *Transport) EnableForceHTTP2() *Transport {
	t.forceHttpVersion = h2
	return t
}

// EnableH2C enables HTTP2 over TCP without TLS.
func (t *Transport) EnableH2C() *Transport {
	t.Options.EnableH2C = true
	t.t2.AllowHTTP = true
	t.DialTLSContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return net.Dial(network, addr)
	}
	return t
}

// DisableH2C disables HTTP2 over TCP without TLS.
func (t *Transport) DisableH2C() *Transport {
	t.Options.EnableH2C = false
	t.t2.AllowHTTP = false
	t.t2.DialTLSContext = nil
	return t
}

// EnableForceHTTP3 enable force using HTTP3 for https requests
// (disabled by default).
func (t *Transport) EnableForceHTTP3() *Transport {
	t.EnableHTTP3()
	t.forceHttpVersion = h3
	return t
}

// DisableForceHttpVersion disable force using specified http
// version (disabled by default).
func (t *Transport) DisableForceHttpVersion() *Transport {
	t.forceHttpVersion = ""
	return t
}

func (t *Transport) DisableHTTP3() {
	t.altSvcJar = nil
	t.pendingAltSvcs = nil
	t.t3 = nil
}

func (t *Transport) EnableHTTP3() {
	if t.t3 != nil {
		return
	}

	v := runtime.Version()
	ss := strings.Split(v, ".")

	if len(ss) < 2 || ss[0] != "go1" {
		if t.Debugf != nil {
			t.Debugf("bad go version format: %s", v)
		}
		return
	}
	minorVersion, err := strconv.Atoi(ss[1])
	if err != nil {
		if t.Debugf != nil {
			t.Debugf("bad go minor version: %s", v)
		}
		return
	}
	if !(minorVersion >= 18 && minorVersion <= 20) {
		if t.Debugf != nil {
			t.Debugf("%s is not support http3", v)
		}
		return
	}

	if t.altSvcJar == nil {
		t.altSvcJar = altsvc.NewAltSvcJar()
	}
	if t.pendingAltSvcs == nil {
		t.pendingAltSvcs = make(map[string]*pendingAltSvc)
	}
	t3 := &http3.RoundTripper{
		Options: &t.Options,
	}
	t.t3 = t3
}

type wrapResponseBodyKeyType int

const wrapResponseBodyKey wrapResponseBodyKeyType = iota

type wrapResponseBodyFunc func(rc io.ReadCloser) io.ReadCloser

func (t *Transport) handleResponseBody(res *http.Response, req *http.Request) {
	if wrap, ok := req.Context().Value(wrapResponseBodyKey).(wrapResponseBodyFunc); ok {
		t.wrapResponseBody(res, wrap)
	}
	t.autoDecodeResponseBody(res)
	dump.WrapResponseBodyIfNeeded(res, req, t.Dump)
}

var allowedProtocols = map[string]bool{
	"h3": true,
}

func (t *Transport) handleAltSvc(req *http.Request, value string) {
	addr := netutil.AuthorityKey(req.URL)
	as := t.altSvcJar.GetAltSvc(addr)
	if as != nil {
		return
	}

	t.pendingAltSvcsMu.Lock()
	defer t.pendingAltSvcsMu.Unlock()
	_, ok := t.pendingAltSvcs[addr]
	if ok {
		return
	}
	ass, err := altsvcutil.ParseHeader(value)
	if err != nil {
		if t.Debugf != nil {
			t.Debugf("failed to parse alt-svc header: %s", err.Error())
		}
		return
	}
	var entries []*altsvc.AltSvc
	for _, a := range ass {
		if allowedProtocols[a.Protocol] {
			entries = append(entries, a)
		}
	}
	if len(entries) > 0 {
		pas := &pendingAltSvc{
			Entries: entries,
		}
		t.pendingAltSvcs[addr] = pas
		go t.handlePendingAltSvc(req.URL, pas)
	}
}

func (t *Transport) handlePendingAltSvc(u *url.URL, pas *pendingAltSvc) {
	for i := pas.CurrentIndex; i < len(pas.Entries); i++ {
		switch pas.Entries[i].Protocol {
		case "h3": // only support h3 in alt-svc for now
			u2 := altsvcutil.ConvertURL(pas.Entries[i], u)
			hostname := u2.Host
			err := t.t3.AddConn(hostname)
			if err != nil {
				if t.Debugf != nil {
					t.Debugf("failed to get http3 connection: %s", err.Error())
				}
			} else {
				pas.CurrentIndex = i
				pas.Transport = t.t3
				if t.Debugf != nil {
					t.Debugf("detected that the server %s supports http3, will try to use http3 protocol in subsequent requests", hostname)
				}
				return
			}
		}
	}
}

func (t *Transport) wrapResponseBody(res *http.Response, wrap wrapResponseBodyFunc) {
	switch b := res.Body.(type) {
	case *gzipReader:
		b.body.body = wrap(b.body.body)
	case *http2.GzipReader:
		b.Body = wrap(b.Body)
	case *http3.GzipReader:
		b.Body = wrap(b.Body)
	default:
		res.Body = wrap(res.Body)
	}
}

func (t *Transport) autoDecodeResponseBody(res *http.Response) {
	if t.disableAutoDecode {
		return
	}
	contentType := res.Header.Get("Content-Type")
	var shouldDecode func(contentType string) bool
	if t.autoDecodeContentType != nil {
		shouldDecode = t.autoDecodeContentType
	} else {
		shouldDecode = autoDecodeText
	}
	if !shouldDecode(contentType) {
		return
	}
	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		if t.Debugf != nil {
			t.Debugf("failed to parse content type %q: %v", contentType, err)
		}
	} else if charset, ok := params["charset"]; ok {
		charset = strings.ToLower(charset)
		if strings.Contains(charset, "utf-8") || strings.Contains(charset, "utf8") { // do not decode utf-8
			return
		}
		enc, _ := htmlcharset.Lookup(charset)
		if enc == nil {
			enc, err = ianaindex.MIME.Encoding(charset)
			if err != nil || enc == nil {
				if t.Debugf != nil {
					t.Debugf("ignore charset %s which is detected in Content-Type but not supported", charset)
				}
				return
			}
		}
		if t.Debugf != nil {
			t.Debugf("charset %s detected in Content-Type, auto-decode to utf-8", charset)
		}
		decodeReader := enc.NewDecoder().Reader(res.Body)
		res.Body = &decodeReaderCloser{res.Body, decodeReader}
		return
	}
	res.Body = newAutoDecodeReadCloser(res.Body, t)
}

// A cancelKey is the key of the reqCanceler map.
// We wrap the *http.Request in this type since we want to use the original request,
// not any transient one created by roundTrip.
type cancelKey struct {
	req *http.Request
}

func (t *Transport) writeBufferSize() int {
	if t.WriteBufferSize > 0 {
		return t.WriteBufferSize
	}
	return 4 << 10
}

func (t *Transport) readBufferSize() int {
	if t.ReadBufferSize > 0 {
		return t.ReadBufferSize
	}
	return 4 << 10
}

// Clone returns a deep copy of t's exported fields.
func (t *Transport) Clone() *Transport {
	tt := &Transport{
		Options:               t.Options.Clone(),
		disableAutoDecode:     t.disableAutoDecode,
		autoDecodeContentType: t.autoDecodeContentType,
		forceHttpVersion:      t.forceHttpVersion,
		httpRoundTripWrappers: t.httpRoundTripWrappers,
	}
	if len(tt.httpRoundTripWrappers) > 0 { // clone transport middleware
		fn := func(req *http.Request) (*http.Response, error) {
			return tt.roundTrip(req)
		}
		tt.wrappedRoundTrip = HttpRoundTripFunc(fn)
		for _, w := range tt.httpRoundTripWrappers {
			tt.wrappedRoundTrip = w(tt.wrappedRoundTrip)
		}
	}
	if t.t2 != nil {
		tt.t2 = &http2.Transport{Options: &tt.Options}
	}
	if t.t3 != nil {
		tt.EnableHTTP3()
	}
	return tt
}

// EnableDump enables the dump for all requests with specified dump options.
func (t *Transport) EnableDump(opt *DumpOptions) {
	dump := newDumper(opt)
	t.Dump = dump
	go dump.Start()
}

// DisableDump disables the dump.
func (t *Transport) DisableDump() {
	if t.Dump != nil {
		t.Dump.Stop()
		t.Dump = nil
	}
}

func (t *Transport) hasCustomTLSDialer() bool {
	return t.DialTLSContext != nil
}

// transportRequest is a wrapper around a *http.Request that adds
// optional extra headers to write and stores any error to return
// from roundTrip.
type transportRequest struct {
	*http.Request                        // original request, not to be mutated
	extra         http.Header            // extra headers to write, or nil
	trace         *httptrace.ClientTrace // optional
	cancelKey     cancelKey

	mu  sync.Mutex // guards err
	err error      // first setError value for mapRoundTripError to consider
}

func (tr *transportRequest) extraHeaders() http.Header {
	if tr.extra == nil {
		tr.extra = make(http.Header)
	}
	return tr.extra
}

func (tr *transportRequest) setError(err error) {
	tr.mu.Lock()
	if tr.err == nil {
		tr.err = err
	}
	tr.mu.Unlock()
}

func (t *Transport) roundTripAltSvc(req *http.Request, as *altsvc.AltSvc) (resp *http.Response, err error) {
	r := req.Clone(req.Context())
	r.URL = altsvcutil.ConvertURL(as, req.URL)
	switch as.Protocol {
	case "h3":
		resp, err = t.t3.RoundTrip(r)
	case "h2":
		resp, err = t.t2.RoundTrip(r)
	default:
		// impossible!
		panic(fmt.Sprintf("unknown protocol %q", as.Protocol))
	}
	return
}

// roundTrip implements a http.RoundTripper over HTTP.
func (t *Transport) roundTrip(req *http.Request) (resp *http.Response, err error) {
	ctx := req.Context()
	trace := httptrace.ContextClientTrace(ctx)

	if req.URL == nil {
		closeBody(req)
		return nil, errors.New("http: nil Request.URL")
	}

	if t.altSvcJar != nil {
		addr := netutil.AuthorityKey(req.URL)
		pas, ok := t.pendingAltSvcs[addr]
		if ok {
			if pas.Transport != nil {
				pas.Mu.Lock()
				if pas.Transport != nil {
					pas.LastTime = time.Now()
					r := req.Clone(req.Context())
					r.URL = altsvcutil.ConvertURL(pas.Entries[pas.CurrentIndex], req.URL)
					resp, err = pas.Transport.RoundTrip(r)
					if err != nil {
						pas.Transport = nil
						if pas.CurrentIndex+1 < len(pas.Entries) {
							pas.CurrentIndex++
							go t.handlePendingAltSvc(req.URL, pas)
						}
					} else {
						t.altSvcJar.SetAltSvc(addr, pas.Entries[pas.CurrentIndex])
						delete(t.pendingAltSvcs, addr)
					}
				}
				pas.Mu.Unlock()
				return
			}
		}
		as := t.altSvcJar.GetAltSvc(addr)
		if as != nil {
			return t.roundTripAltSvc(req, as)
		}
	}

	if req.Header == nil {
		closeBody(req)
		return nil, errors.New("http: nil Request.Header")
	}
	scheme := req.URL.Scheme
	isHTTP := scheme == "http" || scheme == "https"
	if isHTTP {
		for k, vv := range req.Header {
			if !httpguts.ValidHeaderFieldName(k) {
				closeBody(req)
				return nil, fmt.Errorf("net/http: invalid header field name %q", k)
			}
			for _, v := range vv {
				if !httpguts.ValidHeaderFieldValue(v) {
					closeBody(req)
					// Don't include the value in the error, because it may be sensitive.
					return nil, fmt.Errorf("net/http: invalid header field value for %q", k)
				}
			}
		}
	}

	if t.forceHttpVersion != "" {
		switch t.forceHttpVersion {
		case h3:
			return t.t3.RoundTrip(req)
		case h2:
			return t.t2.RoundTrip(req)
		}
	}

	origReq := req
	cancelKey := cancelKey{origReq}
	req = setupRewindBody(req)

	if scheme == "https" && t.forceHttpVersion != h1 {
		resp, err := t.t2.RoundTripOnlyCachedConn(req)
		if err != http2.ErrNoCachedConn {
			return resp, err
		}
		req, err = rewindBody(req)
		if err != nil {
			return nil, err
		}
		if t.t3 != nil {
			resp, err = t.t3.RoundTripOnlyCachedConn(req)
			if err != http3.ErrNoCachedConn {
				return resp, err
			}
			req, err = rewindBody(req)
			if err != nil {
				return nil, err
			}
		}
	}

	if !isHTTP {
		closeBody(req)
		return nil, badStringError("unsupported protocol scheme", scheme)
	}
	if req.Method != "" && !validMethod(req.Method) {
		closeBody(req)
		return nil, fmt.Errorf("net/http: invalid method %q", req.Method)
	}
	if req.URL.Host == "" {
		closeBody(req)
		return nil, errors.New("http: no Host in request URL")
	}

	for {
		select {
		case <-ctx.Done():
			closeBody(req)
			return nil, ctx.Err()
		default:
		}

		// treq gets modified by roundTrip, so we need to recreate for each retry.
		treq := &transportRequest{Request: req, trace: trace, cancelKey: cancelKey}
		cm, err := t.connectMethodForRequest(treq)
		if err != nil {
			closeBody(req)
			return nil, err
		}

		// Get the cached or newly-created connection to either the
		// host (for http or https), the http proxy, or the http proxy
		// pre-CONNECTed to https server. In any case, we'll be ready
		// to send it requests.
		pconn, err := t.getConn(treq, cm)
		if err != nil {
			t.setReqCanceler(cancelKey, nil)
			closeBody(req)
			return nil, err
		}

		var resp *http.Response
		if t.forceHttpVersion != h1 && pconn.alt != nil {
			// HTTP/2 path.
			t.setReqCanceler(cancelKey, nil) // not cancelable with CancelRequest
			resp, err = pconn.alt.RoundTrip(req)
		} else {
			resp, err = pconn.roundTrip(treq)
		}
		if err == nil {
			resp.Request = origReq
			return resp, nil
		}

		// Failed. Clean up and determine whether to retry.
		if http2.IsNoCachedConnError(err) {
			if t.removeIdleConn(pconn) {
				t.decConnsPerHost(pconn.cacheKey)
			}
		} else if !pconn.shouldRetryRequest(req, err) {
			// Issue 16465: return underlying net.Conn.Read error from peek,
			// as we've historically done.
			if e, ok := err.(nothingWrittenError); ok {
				err = e.error
			}
			if e, ok := err.(transportReadFromServerError); ok {
				err = e.err
			}
			if b, ok := req.Body.(*readTrackingBody); ok && !b.didClose {
				// Issue 49621: Close the request body if pconn.roundTrip
				// didn't do so already. This can happen if the pconn
				// write loop exits without reading the write request.
				closeBody(req)
			}
			return nil, err
		}
		testHookRoundTripRetried()

		// Rewind the body if we're able to.
		req, err = rewindBody(req)
		if err != nil {
			return nil, err
		}
	}
}

var errCannotRewind = errors.New("net/http: cannot rewind body after connection loss")

type readTrackingBody struct {
	io.ReadCloser
	didRead  bool
	didClose bool
}

func (r *readTrackingBody) Read(data []byte) (int, error) {
	r.didRead = true
	return r.ReadCloser.Read(data)
}

func (r *readTrackingBody) Close() error {
	r.didClose = true
	return r.ReadCloser.Close()
}

// setupRewindBody returns a new request with a custom body wrapper
// that can report whether the body needs rewinding.
// This lets rewindBody avoid an error result when the request
// does not have GetBody but the body hasn't been read at all yet.
func setupRewindBody(req *http.Request) *http.Request {
	if req.Body == nil || req.Body == NoBody {
		return req
	}
	newReq := *req
	newReq.Body = &readTrackingBody{ReadCloser: req.Body}
	return &newReq
}

// rewindBody returns a new request with the body rewound.
// It returns req unmodified if the body does not need rewinding.
// rewindBody takes care of closing req.Body when appropriate
// (in all cases except when rewindBody returns req unmodified).
func rewindBody(req *http.Request) (rewound *http.Request, err error) {
	if req.Body == nil || req.Body == NoBody || (!req.Body.(*readTrackingBody).didRead && !req.Body.(*readTrackingBody).didClose) {
		return req, nil // nothing to rewind
	}
	if !req.Body.(*readTrackingBody).didClose {
		closeBody(req)
	}
	if req.GetBody == nil {
		return nil, errCannotRewind
	}
	body, err := req.GetBody()
	if err != nil {
		return nil, err
	}
	newReq := *req
	newReq.Body = &readTrackingBody{ReadCloser: body}
	return &newReq, nil
}

// shouldRetryRequest reports whether we should retry sending a failed
// HTTP request on a new connection. The non-nil input error is the
// error from roundTrip.
func (pc *persistConn) shouldRetryRequest(req *http.Request, err error) bool {
	if http2.IsNoCachedConnError(err) {
		// Issue 16582: if the user started a bunch of
		// requests at once, they can all pick the same conn
		// and violate the server's max concurrent streams.
		// Instead, match the HTTP/1 behavior for now and dial
		// again to get a new TCP connection, rather than failing
		// this request.
		return true
	}
	if err == errMissingHost {
		// User error.
		return false
	}
	if !pc.isReused() {
		// This was a fresh connection. There's no reason the server
		// should've hung up on us.
		//
		// Also, if we retried now, we could loop forever
		// creating new connections and retrying if the server
		// is just hanging up on us because it doesn't like
		// our request (as opposed to sending an error).
		return false
	}
	if _, ok := err.(nothingWrittenError); ok {
		// We never wrote anything, so it's safe to retry, if there's no body or we
		// can "rewind" the body with GetBody.
		return outgoingLength(req) == 0 || req.GetBody != nil
	}
	if !isReplayable(req) {
		// Don't retry non-idempotent requests.
		return false
	}
	if _, ok := err.(transportReadFromServerError); ok {
		// We got some non-EOF net.Conn.Read failure reading
		// the 1st response byte from the server.
		return true
	}
	if err == errServerClosedIdle {
		// The server replied with io.EOF while we were trying to
		// read the response. Probably an unfortunately keep-alive
		// timeout, just as the client was writing a request.
		return true
	}
	return false // conservatively
}

// CloseIdleConnections closes any connections which were previously
// connected from previous requests but are now sitting idle in
// a "keep-alive" state. It does not interrupt any connections currently
// in use.
func (t *Transport) CloseIdleConnections() {
	t.idleMu.Lock()
	m := t.idleConn
	t.idleConn = nil
	t.closeIdle = true // close newly idle connections
	t.idleLRU = connLRU{}
	t.idleMu.Unlock()
	for _, conns := range m {
		for _, pconn := range conns {
			pconn.close(errCloseIdleConns)
		}
	}
	if t2 := t.t2; t2 != nil {
		t2.CloseIdleConnections()
	}
}

// CancelRequest cancels an in-flight request by closing its connection.
// CancelRequest should only be called after RoundTrip has returned.
//
// Deprecated: Use Request.WithContext to create a request with a
// cancelable context instead. CancelRequest cannot cancel HTTP/2
// requests.
func (t *Transport) CancelRequest(req *http.Request) {
	t.cancelRequest(cancelKey{req}, common.ErrRequestCanceled)
}

// Cancel an in-flight request, recording the error value.
// Returns whether the request was canceled.
func (t *Transport) cancelRequest(key cancelKey, err error) bool {
	// This function must not return until the cancel func has completed.
	// See: https://golang.org/issue/34658
	t.reqMu.Lock()
	defer t.reqMu.Unlock()
	cancel := t.reqCanceler[key]
	delete(t.reqCanceler, key)
	if cancel != nil {
		cancel(err)
	}

	return cancel != nil
}

// resetProxyConfig is used by tests.
func resetProxyConfig() {
}

func (t *Transport) connectMethodForRequest(treq *transportRequest) (cm connectMethod, err error) {
	cm.targetScheme = treq.URL.Scheme
	cm.targetAddr = canonicalAddr(treq.URL)
	if t.Proxy != nil {
		cm.proxyURL, err = t.Proxy(treq.Request)
	}
	cm.onlyH1 = t.forceHttpVersion == h1 || requestRequiresHTTP1(treq.Request)
	return cm, err
}

// proxyAuth returns the Proxy-Authorization header to set
// on requests, if applicable.
func (cm *connectMethod) proxyAuth() string {
	if cm.proxyURL == nil {
		return ""
	}
	if u := cm.proxyURL.User; u != nil {
		username := u.Username()
		password, _ := u.Password()
		return "Basic " + basicAuth(username, password)
	}
	return ""
}

// error values for debugging and testing, not seen by users.
var (
	errKeepAlivesDisabled = errors.New("http: putIdleConn: keep alives disabled")
	errConnBroken         = errors.New("http: putIdleConn: connection is in bad state")
	errCloseIdle          = errors.New("http: putIdleConn: CloseIdleConnections was called")
	errTooManyIdle        = errors.New("http: putIdleConn: too many idle connections")
	errTooManyIdleHost    = errors.New("http: putIdleConn: too many idle connections for host")
	errCloseIdleConns     = errors.New("http: CloseIdleConnections called")
	errReadLoopExiting    = errors.New("http: persistConn.readLoop exiting")
	errIdleConnTimeout    = errors.New("http: idle connection timeout")

	// errServerClosedIdle is not seen by users for idempotent requests, but may be
	// seen by a user if the server shuts down an idle connection and sends its FIN
	// in flight with already-written POST body bytes from the client.
	// See https://github.com/golang/go/issues/19943#issuecomment-355607646
	errServerClosedIdle = errors.New("http: server closed idle connection")
)

// transportReadFromServerError is used by Transport.readLoop when the
// 1 byte peek read fails and we're actually anticipating a response.
// Usually this is just due to the inherent keep-alive shut down race,
// where the server closed the connection at the same time the client
// wrote. The underlying err field is usually io.EOF or some
// ECONNRESET sort of thing which varies by platform. But it might be
// the user's custom net.Conn.Read error too, so we carry it along for
// them to return from Transport.RoundTrip.
type transportReadFromServerError struct {
	err error
}

func (e transportReadFromServerError) Unwrap() error { return e.err }

func (e transportReadFromServerError) Error() string {
	return fmt.Sprintf("net/http: Transport failed to read from server: %v", e.err)
}

func (t *Transport) putOrCloseIdleConn(pconn *persistConn) {
	if err := t.tryPutIdleConn(pconn); err != nil {
		pconn.close(err)
	}
}

func (t *Transport) maxIdleConnsPerHost() int {
	if v := t.MaxIdleConnsPerHost; v != 0 {
		return v
	}
	return defaultMaxIdleConnsPerHost
}

// tryPutIdleConn adds pconn to the list of idle persistent connections awaiting
// a new request.
// If pconn is no longer needed or not in a good state, tryPutIdleConn returns
// an error explaining why it wasn't registered.
// tryPutIdleConn does not close pconn. Use putOrCloseIdleConn instead for that.
func (t *Transport) tryPutIdleConn(pconn *persistConn) error {
	if t.DisableKeepAlives || t.MaxIdleConnsPerHost < 0 {
		return errKeepAlivesDisabled
	}
	if pconn.isBroken() {
		return errConnBroken
	}
	pconn.markReused()

	t.idleMu.Lock()
	defer t.idleMu.Unlock()

	// HTTP/2 (pconn.alt != nil) connections do not come out of the idle list,
	// because multiple goroutines can use them simultaneously.
	// If this is an HTTP/2 connection being “returned,” we're done.
	if pconn.alt != nil && t.idleLRU.m[pconn] != nil {
		return nil
	}

	// Deliver pconn to goroutine waiting for idle connection, if any.
	// (They may be actively dialing, but this conn is ready first.
	// Chrome calls this socket late binding.
	// See https://www.chromium.org/developers/design-documents/network-stack#TOC-Connection-Management.)
	key := pconn.cacheKey
	if q, ok := t.idleConnWait[key]; ok {
		done := false
		if pconn.alt == nil {
			// HTTP/1.
			// Loop over the waiting list until we find a w that isn't done already, and hand it pconn.
			for q.len() > 0 {
				w := q.popFront()
				if w.tryDeliver(pconn, nil) {
					done = true
					break
				}
			}
		} else {
			// HTTP/2.
			// Can hand the same pconn to everyone in the waiting list,
			// and we still won't be done: we want to put it in the idle
			// list unconditionally, for any future clients too.
			for q.len() > 0 {
				w := q.popFront()
				w.tryDeliver(pconn, nil)
			}
		}
		if q.len() == 0 {
			delete(t.idleConnWait, key)
		} else {
			t.idleConnWait[key] = q
		}
		if done {
			return nil
		}
	}

	if t.closeIdle {
		return errCloseIdle
	}
	if t.idleConn == nil {
		t.idleConn = make(map[connectMethodKey][]*persistConn)
	}
	idles := t.idleConn[key]
	if len(idles) >= t.maxIdleConnsPerHost() {
		return errTooManyIdleHost
	}
	for _, exist := range idles {
		if exist == pconn {
			log.Fatalf("dup idle pconn %p in freelist", pconn)
		}
	}
	t.idleConn[key] = append(idles, pconn)
	t.idleLRU.add(pconn)
	if t.MaxIdleConns != 0 && t.idleLRU.len() > t.MaxIdleConns {
		oldest := t.idleLRU.removeOldest()
		oldest.close(errTooManyIdle)
		t.removeIdleConnLocked(oldest)
	}

	// Set idle timer, but only for HTTP/1 (pconn.alt == nil).
	// The HTTP/2 implementation manages the idle timer itself
	// (see idleConnTimeout in h2_bundle.go).
	if t.IdleConnTimeout > 0 && pconn.alt == nil {
		if pconn.idleTimer != nil {
			pconn.idleTimer.Reset(t.IdleConnTimeout)
		} else {
			pconn.idleTimer = time.AfterFunc(t.IdleConnTimeout, pconn.closeConnIfStillIdle)
		}
	}
	pconn.idleAt = time.Now()
	return nil
}

// queueForIdleConn queues w to receive the next idle connection for w.cm.
// As an optimization hint to the caller, queueForIdleConn reports whether
// it successfully delivered an already-idle connection.
func (t *Transport) queueForIdleConn(w *wantConn) (delivered bool) {
	if t.DisableKeepAlives {
		return false
	}

	t.idleMu.Lock()
	defer t.idleMu.Unlock()

	// Stop closing connections that become idle - we might want one.
	// (That is, undo the effect of t.CloseIdleConnections.)
	t.closeIdle = false

	if w == nil {
		// Happens in test hook.
		return false
	}

	// If IdleConnTimeout is set, calculate the oldest
	// persistConn.idleAt time we're willing to use a cached idle
	// conn.
	var oldTime time.Time
	if t.IdleConnTimeout > 0 {
		oldTime = time.Now().Add(-t.IdleConnTimeout)
	}

	// Look for most recently-used idle connection.
	if list, ok := t.idleConn[w.key]; ok {
		stop := false
		delivered := false
		for len(list) > 0 && !stop {
			pconn := list[len(list)-1]

			// See whether this connection has been idle too long, considering
			// only the wall time (the Round(0)), in case this is a laptop or VM
			// coming out of suspend with previously cached idle connections.
			tooOld := !oldTime.IsZero() && pconn.idleAt.Round(0).Before(oldTime)
			if tooOld {
				// Async cleanup. Launch in its own goroutine (as if a
				// time.AfterFunc called it); it acquires idleMu, which we're
				// holding, and does a synchronous net.Conn.Close.
				go pconn.closeConnIfStillIdle()
			}
			if pconn.isBroken() || tooOld {
				// If either persistConn.readLoop has marked the connection
				// broken, but Transport.removeIdleConn has not yet removed it
				// from the idle list, or if this persistConn is too old (it was
				// idle too long), then ignore it and look for another. In both
				// cases it's already in the process of being closed.
				list = list[:len(list)-1]
				continue
			}
			delivered = w.tryDeliver(pconn, nil)
			if delivered {
				if pconn.alt != nil {
					// HTTP/2: multiple clients can share pconn.
					// Leave it in the list.
				} else {
					// HTTP/1: only one client can use pconn.
					// Remove it from the list.
					t.idleLRU.remove(pconn)
					list = list[:len(list)-1]
				}
			}
			stop = true
		}
		if len(list) > 0 {
			t.idleConn[w.key] = list
		} else {
			delete(t.idleConn, w.key)
		}
		if stop {
			return delivered
		}
	}

	// Register to receive next connection that becomes idle.
	if t.idleConnWait == nil {
		t.idleConnWait = make(map[connectMethodKey]wantConnQueue)
	}
	q := t.idleConnWait[w.key]
	q.cleanFront()
	q.pushBack(w)
	t.idleConnWait[w.key] = q
	return false
}

// removeIdleConn marks pconn as dead.
func (t *Transport) removeIdleConn(pconn *persistConn) bool {
	t.idleMu.Lock()
	defer t.idleMu.Unlock()
	return t.removeIdleConnLocked(pconn)
}

// t.idleMu must be held.
func (t *Transport) removeIdleConnLocked(pconn *persistConn) bool {
	if pconn.idleTimer != nil {
		pconn.idleTimer.Stop()
	}
	t.idleLRU.remove(pconn)
	key := pconn.cacheKey
	pconns := t.idleConn[key]
	var removed bool
	switch len(pconns) {
	case 0:
		// Nothing
	case 1:
		if pconns[0] == pconn {
			delete(t.idleConn, key)
			removed = true
		}
	default:
		for i, v := range pconns {
			if v != pconn {
				continue
			}
			// Slide down, keeping most recently-used
			// conns at the end.
			copy(pconns[i:], pconns[i+1:])
			t.idleConn[key] = pconns[:len(pconns)-1]
			removed = true
			break
		}
	}
	return removed
}

func (t *Transport) setReqCanceler(key cancelKey, fn func(error)) {
	t.reqMu.Lock()
	defer t.reqMu.Unlock()
	if t.reqCanceler == nil {
		t.reqCanceler = make(map[cancelKey]func(error))
	}
	if fn != nil {
		t.reqCanceler[key] = fn
	} else {
		delete(t.reqCanceler, key)
	}
}

// replaceReqCanceler replaces an existing cancel function. If there is no cancel function
// for the request, we don't set the function and return false.
// Since CancelRequest will clear the canceler, we can use the return value to detect if
// the request was canceled since the last setReqCancel call.
func (t *Transport) replaceReqCanceler(key cancelKey, fn func(error)) bool {
	t.reqMu.Lock()
	defer t.reqMu.Unlock()
	_, ok := t.reqCanceler[key]
	if !ok {
		return false
	}
	if fn != nil {
		t.reqCanceler[key] = fn
	} else {
		delete(t.reqCanceler, key)
	}
	return true
}

var zeroDialer net.Dialer

func (t *Transport) dial(ctx context.Context, network, addr string) (net.Conn, error) {
	if t.DialContext != nil {
		c, err := t.DialContext(ctx, network, addr)
		if c == nil && err == nil {
			err = errors.New("net/http: Transport.DialContext hook returned (nil, nil)")
		}
		return c, err
	}
	return zeroDialer.DialContext(ctx, network, addr)
}

// A wantConn records state about a wanted connection
// (that is, an active call to getConn).
// The conn may be gotten by dialing or by finding an idle connection,
// or a cancellation may make the conn no longer wanted.
// These three options are racing against each other and use
// wantConn to coordinate and agree about the winning outcome.
type wantConn struct {
	cm    connectMethod
	key   connectMethodKey // cm.key()
	ctx   context.Context  // context for dial
	ready chan struct{}    // closed when pc, err pair is delivered

	// hooks for testing to know when dials are done
	// beforeDial is called in the getConn goroutine when the dial is queued.
	// afterDial is called when the dial is completed or canceled.
	beforeDial func()
	afterDial  func()

	mu  sync.Mutex // protects pc, err, close(ready)
	pc  *persistConn
	err error
}

// waiting reports whether w is still waiting for an answer (connection or error).
func (w *wantConn) waiting() bool {
	select {
	case <-w.ready:
		return false
	default:
		return true
	}
}

// tryDeliver attempts to deliver pc, err to w and reports whether it succeeded.
func (w *wantConn) tryDeliver(pc *persistConn, err error) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.pc != nil || w.err != nil {
		return false
	}

	w.pc = pc
	w.err = err
	if w.pc == nil && w.err == nil {
		panic("net/http: internal error: misuse of tryDeliver")
	}
	close(w.ready)
	return true
}

// cancel marks w as no longer wanting a result (for example, due to cancellation).
// If a connection has been delivered already, cancel returns it with t.putOrCloseIdleConn.
func (w *wantConn) cancel(t *Transport, err error) {
	w.mu.Lock()
	if w.pc == nil && w.err == nil {
		close(w.ready) // catch misbehavior in future delivery
	}
	pc := w.pc
	w.pc = nil
	w.err = err
	w.mu.Unlock()

	if pc != nil {
		t.putOrCloseIdleConn(pc)
	}
}

// A wantConnQueue is a queue of wantConns.
type wantConnQueue struct {
	// This is a queue, not a deque.
	// It is split into two stages - head[headPos:] and tail.
	// popFront is trivial (headPos++) on the first stage, and
	// pushBack is trivial (append) on the second stage.
	// If the first stage is empty, popFront can swap the
	// first and second stages to remedy the situation.
	//
	// This two-stage split is analogous to the use of two lists
	// in Okasaki's purely functional queue but without the
	// overhead of reversing the list when swapping stages.
	head    []*wantConn
	headPos int
	tail    []*wantConn
}

// len returns the number of items in the queue.
func (q *wantConnQueue) len() int {
	return len(q.head) - q.headPos + len(q.tail)
}

// pushBack adds w to the back of the queue.
func (q *wantConnQueue) pushBack(w *wantConn) {
	q.tail = append(q.tail, w)
}

// popFront removes and returns the wantConn at the front of the queue.
func (q *wantConnQueue) popFront() *wantConn {
	if q.headPos >= len(q.head) {
		if len(q.tail) == 0 {
			return nil
		}
		// Pick up tail as new head, clear tail.
		q.head, q.headPos, q.tail = q.tail, 0, q.head[:0]
	}
	w := q.head[q.headPos]
	q.head[q.headPos] = nil
	q.headPos++
	return w
}

// peekFront returns the wantConn at the front of the queue without removing it.
func (q *wantConnQueue) peekFront() *wantConn {
	if q.headPos < len(q.head) {
		return q.head[q.headPos]
	}
	if len(q.tail) > 0 {
		return q.tail[0]
	}
	return nil
}

// cleanFront pops any wantConns that are no longer waiting from the head of the
// queue, reporting whether any were popped.
func (q *wantConnQueue) cleanFront() (cleaned bool) {
	for {
		w := q.peekFront()
		if w == nil || w.waiting() {
			return cleaned
		}
		q.popFront()
		cleaned = true
	}
}

func (t *Transport) customDialTLS(ctx context.Context, network, addr string) (conn net.Conn, err error) {
	conn, err = t.DialTLSContext(ctx, network, addr)

	if conn == nil && err == nil {
		err = errors.New("net/http: Transport.DialTLS or DialTLSContext returned (nil, nil)")
	}
	return
}

// getConn dials and creates a new persistConn to the target as
// specified in the connectMethod. This includes doing a proxy CONNECT
// and/or setting up TLS.  If this doesn't return an error, the persistConn
// is ready to write requests to.
func (t *Transport) getConn(treq *transportRequest, cm connectMethod) (pc *persistConn, err error) {
	req := treq.Request
	trace := treq.trace
	ctx := req.Context()
	if trace != nil && trace.GetConn != nil {
		trace.GetConn(cm.addr())
	}

	w := &wantConn{
		cm:         cm,
		key:        cm.key(),
		ctx:        ctx,
		ready:      make(chan struct{}, 1),
		beforeDial: testHookPrePendingDial,
		afterDial:  testHookPostPendingDial,
	}
	defer func() {
		if err != nil {
			w.cancel(t, err)
		}
	}()

	// Queue for idle connection.
	if delivered := t.queueForIdleConn(w); delivered {
		pc := w.pc
		// Trace only for HTTP/1.
		// HTTP/2 calls trace.GotConn itself.
		if pc.alt == nil && trace != nil && trace.GotConn != nil {
			trace.GotConn(pc.gotIdleConnTrace(pc.idleAt))
		}
		// set request canceler to some non-nil function so we
		// can detect whether it was cleared between now and when
		// we enter roundTrip
		t.setReqCanceler(treq.cancelKey, func(error) {})
		return pc, nil
	}

	cancelc := make(chan error, 1)
	t.setReqCanceler(treq.cancelKey, func(err error) { cancelc <- err })

	// Queue for permission to dial.
	t.queueForDial(w)

	// Wait for completion or cancellation.
	select {
	case <-w.ready:
		// Trace success but only for HTTP/1.
		// HTTP/2 calls trace.GotConn itself.
		if w.pc != nil && w.pc.alt == nil && trace != nil && trace.GotConn != nil {
			trace.GotConn(httptrace.GotConnInfo{Conn: w.pc.conn, Reused: w.pc.isReused()})
		}
		if w.err != nil {
			// If the request has been canceled, that's probably
			// what caused w.err; if so, prefer to return the
			// cancellation error (see golang.org/issue/16049).
			select {
			case <-req.Cancel:
				return nil, errRequestCanceledConn
			case <-req.Context().Done():
				return nil, req.Context().Err()
			case err := <-cancelc:
				if err == common.ErrRequestCanceled {
					err = errRequestCanceledConn
				}
				return nil, err
			default:
				// return below
			}
		}
		return w.pc, w.err
	case <-req.Cancel:
		return nil, errRequestCanceledConn
	case <-req.Context().Done():
		return nil, req.Context().Err()
	case err := <-cancelc:
		if err == common.ErrRequestCanceled {
			err = errRequestCanceledConn
		}
		return nil, err
	}
}

// queueForDial queues w to wait for permission to begin dialing.
// Once w receives permission to dial, it will do so in a separate goroutine.
func (t *Transport) queueForDial(w *wantConn) {
	w.beforeDial()
	if t.MaxConnsPerHost <= 0 {
		go t.dialConnFor(w)
		return
	}

	t.connsPerHostMu.Lock()
	defer t.connsPerHostMu.Unlock()

	if n := t.connsPerHost[w.key]; n < t.MaxConnsPerHost {
		if t.connsPerHost == nil {
			t.connsPerHost = make(map[connectMethodKey]int)
		}
		t.connsPerHost[w.key] = n + 1
		go t.dialConnFor(w)
		return
	}

	if t.connsPerHostWait == nil {
		t.connsPerHostWait = make(map[connectMethodKey]wantConnQueue)
	}
	q := t.connsPerHostWait[w.key]
	q.cleanFront()
	q.pushBack(w)
	t.connsPerHostWait[w.key] = q
}

// dialConnFor dials on behalf of w and delivers the result to w.
// dialConnFor has received permission to dial w.cm and is counted in t.connCount[w.cm.key()].
// If the dial is canceled or unsuccessful, dialConnFor decrements t.connCount[w.cm.key()].
func (t *Transport) dialConnFor(w *wantConn) {
	defer w.afterDial()

	pc, err := t.dialConn(w.ctx, w.cm)
	delivered := w.tryDeliver(pc, err)
	if err == nil && (!delivered || pc.alt != nil) {
		// pconn was not passed to w,
		// or it is HTTP/2 and can be shared.
		// Add to the idle connection pool.
		t.putOrCloseIdleConn(pc)
	}
	if err != nil {
		t.decConnsPerHost(w.key)
	}
}

// decConnsPerHost decrements the per-host connection count for key,
// which may in turn give a different waiting goroutine permission to dial.
func (t *Transport) decConnsPerHost(key connectMethodKey) {
	if t.MaxConnsPerHost <= 0 {
		return
	}

	t.connsPerHostMu.Lock()
	defer t.connsPerHostMu.Unlock()
	n := t.connsPerHost[key]
	if n == 0 {
		// Shouldn't happen, but if it does, the counting is buggy and could
		// easily lead to a silent deadlock, so report the problem loudly.
		panic("net/http: internal error: connCount underflow")
	}

	// Can we hand this count to a goroutine still waiting to dial?
	// (Some goroutines on the wait list may have timed out or
	// gotten a connection another way. If they're all gone,
	// we don't want to kick off any spurious dial operations.)
	if q := t.connsPerHostWait[key]; q.len() > 0 {
		done := false
		for q.len() > 0 {
			w := q.popFront()
			if w.waiting() {
				go t.dialConnFor(w)
				done = true
				break
			}
		}
		if q.len() == 0 {
			delete(t.connsPerHostWait, key)
		} else {
			// q is a value (like a slice), so we have to store
			// the updated q back into the map.
			t.connsPerHostWait[key] = q
		}
		if done {
			return
		}
	}

	// Otherwise, decrement the recorded count.
	if n--; n == 0 {
		delete(t.connsPerHost, key)
	} else {
		t.connsPerHost[key] = n
	}
}

// Add TLS to a persistent connection, i.e. negotiate a TLS session. If pconn is already a TLS
// tunnel, this function establishes a nested TLS session inside the encrypted channel.
// The remote endpoint's name may be overridden by TLSClientConfig.ServerName.
func (pc *persistConn) addTLS(ctx context.Context, name string, trace *httptrace.ClientTrace, forProxy bool) error {

	// Initiate TLS and check remote host name against certificate.
	cfg := cloneTLSConfig(pc.t.TLSClientConfig)
	if cfg.ServerName == "" {
		cfg.ServerName = name
	}
	if pc.cacheKey.onlyH1 {
		cfg.NextProtos = nil
	}
	plainConn := pc.conn
	tlsConn := tls.Client(plainConn, cfg)
	errc := make(chan error, 2)
	var timer *time.Timer // for canceling TLS handshake
	if d := pc.t.TLSHandshakeTimeout; d != 0 {
		timer = time.AfterFunc(d, func() {
			errc <- tlsHandshakeTimeoutError{}
		})
	}
	go func() {
		if trace != nil && trace.TLSHandshakeStart != nil {
			trace.TLSHandshakeStart()
		}
		err := tlsConn.HandshakeContext(ctx)
		if timer != nil {
			timer.Stop()
		}
		errc <- err
	}()
	if err := <-errc; err != nil {
		plainConn.Close()
		if trace != nil && trace.TLSHandshakeDone != nil {
			trace.TLSHandshakeDone(tls.ConnectionState{}, err)
		}
		return err
	}
	cs := tlsConn.ConnectionState()
	if trace != nil && trace.TLSHandshakeDone != nil {
		trace.TLSHandshakeDone(cs, nil)
	}
	pc.tlsState = &cs
	pc.conn = tlsConn
	if !forProxy && pc.t.forceHttpVersion == h2 && cs.NegotiatedProtocol != http2.NextProtoTLS {
		return newHttp2NotSupportedError(cs.NegotiatedProtocol)
	}
	return nil
}

func newHttp2NotSupportedError(negotiatedProtocol string) error {
	errMsg := "server does not support http2"
	if negotiatedProtocol != "" {
		errMsg += fmt.Sprintf(", you can use %s which is supported", negotiatedProtocol)
	}
	return errors.New(errMsg)
}

func (t *Transport) customTlsHandshake(ctx context.Context, trace *httptrace.ClientTrace, addr string, pconn *persistConn) error {
	errc := make(chan error, 2)
	var timer *time.Timer // for canceling TLS handshake
	if d := t.TLSHandshakeTimeout; d != 0 {
		timer = time.AfterFunc(d, func() {
			errc <- tlsHandshakeTimeoutError{}
		})
	}
	go func() {
		if trace != nil && trace.TLSHandshakeStart != nil {
			trace.TLSHandshakeStart()
		}
		conn, tlsState, err := t.TLSHandshakeContext(ctx, addr, pconn.conn)
		if err != nil {
			if timer != nil {
				timer.Stop()
			}
			if trace != nil && trace.TLSHandshakeDone != nil {
				trace.TLSHandshakeDone(tls.ConnectionState{}, err)
			}
		} else {
			pconn.conn = conn
			pconn.tlsState = tlsState
			if trace != nil && trace.TLSHandshakeDone != nil {
				trace.TLSHandshakeDone(*tlsState, nil)
			}
		}
		errc <- err
	}()
	if err := <-errc; err != nil {
		pconn.conn.Close()
		return err
	}
	return nil
}

func (t *Transport) dialConn(ctx context.Context, cm connectMethod) (pconn *persistConn, err error) {
	pconn = &persistConn{
		t:             t,
		cacheKey:      cm.key(),
		reqch:         make(chan requestAndChan, 1),
		writech:       make(chan writeRequest, 1),
		closech:       make(chan struct{}),
		writeErrCh:    make(chan error, 1),
		writeLoopDone: make(chan struct{}),
	}
	trace := httptrace.ContextClientTrace(ctx)
	wrapErr := func(err error) error {
		if cm.proxyURL != nil {
			// Return a typed error, per Issue 16997
			return &net.OpError{Op: "proxyconnect", Net: "tcp", Err: err}
		}
		return err
	}
	if cm.scheme() == "https" && t.hasCustomTLSDialer() {
		var err error
		pconn.conn, err = t.customDialTLS(ctx, "tcp", cm.addr())
		if err != nil {
			return nil, wrapErr(err)
		}
		if tc, ok := pconn.conn.(reqtls.Conn); ok {
			// Handshake here, in case DialTLS didn't. TLSNextProto below
			// depends on it for knowing the connection state.
			if trace != nil && trace.TLSHandshakeStart != nil {
				trace.TLSHandshakeStart()
			}
			if err := tc.HandshakeContext(ctx); err != nil {
				go pconn.conn.Close()
				if trace != nil && trace.TLSHandshakeDone != nil {
					trace.TLSHandshakeDone(tls.ConnectionState{}, err)
				}
				return nil, err
			}
			cs := tc.ConnectionState()
			if trace != nil && trace.TLSHandshakeDone != nil {
				trace.TLSHandshakeDone(cs, nil)
			}
			pconn.tlsState = &cs
			if cm.proxyURL == nil && pconn.t.forceHttpVersion == h2 && cs.NegotiatedProtocol != http2.NextProtoTLS {
				return nil, newHttp2NotSupportedError(cs.NegotiatedProtocol)
			}
		}
	} else {
		conn, err := t.dial(ctx, "tcp", cm.addr())
		if err != nil {
			return nil, wrapErr(err)
		}
		pconn.conn = conn
		if cm.scheme() == "https" {
			var firstTLSHost string
			if firstTLSHost, _, err = net.SplitHostPort(cm.addr()); err != nil {
				return nil, wrapErr(err)
			}
			if t.TLSHandshakeContext != nil && cm.proxyURL == nil {
				err = t.customTlsHandshake(ctx, trace, firstTLSHost, pconn)
				if err != nil {
					return nil, err
				}
			} else {
				if err = pconn.addTLS(ctx, firstTLSHost, trace, cm.proxyURL != nil); err != nil {
					return nil, wrapErr(err)
				}
			}
		}
	}

	if t.Debugf != nil && cm.proxyURL != nil {
		t.Debugf("connect %s via proxy %s", cm.targetAddr, cm.proxyURL.String())
	}

	// Proxy setup.
	switch {
	case cm.proxyURL == nil:
		// Do nothing. Not using a proxy.
	case cm.proxyURL.Scheme == "socks5":
		conn := pconn.conn
		d := socks.NewDialer("tcp", conn.RemoteAddr().String())
		if u := cm.proxyURL.User; u != nil {
			auth := &socks.UsernamePassword{
				Username: u.Username(),
			}
			auth.Password, _ = u.Password()
			d.AuthMethods = []socks.AuthMethod{
				socks.AuthMethodNotRequired,
				socks.AuthMethodUsernamePassword,
			}
			d.Authenticate = auth.Authenticate
		}
		if _, err := d.DialWithConn(ctx, conn, "tcp", cm.targetAddr); err != nil {
			conn.Close()
			return nil, err
		}
	case cm.targetScheme == "http":
		pconn.isProxy = true
		if pa := cm.proxyAuth(); pa != "" {
			pconn.mutateHeaderFunc = func(h http.Header) {
				h.Set("Proxy-Authorization", pa)
			}
		}
	case cm.targetScheme == "https":
		conn := pconn.conn
		var hdr http.Header
		if t.GetProxyConnectHeader != nil {
			var err error
			hdr, err = t.GetProxyConnectHeader(ctx, cm.proxyURL, cm.targetAddr)
			if err != nil {
				conn.Close()
				return nil, err
			}
		} else {
			hdr = t.ProxyConnectHeader
		}
		if hdr == nil {
			hdr = make(http.Header)
		}
		if pa := cm.proxyAuth(); pa != "" {
			hdr = hdr.Clone()
			hdr.Set("Proxy-Authorization", pa)
		}
		connectReq := &http.Request{
			Method: "CONNECT",
			URL:    &url.URL{Opaque: cm.targetAddr},
			Host:   cm.targetAddr,
			Header: hdr,
		}

		// If there's no done channel (no deadline or cancellation
		// from the caller possible), at least set some (long)
		// timeout here. This will make sure we don't block forever
		// and leak a goroutine if the connection stops replying
		// after the TCP connect.
		connectCtx := ctx
		if ctx.Done() == nil {
			newCtx, cancel := context.WithTimeout(ctx, 1*time.Minute)
			defer cancel()
			connectCtx = newCtx
		}

		didReadResponse := make(chan struct{}) // closed after CONNECT write+read is done or fails
		var (
			resp *http.Response
			err  error // write or read error
		)
		// Write the CONNECT request & read the response.
		go func() {
			defer close(didReadResponse)
			err = connectReq.Write(conn)
			if err != nil {
				return
			}
			// Okay to use and discard buffered reader here, because
			// TLS server will not speak until spoken to.
			br := bufio.NewReader(conn)
			resp, err = http.ReadResponse(br, connectReq)
		}()
		select {
		case <-connectCtx.Done():
			conn.Close()
			<-didReadResponse
			return nil, connectCtx.Err()
		case <-didReadResponse:
			// resp or err now set
		}
		if err != nil {
			conn.Close()
			return nil, err
		}

		if t.OnProxyConnectResponse != nil {
			err = t.OnProxyConnectResponse(ctx, cm.proxyURL, connectReq, resp)
			if err != nil {
				return nil, err
			}
		}

		if resp.StatusCode != 200 {
			_, text, ok := util.CutString(resp.Status, " ")
			conn.Close()
			if !ok {
				return nil, errors.New("unknown status code")
			}
			return nil, errors.New(text)
		}
	}

	if cm.proxyURL != nil && cm.targetScheme == "https" {
		if t.TLSHandshakeContext != nil {
			err := t.customTlsHandshake(ctx, trace, cm.tlsHost(), pconn)
			if err != nil {
				return nil, err
			}
		} else {
			if err := pconn.addTLS(ctx, cm.tlsHost(), trace, false); err != nil {
				return nil, err
			}
		}
	}

	if s := pconn.tlsState; t.forceHttpVersion != h1 && s != nil && s.NegotiatedProtocolIsMutual && s.NegotiatedProtocol != "" {
		if s.NegotiatedProtocol == http2.NextProtoTLS {
			if used, err := t.t2.AddConn(pconn.conn, cm.targetAddr); err != nil {
				go pconn.conn.Close()
				return nil, err
			} else if !used {
				go pconn.conn.Close()
			}
			return &persistConn{t: t, cacheKey: pconn.cacheKey, alt: t.t2}, nil
		}
	}

	pconn.br = bufio.NewReaderSize(pconn, t.readBufferSize())
	pconn.bw = bufio.NewWriterSize(persistConnWriter{pconn}, t.writeBufferSize())

	go pconn.readLoop()
	go pconn.writeLoop()
	return pconn, nil
}

// persistConnWriter is the io.Writer written to by pc.bw.
// It accumulates the number of bytes written to the underlying conn,
// so the retry logic can determine whether any bytes made it across
// the wire.
// This is exactly 1 pointer field wide so it can go into an interface
// without allocation.
type persistConnWriter struct {
	pc *persistConn
}

func (w persistConnWriter) Write(p []byte) (n int, err error) {
	n, err = w.pc.conn.Write(p)
	w.pc.nwrite += int64(n)
	return
}

// ReadFrom exposes persistConnWriter's underlying Conn to io.Copy and if
// the Conn implements io.ReaderFrom, it can take advantage of optimizations
// such as sendfile.
func (w persistConnWriter) ReadFrom(r io.Reader) (n int64, err error) {
	n, err = io.Copy(w.pc.conn, r)
	w.pc.nwrite += n
	return
}

var _ io.ReaderFrom = (*persistConnWriter)(nil)

// connectMethod is the map key (in its String form) for keeping persistent
// TCP connections alive for subsequent HTTP requests.
//
// A connect method may be of the following types:
//
//	connectMethod.key().String()      Description
//	------------------------------    -------------------------
//	|http|foo.com                     http directly to server, no proxy
//	|https|foo.com                    https directly to server, no proxy
//	|https,h1|foo.com                 https directly to server w/o HTTP/2, no proxy
//	http://proxy.com|https|foo.com    http to proxy, then CONNECT to foo.com
//	http://proxy.com|http             http to proxy, http to anywhere after that
//	socks5://proxy.com|http|foo.com   socks5 to proxy, then http to foo.com
//	socks5://proxy.com|https|foo.com  socks5 to proxy, then https to foo.com
//	https://proxy.com|https|foo.com   https to proxy, then CONNECT to foo.com
//	https://proxy.com|http            https to proxy, http to anywhere after that
type connectMethod struct {
	_            incomparable
	proxyURL     *url.URL // nil for no proxy, else full proxy URL
	targetScheme string   // "http" or "https"
	// If proxyURL specifies an http or https proxy, and targetScheme is http (not https),
	// then targetAddr is not included in the connect method key, because the socket can
	// be reused for different targetAddr values.
	targetAddr string
	onlyH1     bool // whether to disable HTTP/2 and force HTTP/1
}

func (cm *connectMethod) key() connectMethodKey {
	proxyStr := ""
	targetAddr := cm.targetAddr
	if cm.proxyURL != nil {
		proxyStr = cm.proxyURL.String()
		if (cm.proxyURL.Scheme == "http" || cm.proxyURL.Scheme == "https") && cm.targetScheme == "http" {
			targetAddr = ""
		}
	}
	return connectMethodKey{
		proxy:  proxyStr,
		scheme: cm.targetScheme,
		addr:   targetAddr,
		onlyH1: cm.onlyH1,
	}
}

// scheme returns the first hop scheme: http, https, or socks5
func (cm *connectMethod) scheme() string {
	if cm.proxyURL != nil {
		return cm.proxyURL.Scheme
	}
	return cm.targetScheme
}

// addr returns the first hop "host:port" to which we need to TCP connect.
func (cm *connectMethod) addr() string {
	if cm.proxyURL != nil {
		return canonicalAddr(cm.proxyURL)
	}
	return cm.targetAddr
}

// tlsHost returns the host name to match against the peer's
// TLS certificate.
func (cm *connectMethod) tlsHost() string {
	h := cm.targetAddr
	if hasPort(h) {
		h = h[:strings.LastIndex(h, ":")]
	}
	return h
}

// connectMethodKey is the map key version of connectMethod, with a
// stringified proxy URL (or the empty string) instead of a pointer to
// a URL.
type connectMethodKey struct {
	proxy, scheme, addr string
	onlyH1              bool
}

func (k connectMethodKey) String() string {
	// Only used by tests.
	var h1 string
	if k.onlyH1 {
		h1 = ",h1"
	}
	return fmt.Sprintf("%s|%s%s|%s", k.proxy, k.scheme, h1, k.addr)
}

// persistConn wraps a connection, usually a persistent one
// (but may be used for non-keep-alive requests as well)
type persistConn struct {
	// alt optionally specifies the TLS NextProto http.RoundTripper.
	// This is used for HTTP/2 today and future protocols later.
	// If it's non-nil, the rest of the fields are unused.
	alt http.RoundTripper

	t         *Transport
	cacheKey  connectMethodKey
	conn      net.Conn
	tlsState  *tls.ConnectionState
	br        *bufio.Reader       // from conn
	bw        *bufio.Writer       // to conn
	nwrite    int64               // bytes written
	reqch     chan requestAndChan // written by roundTrip; read by readLoop
	writech   chan writeRequest   // written by roundTrip; read by writeLoop
	closech   chan struct{}       // closed when conn closed
	isProxy   bool
	sawEOF    bool  // whether we've seen EOF from conn; owned by readLoop
	readLimit int64 // bytes allowed to be read; owned by readLoop
	// writeErrCh passes the request write error (usually nil)
	// from the writeLoop goroutine to the readLoop which passes
	// it off to the res.Body reader, which then uses it to decide
	// whether or not a connection can be reused. Issue 7569.
	writeErrCh chan error

	writeLoopDone chan struct{} // closed when write loop ends

	// Both guarded by Transport.idleMu:
	idleAt    time.Time   // time it last become idle
	idleTimer *time.Timer // holding an AfterFunc to close it

	mu                   sync.Mutex // guards following fields
	numExpectedResponses int
	closed               error // set non-nil when conn is closed, before closech is closed
	canceledErr          error // set non-nil if conn is canceled
	broken               bool  // an error has happened on this connection; marked broken so it's not reused.
	reused               bool  // whether conn has had successful request/response and is being reused.
	// mutateHeaderFunc is an optional func to modify extra
	// headers on each outbound request before it's written. (the
	// original Request given to RoundTrip is not modified)
	mutateHeaderFunc func(http.Header)
}

// RFC 7234, section 5.4: Should treat
//
//	Pragma: no-cache
//
// like
//
//	Cache-Control: no-cache
func fixPragmaCacheControl(header http.Header) {
	if hp, ok := header["Pragma"]; ok && len(hp) > 0 && hp[0] == "no-cache" {
		if _, presentcc := header["Cache-Control"]; !presentcc {
			header["Cache-Control"] = []string{"no-cache"}
		}
	}
}

// readResponse reads an HTTP response (or two, in the case of "Expect:
// 100-continue") from the server. It returns the final non-100 one.
// trace is optional.
func (pc *persistConn) _readResponse(req *http.Request) (*http.Response, error) {
	ds := dump.GetResponseHeaderDumpers(req.Context(), pc.t.Dump)
	tp := newTextprotoReader(pc.br, ds)
	resp := &http.Response{
		Request: req,
	}

	// Parse the first line of the response.
	line, err := tp.ReadLine()
	if err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return nil, err
	}
	proto, status, ok := util.CutString(line, " ")
	if !ok {
		return nil, badStringError("malformed HTTP response", line)
	}
	resp.Proto = proto
	resp.Status = strings.TrimLeft(status, " ")

	statusCode, _, _ := util.CutString(resp.Status, " ")
	if len(statusCode) != 3 {
		return nil, badStringError("malformed HTTP status code", statusCode)
	}
	resp.StatusCode, err = strconv.Atoi(statusCode)
	if err != nil || resp.StatusCode < 0 {
		return nil, badStringError("malformed HTTP status code", statusCode)
	}
	if resp.ProtoMajor, resp.ProtoMinor, ok = http.ParseHTTPVersion(resp.Proto); !ok {
		return nil, badStringError("malformed HTTP version", resp.Proto)
	}

	// Parse the response headers.
	mimeHeader, err := tp.ReadMIMEHeader()
	if err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return nil, err
	}
	resp.Header = http.Header(mimeHeader)

	fixPragmaCacheControl(resp.Header)

	err = readTransfer(resp, pc.br)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (pc *persistConn) maxHeaderResponseSize() int64 {
	if v := pc.t.MaxResponseHeaderBytes; v != 0 {
		return v
	}
	return 10 << 20 // conservative default; same as http2
}

func (pc *persistConn) Read(p []byte) (n int, err error) {
	if pc.readLimit <= 0 {
		return 0, fmt.Errorf("read limit of %d bytes exhausted", pc.maxHeaderResponseSize())
	}
	if int64(len(p)) > pc.readLimit {
		p = p[:pc.readLimit]
	}
	n, err = pc.conn.Read(p)
	if err == io.EOF {
		pc.sawEOF = true
	}
	pc.readLimit -= int64(n)
	return
}

// isBroken reports whether this connection is in a known broken state.
func (pc *persistConn) isBroken() bool {
	pc.mu.Lock()
	b := pc.closed != nil
	pc.mu.Unlock()
	return b
}

// canceled returns non-nil if the connection was closed due to
// CancelRequest or due to context cancellation.
func (pc *persistConn) canceled() error {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	return pc.canceledErr
}

// isReused reports whether this connection has been used before.
func (pc *persistConn) isReused() bool {
	pc.mu.Lock()
	r := pc.reused
	pc.mu.Unlock()
	return r
}

func (pc *persistConn) gotIdleConnTrace(idleAt time.Time) (t httptrace.GotConnInfo) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	t.Reused = pc.reused
	t.Conn = pc.conn
	t.WasIdle = true
	if !idleAt.IsZero() {
		t.IdleTime = time.Since(idleAt)
	}
	return
}

func (pc *persistConn) cancelRequest(err error) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.canceledErr = err
	pc.closeLocked(common.ErrRequestCanceled)
}

// closeConnIfStillIdle closes the connection if it's still sitting idle.
// This is what's called by the persistConn's idleTimer, and is run in its
// own goroutine.
func (pc *persistConn) closeConnIfStillIdle() {
	t := pc.t
	t.idleMu.Lock()
	defer t.idleMu.Unlock()
	if _, ok := t.idleLRU.m[pc]; !ok {
		// Not idle.
		return
	}
	t.removeIdleConnLocked(pc)
	pc.close(errIdleConnTimeout)
}

// mapRoundTripError returns the appropriate error value for
// persistConn.roundTrip.
//
// The provided err is the first error that (*persistConn).roundTrip
// happened to receive from its select statement.
//
// The startBytesWritten value should be the value of pc.nwrite before the roundTrip
// started writing the request.
func (pc *persistConn) mapRoundTripError(req *transportRequest, startBytesWritten int64, err error) error {
	if err == nil {
		return nil
	}

	// Wait for the writeLoop goroutine to terminate to avoid data
	// races on callers who mutate the request on failure.
	//
	// When resc in pc.roundTrip and hence rc.ch receives a responseAndError
	// with a non-nil error it implies that the persistConn is either closed
	// or closing. Waiting on pc.writeLoopDone is hence safe as all callers
	// close closech which in turn ensures writeLoop returns.
	<-pc.writeLoopDone

	// If the request was canceled, that's better than network
	// failures that were likely the result of tearing down the
	// connection.
	if cerr := pc.canceled(); cerr != nil {
		return cerr
	}

	// See if an error was set explicitly.
	req.mu.Lock()
	reqErr := req.err
	req.mu.Unlock()
	if reqErr != nil {
		return reqErr
	}

	if err == errServerClosedIdle {
		// Don't decorate
		return err
	}

	if _, ok := err.(transportReadFromServerError); ok {
		if pc.nwrite == startBytesWritten {
			return nothingWrittenError{err}
		}
		// Don't decorate
		return err
	}
	if pc.isBroken() {
		if pc.nwrite == startBytesWritten {
			return nothingWrittenError{err}
		}
		return fmt.Errorf("net/http: HTTP/1.x transport connection broken: %w", err)
	}
	return err
}

// errCallerOwnsConn is an internal sentinel error used when we hand
// off a writable response.Body to the caller. We use this to prevent
// closing a net.Conn that is now owned by the caller.
var errCallerOwnsConn = errors.New("read loop ending; caller owns writable underlying conn")

func (pc *persistConn) readLoop() {
	closeErr := errReadLoopExiting // default value, if not changed below
	defer func() {
		pc.close(closeErr)
		pc.t.removeIdleConn(pc)
	}()

	tryPutIdleConn := func(trace *httptrace.ClientTrace) bool {
		if err := pc.t.tryPutIdleConn(pc); err != nil {
			closeErr = err
			if trace != nil && trace.PutIdleConn != nil && err != errKeepAlivesDisabled {
				trace.PutIdleConn(err)
			}
			return false
		}
		if trace != nil && trace.PutIdleConn != nil {
			trace.PutIdleConn(nil)
		}
		return true
	}

	// eofc is used to block caller goroutines reading from Response.Body
	// at EOF until this goroutines has (potentially) added the connection
	// back to the idle pool.
	eofc := make(chan struct{})
	defer close(eofc) // unblock reader on errors

	// Read this once, before loop starts. (to avoid races in tests)
	testHookMu.Lock()
	testHookReadLoopBeforeNextRead := testHookReadLoopBeforeNextRead
	testHookMu.Unlock()

	alive := true
	for alive {
		pc.readLimit = pc.maxHeaderResponseSize()
		_, err := pc.br.Peek(1)

		pc.mu.Lock()
		if pc.numExpectedResponses == 0 {
			pc.readLoopPeekFailLocked(err)
			pc.mu.Unlock()
			return
		}
		pc.mu.Unlock()

		rc := <-pc.reqch
		trace := httptrace.ContextClientTrace(rc.req.Context())

		var resp *http.Response
		if err == nil {
			resp, err = pc.readResponse(rc, trace)
		} else {
			err = transportReadFromServerError{err}
			closeErr = err
		}

		if err != nil {
			if pc.readLimit <= 0 {
				err = fmt.Errorf("net/http: server response headers exceeded %d bytes; aborted", pc.maxHeaderResponseSize())
			}

			select {
			case rc.ch <- responseAndError{err: err}:
			case <-rc.callerGone:
				return
			}
			return
		}
		pc.readLimit = maxInt64 // effectively no limit for response bodies

		pc.mu.Lock()
		pc.numExpectedResponses--
		pc.mu.Unlock()

		bodyWritable := bodyIsWritable(resp)
		hasBody := rc.req.Method != "HEAD" && resp.ContentLength != 0

		if resp.Close || rc.req.Close || resp.StatusCode <= 199 || bodyWritable {
			// Don't do keep-alive on error if either party requested a close
			// or we get an unexpected informational (1xx) response.
			// StatusCode 100 is already handled above.
			alive = false
		}

		if !hasBody || bodyWritable {
			replaced := pc.t.replaceReqCanceler(rc.cancelKey, nil)

			// Put the idle conn back into the pool before we send the response
			// so if they process it quickly and make another request, they'll
			// get this same conn. But we use the unbuffered channel 'rc'
			// to guarantee that persistConn.roundTrip got out of its select
			// potentially waiting for this persistConn to close.
			alive = alive &&
				!pc.sawEOF &&
				pc.wroteRequest() &&
				replaced && tryPutIdleConn(trace)

			if bodyWritable {
				closeErr = errCallerOwnsConn
			}

			select {
			case rc.ch <- responseAndError{res: resp}:
			case <-rc.callerGone:
				return
			}

			// Now that they've read from the unbuffered channel, they're safely
			// out of the select that also waits on this goroutine to die, so
			// we're allowed to exit now if needed (if alive is false)
			testHookReadLoopBeforeNextRead()
			continue
		}

		waitForBodyRead := make(chan bool, 2)
		body := &bodyEOFSignal{
			body: resp.Body,
			earlyCloseFn: func() error {
				waitForBodyRead <- false
				<-eofc // will be closed by deferred call at the end of the function
				return nil

			},
			fn: func(err error) error {
				isEOF := err == io.EOF
				waitForBodyRead <- isEOF
				if isEOF {
					<-eofc // see comment above eofc declaration
				} else if err != nil {
					if cerr := pc.canceled(); cerr != nil {
						return cerr
					}
				}
				return err
			},
		}

		resp.Body = body
		if rc.addedGzip && ascii.EqualFold(resp.Header.Get("Content-Encoding"), "gzip") {
			resp.Body = &gzipReader{body: body}
			resp.Header.Del("Content-Encoding")
			resp.Header.Del("Content-Length")
			resp.ContentLength = -1
			resp.Uncompressed = true
		}

		select {
		case rc.ch <- responseAndError{res: resp}:
		case <-rc.callerGone:
			return
		}

		// Before looping back to the top of this function and peeking on
		// the bufio.textprotoReader, wait for the caller goroutine to finish
		// reading the response body. (or for cancellation or death)
		select {
		case bodyEOF := <-waitForBodyRead:
			replaced := pc.t.replaceReqCanceler(rc.cancelKey, nil) // before pc might return to idle pool
			alive = alive &&
				bodyEOF &&
				!pc.sawEOF &&
				pc.wroteRequest() &&
				replaced && tryPutIdleConn(trace)
			if bodyEOF {
				eofc <- struct{}{}
			}
		case <-rc.req.Cancel:
			alive = false
			pc.t.CancelRequest(rc.req)
		case <-rc.req.Context().Done():
			alive = false
			pc.t.cancelRequest(rc.cancelKey, rc.req.Context().Err())
		case <-pc.closech:
			alive = false
		}

		testHookReadLoopBeforeNextRead()
	}
}

func (pc *persistConn) readLoopPeekFailLocked(peekErr error) {
	if pc.closed != nil {
		return
	}
	if n := pc.br.Buffered(); n > 0 {
		buf, _ := pc.br.Peek(n)
		if is408Message(buf) {
			pc.closeLocked(errServerClosedIdle)
			return
		}
		log.Printf("Unsolicited response received on idle HTTP channel starting with %q; err=%v", buf, peekErr)
	}
	if peekErr == io.EOF {
		// common case.
		pc.closeLocked(errServerClosedIdle)
	} else {
		pc.closeLocked(fmt.Errorf("readLoopPeekFailLocked: %w", peekErr))
	}
}

// is408Message reports whether buf has the prefix of an
// HTTP 408 Request Timeout response.
// See golang.org/issue/32310.
func is408Message(buf []byte) bool {
	if len(buf) < len("HTTP/1.x 408") {
		return false
	}
	if string(buf[:7]) != "HTTP/1." {
		return false
	}
	return string(buf[8:12]) == " 408"
}

// readResponse reads an HTTP response (or two, in the case of "Expect:
// 100-continue") from the server. It returns the final non-100 one.
// trace is optional.
func (pc *persistConn) readResponse(rc requestAndChan, trace *httptrace.ClientTrace) (resp *http.Response, err error) {
	if trace != nil && trace.GotFirstResponseByte != nil {
		if peek, err := pc.br.Peek(1); err == nil && len(peek) == 1 {
			trace.GotFirstResponseByte()
		}
	}
	num1xx := 0               // number of informational 1xx headers received
	const max1xxResponses = 5 // arbitrary bound on number of informational responses

	continueCh := rc.continueCh
	for {
		resp, err = pc._readResponse(rc.req)
		if err != nil {
			return
		}
		resCode := resp.StatusCode
		if continueCh != nil {
			if resCode == 100 {
				if trace != nil && trace.Got100Continue != nil {
					trace.Got100Continue()
				}
				continueCh <- struct{}{}
				continueCh = nil
			} else if resCode >= 200 {
				close(continueCh)
				continueCh = nil
			}
		}
		is1xx := 100 <= resCode && resCode <= 199
		// treat 101 as a terminal status, see issue 26161
		is1xxNonTerminal := is1xx && resCode != http.StatusSwitchingProtocols

		if is1xxNonTerminal {
			num1xx++
			if num1xx > max1xxResponses {
				return nil, errors.New("net/http: too many 1xx informational responses")
			}
			pc.readLimit = pc.maxHeaderResponseSize() // reset the limit
			if trace != nil && trace.Got1xxResponse != nil {
				if err := trace.Got1xxResponse(resCode, textproto.MIMEHeader(resp.Header)); err != nil {
					return nil, err
				}
			}
			continue
		}
		break
	}
	if isProtocolSwitch(resp) {
		resp.Body = newReadWriteCloserBody(pc.br, pc.conn)
	}

	resp.TLS = pc.tlsState
	return
}

// waitForContinue returns the function to block until
// any response, timeout or connection close. After any of them,
// the function returns a bool which indicates if the body should be sent.
func (pc *persistConn) waitForContinue(continueCh <-chan struct{}) func() bool {
	if continueCh == nil {
		return nil
	}
	return func() bool {
		timer := time.NewTimer(pc.t.ExpectContinueTimeout)
		defer timer.Stop()

		select {
		case _, ok := <-continueCh:
			return ok
		case <-timer.C:
			return true
		case <-pc.closech:
			return false
		}
	}
}

func newReadWriteCloserBody(br *bufio.Reader, rwc io.ReadWriteCloser) io.ReadWriteCloser {
	body := &readWriteCloserBody{ReadWriteCloser: rwc}
	if br.Buffered() != 0 {
		body.br = br
	}
	return body
}

// readWriteCloserBody is the Response.Body type used when we want to
// give users write access to the Body through the underlying
// connection (TCP, unless using custom dialers). This is then
// the concrete type for a Response.Body on the 101 Switching
// Protocols response, as used by WebSockets, h2c, etc.
type readWriteCloserBody struct {
	_  incomparable
	br *bufio.Reader // used until empty
	io.ReadWriteCloser
}

func (b *readWriteCloserBody) Read(p []byte) (n int, err error) {
	if b.br != nil {
		if n := b.br.Buffered(); len(p) > n {
			p = p[:n]
		}
		n, err = b.br.Read(p)
		if b.br.Buffered() == 0 {
			b.br = nil
		}
		return n, err
	}
	return b.ReadWriteCloser.Read(p)
}

// nothingWrittenError wraps a write errors which ended up writing zero bytes.
type nothingWrittenError struct {
	error
}

func (nwe nothingWrittenError) Unwrap() error {
	return nwe.error
}

func (pc *persistConn) writeLoop() {
	defer close(pc.writeLoopDone)
	for {
		select {
		case wr := <-pc.writech:
			startBytesWritten := pc.nwrite
			err := pc.writeRequest(wr.req.Request, pc.bw, pc.isProxy, wr.req.extra, pc.waitForContinue(wr.continueCh))
			if bre, ok := err.(requestBodyReadError); ok {
				err = bre.error
				// Errors reading from the user's
				// Request.Body are high priority.
				// Set it here before sending on the
				// channels below or calling
				// pc.close() which tears down
				// connections and causes other
				// errors.
				wr.req.setError(err)
			}
			if err == nil {
				err = pc.bw.Flush()
			}
			if err != nil {
				if pc.nwrite == startBytesWritten {
					err = nothingWrittenError{err}
				}
			}
			pc.writeErrCh <- err // to the body reader, which might recycle us
			wr.ch <- err         // to the roundTrip function
			if err != nil {
				pc.close(err)
				return
			}
		case <-pc.closech:
			return
		}
	}
}

// extraHeaders may be nil
// waitForContinue may be nil
// always closes body
func (pc *persistConn) writeRequest(r *http.Request, w io.Writer, usingProxy bool, extraHeaders http.Header, waitForContinue func() bool) (err error) {
	trace := httptrace.ContextClientTrace(r.Context())
	if trace != nil && trace.WroteRequest != nil {
		defer func() {
			trace.WroteRequest(httptrace.WroteRequestInfo{
				Err: err,
			})
		}()
	}
	closed := false
	defer func() {
		if closed {
			return
		}
		if closeErr := closeRequestBody(r); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	// Find the target host. Prefer the Host: header, but if that
	// is not given, use the host from the request URL.
	//
	// Clean the host, in case it arrives with unexpected stuff in it.
	host := cleanHost(r.Host)
	if host == "" {
		if r.URL == nil {
			return errMissingHost
		}
		host = cleanHost(r.URL.Host)
	}

	// According to RFC 6874, an HTTP client, proxy, or other
	// intermediary must remove any IPv6 zone identifier attached
	// to an outgoing URI.
	host = removeZone(host)

	ruri := r.URL.RequestURI()
	if usingProxy && r.URL.Scheme != "" && r.URL.Opaque == "" {
		ruri = r.URL.Scheme + "://" + host + ruri
	} else if r.Method == "CONNECT" && r.URL.Path == "" {
		// CONNECT requests normally give just the host and port, not a full URL.
		ruri = host
		if r.URL.Opaque != "" {
			ruri = r.URL.Opaque
		}
	}
	if stringContainsCTLByte(ruri) {
		return errors.New("net/http: can't write control character in Request.URL")
	}
	// TODO: validate r.Method too? At least it's less likely to
	// come from an attacker (more likely to be a constant in
	// code).

	// Wrap the writer in a bufio Writer if it's not already buffered.
	// Don't always call NewWriter, as that forces a bytes.Buffer
	// and other small bufio Writers to have a minimum 4k buffer
	// size.
	var bw *bufio.Writer
	if _, ok := w.(io.ByteWriter); !ok {
		bw = bufio.NewWriter(w)
		w = bw
	}

	rw := w // raw writer
	dumps := dump.GetDumpers(r.Context(), pc.t.Dump)
	for _, dump := range dumps {
		if dump.RequestHeader() {
			w = dump.WrapRequestHeaderWriter(w)
		}
	}

	_, err = fmt.Fprintf(w, "%s %s HTTP/1.1\r\n", valueOrDefault(r.Method, "GET"), ruri)
	if err != nil {
		return err
	}

	// Header lines
	_, err = fmt.Fprintf(w, "Host: %s\r\n", host)
	if err != nil {
		return err
	}
	if trace != nil && trace.WroteHeaderField != nil {
		trace.WroteHeaderField("Host", []string{host})
	}

	// Use the defaultUserAgent unless the Header contains one, which
	// may be blank to not send the header.
	userAgent := header.DefaultUserAgent
	if headerHas(r.Header, "User-Agent") {
		userAgent = r.Header.Get("User-Agent")
	}
	if userAgent != "" {
		_, err = fmt.Fprintf(w, "User-Agent: %s\r\n", userAgent)
		if err != nil {
			return err
		}
		if trace != nil && trace.WroteHeaderField != nil {
			trace.WroteHeaderField("User-Agent", []string{userAgent})
		}
	}

	// Process Body,ContentLength,Close,Trailer
	tw, err := newTransferWriter(r)
	if err != nil {
		return err
	}
	err = tw.writeHeader(w, trace)
	if err != nil {
		return err
	}

	err = headerWriteSubset(r.Header, w, reqWriteExcludeHeader, trace)
	if err != nil {
		return err
	}

	if extraHeaders != nil {
		err = headerWrite(extraHeaders, w, trace)
		if err != nil {
			return err
		}
	}

	_, err = io.WriteString(w, "\r\n")
	if err != nil {
		return err
	}

	if trace != nil && trace.WroteHeaders != nil {
		trace.WroteHeaders()
	}

	// Flush and wait for 100-continue if expected.
	if waitForContinue != nil {
		if bw, ok := w.(*bufio.Writer); ok {
			err = bw.Flush()
			if err != nil {
				return err
			}
		}
		if trace != nil && trace.Wait100Continue != nil {
			trace.Wait100Continue()
		}
		if !waitForContinue() {
			closed = true
			closeRequestBody(r)
			return nil
		}
	}

	if bw, ok := w.(*bufio.Writer); ok && tw.FlushHeaders {
		if err := bw.Flush(); err != nil {
			return err
		}
	}

	// Write body and trailer
	closed = true
	err = tw.writeBody(rw, dumps)
	if err != nil {
		if tw.bodyReadError == err {
			err = requestBodyReadError{err}
		}
		return err
	}

	if bw != nil {
		return bw.Flush()
	}
	return nil
}

// maxWriteWaitBeforeConnReuse is how long the a Transport RoundTrip
// will wait to see the Request's Body.Write result after getting a
// response from the server. See comments in (*persistConn).wroteRequest.
//
// In tests, we set this to a large value to avoid flakiness from inconsistent
// recycling of connections.
var maxWriteWaitBeforeConnReuse = 50 * time.Millisecond

// wroteRequest is a check before recycling a connection that the previous write
// (from writeLoop above) happened and was successful.
func (pc *persistConn) wroteRequest() bool {
	select {
	case err := <-pc.writeErrCh:
		// Common case: the write happened well before the response, so
		// avoid creating a timer.
		return err == nil
	default:
		// Rare case: the request was written in writeLoop above but
		// before it could send to pc.writeErrCh, the reader read it
		// all, processed it, and called us here. In this case, give the
		// write goroutine a bit of time to finish its send.
		//
		// Less rare case: We also get here in the legitimate case of
		// Issue 7569, where the writer is still writing (or stalled),
		// but the server has already replied. In this case, we don't
		// want to wait too long, and we want to return false so this
		// connection isn't re-used.
		t := time.NewTimer(maxWriteWaitBeforeConnReuse)
		defer t.Stop()
		select {
		case err := <-pc.writeErrCh:
			return err == nil
		case <-t.C:
			return false
		}
	}
}

// responseAndError is how the goroutine reading from an HTTP/1 server
// communicates with the goroutine doing the RoundTrip.
type responseAndError struct {
	_   incomparable
	res *http.Response // else use this response (see res method)
	err error
}

type requestAndChan struct {
	_         incomparable
	req       *http.Request
	cancelKey cancelKey
	ch        chan responseAndError // unbuffered; always send in select on callerGone

	// whether the Transport (as opposed to the user client code)
	// added the Accept-Encoding gzip header. If the Transport
	// set it, only then do we transparently decode the gzip.
	addedGzip bool

	// Optional blocking chan for Expect: 100-continue (for send).
	// If the request has an "Expect: 100-continue" header and
	// the server responds 100 Continue, readLoop send a value
	// to writeLoop via this chan.
	continueCh chan<- struct{}

	callerGone <-chan struct{} // closed when roundTrip caller has returned
}

// A writeRequest is sent by the caller's goroutine to the
// writeLoop's goroutine to write a request while the read loop
// concurrently waits on both the write response and the server's
// reply.
type writeRequest struct {
	req *transportRequest
	ch  chan<- error

	// Optional blocking chan for Expect: 100-continue (for receive).
	// If not nil, writeLoop blocks sending request body until
	// it receives from this chan.
	continueCh <-chan struct{}
}

type httpError struct {
	err     string
	timeout bool
}

func (e *httpError) Error() string   { return e.err }
func (e *httpError) Timeout() bool   { return e.timeout }
func (e *httpError) Temporary() bool { return true }

var errTimeout error = &httpError{err: "net/http: timeout awaiting response headers", timeout: true}

var errRequestCanceledConn = errors.New("net/http: request canceled while waiting for connection") // TODO: unify?

func nop() {}

// testHooks. Always non-nil.
var (
	testHookEnterRoundTrip   = nop
	testHookWaitResLoop      = nop
	testHookRoundTripRetried = nop
	testHookPrePendingDial   = nop
	testHookPostPendingDial  = nop

	testHookMu                     sync.Locker = fakeLocker{} // guards following
	testHookReadLoopBeforeNextRead             = nop
)

func (pc *persistConn) roundTrip(req *transportRequest) (resp *http.Response, err error) {
	if pc.t.Debugf != nil {
		pc.t.Debugf("HTTP/1.1 %s %s", req.Method, req.URL.String())
	}
	testHookEnterRoundTrip()
	if !pc.t.replaceReqCanceler(req.cancelKey, pc.cancelRequest) {
		pc.t.putOrCloseIdleConn(pc)
		return nil, common.ErrRequestCanceled
	}
	pc.mu.Lock()
	pc.numExpectedResponses++
	headerFn := pc.mutateHeaderFunc
	pc.mu.Unlock()

	if headerFn != nil {
		headerFn(req.extraHeaders())
	}

	// Ask for a compressed version if the caller didn't set their
	// own value for Accept-Encoding. We only attempt to
	// uncompress the gzip stream if we were the layer that
	// requested it.
	requestedGzip := false
	if !pc.t.DisableCompression &&
		req.Header.Get("Accept-Encoding") == "" &&
		req.Header.Get("Range") == "" &&
		req.Method != "HEAD" {
		// Request gzip only, not deflate. Deflate is ambiguous and
		// not as universally supported anyway.
		// See: https://zlib.net/zlib_faq.html#faq39
		//
		// Note that we don't request this for HEAD requests,
		// due to a bug in nginx:
		//   https://trac.nginx.org/nginx/ticket/358
		//   https://golang.org/issue/5522
		//
		// We don't request gzip if the request is for a range, since
		// auto-decoding a portion of a gzipped document will just fail
		// anyway. See https://golang.org/issue/8923
		requestedGzip = true
		req.extraHeaders().Set("Accept-Encoding", "gzip")
	}

	var continueCh chan struct{}
	if req.ProtoAtLeast(1, 1) && req.Body != nil && reqExpectsContinue(req.Request) {
		continueCh = make(chan struct{}, 1)
	}

	if pc.t.DisableKeepAlives &&
		!reqWantsClose(req.Request) &&
		!isProtocolSwitchHeader(req.Header) {
		req.extraHeaders().Set("Connection", "close")
	}

	gone := make(chan struct{})
	defer close(gone)

	defer func() {
		if err != nil {
			pc.t.setReqCanceler(req.cancelKey, nil)
		}
	}()

	const debugRoundTrip = false

	// Write the request concurrently with waiting for a response,
	// in case the server decides to reply before reading our full
	// request body.
	startBytesWritten := pc.nwrite
	writeErrCh := make(chan error, 1)
	pc.writech <- writeRequest{req, writeErrCh, continueCh}

	resc := make(chan responseAndError)
	pc.reqch <- requestAndChan{
		req:        req.Request,
		cancelKey:  req.cancelKey,
		ch:         resc,
		addedGzip:  requestedGzip,
		continueCh: continueCh,
		callerGone: gone,
	}

	var respHeaderTimer <-chan time.Time
	cancelChan := req.Request.Cancel
	ctxDoneChan := req.Context().Done()
	pcClosed := pc.closech
	canceled := false
	for {
		testHookWaitResLoop()
		select {
		case err := <-writeErrCh:
			if debugRoundTrip {
				req.logf("writeErrCh resv: %T/%#v", err, err)
			}
			if err != nil {
				pc.close(fmt.Errorf("write error: %w", err))
				return nil, pc.mapRoundTripError(req, startBytesWritten, err)
			}
			if d := pc.t.ResponseHeaderTimeout; d > 0 {
				if debugRoundTrip {
					req.logf("starting timer for %v", d)
				}
				timer := time.NewTimer(d)
				defer timer.Stop() // prevent leaks
				respHeaderTimer = timer.C
			}
		case <-pcClosed:
			pcClosed = nil
			if canceled || pc.t.replaceReqCanceler(req.cancelKey, nil) {
				if debugRoundTrip {
					req.logf("closech recv: %T %#v", pc.closed, pc.closed)
				}
				return nil, pc.mapRoundTripError(req, startBytesWritten, pc.closed)
			}
		case <-respHeaderTimer:
			if debugRoundTrip {
				req.logf("timeout waiting for response headers.")
			}
			pc.close(errTimeout)
			return nil, errTimeout
		case re := <-resc:
			if (re.res == nil) == (re.err == nil) {
				panic(fmt.Sprintf("internal error: exactly one of res or err should be set; nil=%v", re.res == nil))
			}
			if debugRoundTrip {
				req.logf("resc recv: %p, %T/%#v", re.res, re.err, re.err)
			}
			if re.err != nil {
				return nil, pc.mapRoundTripError(req, startBytesWritten, re.err)
			}
			return re.res, nil
		case <-cancelChan:
			canceled = pc.t.cancelRequest(req.cancelKey, common.ErrRequestCanceled)
			cancelChan = nil
		case <-ctxDoneChan:
			canceled = pc.t.cancelRequest(req.cancelKey, req.Context().Err())
			cancelChan = nil
			ctxDoneChan = nil
		}
	}
}

// tLogKey is a context WithValue key for test debugging contexts containing
// a t.Logf func. See export_test.go's Request.WithT method.
type tLogKey struct{}

func (tr *transportRequest) logf(format string, args ...interface{}) {
	if logf, ok := tr.Request.Context().Value(tLogKey{}).(func(string, ...interface{})); ok {
		logf(time.Now().Format(time.RFC3339Nano)+": "+format, args...)
	}
}

// markReused marks this connection as having been successfully used for a
// request and response.
func (pc *persistConn) markReused() {
	pc.mu.Lock()
	pc.reused = true
	pc.mu.Unlock()
}

// close closes the underlying TCP connection and closes
// the pc.closech channel.
//
// The provided err is only for testing and debugging; in normal
// circumstances it should never be seen by users.
func (pc *persistConn) close(err error) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.closeLocked(err)
}

func (pc *persistConn) closeLocked(err error) {
	if err == nil {
		panic("nil error")
	}
	pc.broken = true
	if pc.closed == nil {
		pc.closed = err
		pc.t.decConnsPerHost(pc.cacheKey)
		// Close HTTP/1 (pc.alt == nil) connection.
		// HTTP/2 closes its connection itself.
		if pc.alt == nil {
			if err != errCallerOwnsConn {
				pc.conn.Close()
			}
			close(pc.closech)
		}
	}
	pc.mutateHeaderFunc = nil
}

var portMap = map[string]string{
	"http":   "80",
	"https":  "443",
	"socks5": "1080",
}

func idnaASCIIFromURL(url *url.URL) string {
	addr := url.Hostname()
	if v, err := idnaASCII(addr); err == nil {
		addr = v
	}
	return addr
}

// canonicalAddr returns url.Host but always with a ":port" suffix.
func canonicalAddr(url *url.URL) string {
	port := url.Port()
	if port == "" {
		port = portMap[url.Scheme]
	}
	return net.JoinHostPort(idnaASCIIFromURL(url), port)
}

// bodyEOFSignal is used by the HTTP/1 transport when reading response
// bodies to make sure we see the end of a response body before
// proceeding and reading on the connection again.
//
// It wraps a ReadCloser but runs fn (if non-nil) at most
// once, right before its final (error-producing) Read or Close call
// returns. fn should return the new error to return from Read or Close.
//
// If earlyCloseFn is non-nil and Close is called before io.EOF is
// seen, earlyCloseFn is called instead of fn, and its return value is
// the return value from Close.
type bodyEOFSignal struct {
	body         io.ReadCloser
	mu           sync.Mutex        // guards following 4 fields
	closed       bool              // whether Close has been called
	rerr         error             // sticky Read error
	fn           func(error) error // err will be nil on Read io.EOF
	earlyCloseFn func() error      // optional alt Close func used if io.EOF not seen
}

var errReadOnClosedResBody = errors.New("http: read on closed response body")

func (es *bodyEOFSignal) Read(p []byte) (n int, err error) {
	es.mu.Lock()
	closed, rerr := es.closed, es.rerr
	es.mu.Unlock()
	if closed {
		return 0, errReadOnClosedResBody
	}
	if rerr != nil {
		return 0, rerr
	}

	n, err = es.body.Read(p)
	if err != nil {
		es.mu.Lock()
		defer es.mu.Unlock()
		if es.rerr == nil {
			es.rerr = err
		}
		err = es.condfn(err)
	}
	return
}

func (es *bodyEOFSignal) Close() error {
	es.mu.Lock()
	defer es.mu.Unlock()
	if es.closed {
		return nil
	}
	es.closed = true
	if es.earlyCloseFn != nil && es.rerr != io.EOF {
		return es.earlyCloseFn()
	}
	err := es.body.Close()
	return es.condfn(err)
}

// caller must hold es.mu.
func (es *bodyEOFSignal) condfn(err error) error {
	if es.fn == nil {
		return err
	}
	err = es.fn(err)
	es.fn = nil
	return err
}

// gzipReader wraps a response body so it can lazily
// call gzip.NewReader on the first call to Read
type gzipReader struct {
	_    incomparable
	body *bodyEOFSignal // underlying HTTP/1 response body framing
	zr   *gzip.Reader   // lazily-initialized gzip reader
	zerr error          // any error from gzip.NewReader; sticky
}

func (gz *gzipReader) Read(p []byte) (n int, err error) {
	if gz.zr == nil {
		if gz.zerr == nil {
			gz.zr, gz.zerr = gzip.NewReader(gz.body)
		}
		if gz.zerr != nil {
			return 0, gz.zerr
		}
	}

	gz.body.mu.Lock()
	if gz.body.closed {
		err = errReadOnClosedResBody
	}
	gz.body.mu.Unlock()

	if err != nil {
		return 0, err
	}
	return gz.zr.Read(p)
}

func (gz *gzipReader) Close() error {
	return gz.body.Close()
}

type tlsHandshakeTimeoutError struct{}

func (tlsHandshakeTimeoutError) Timeout() bool   { return true }
func (tlsHandshakeTimeoutError) Temporary() bool { return true }
func (tlsHandshakeTimeoutError) Error() string   { return "net/http: TLS handshake timeout" }

// fakeLocker is a sync.Locker which does nothing. It's used to guard
// test-only fields when not under test, to avoid runtime atomic
// overhead.
type fakeLocker struct{}

func (fakeLocker) Lock()   {}
func (fakeLocker) Unlock() {}

// cloneTLSConfig returns a shallow clone of cfg, or a new zero tls.Config if
// cfg is nil. This is safe to call even if cfg is in active use by a TLS
// client or server.
func cloneTLSConfig(cfg *tls.Config) *tls.Config {
	if cfg == nil {
		return &tls.Config{}
	}
	return cfg.Clone()
}

type connLRU struct {
	ll *list.List // list.Element.Value type is of *persistConn
	m  map[*persistConn]*list.Element
}

// add adds pc to the head of the linked list.
func (cl *connLRU) add(pc *persistConn) {
	if cl.ll == nil {
		cl.ll = list.New()
		cl.m = make(map[*persistConn]*list.Element)
	}
	ele := cl.ll.PushFront(pc)
	if _, ok := cl.m[pc]; ok {
		panic("persistConn was already in LRU")
	}
	cl.m[pc] = ele
}

func (cl *connLRU) removeOldest() *persistConn {
	ele := cl.ll.Back()
	pc := ele.Value.(*persistConn)
	cl.ll.Remove(ele)
	delete(cl.m, pc)
	return pc
}

// remove removes pc from cl.
func (cl *connLRU) remove(pc *persistConn) {
	if ele, ok := cl.m[pc]; ok {
		cl.ll.Remove(ele)
		delete(cl.m, pc)
	}
}

// len returns the number of items in the cache.
func (cl *connLRU) len() int {
	return len(cl.m)
}
