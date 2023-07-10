package req

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	urlpkg "net/url"
	"os"
	"reflect"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"

	"github.com/imroc/req/v3/internal/header"
	"github.com/imroc/req/v3/internal/util"
	utls "github.com/refraction-networking/utls"
)

// DefaultClient returns the global default Client.
func DefaultClient() *Client {
	return defaultClient
}

// SetDefaultClient override the global default Client.
func SetDefaultClient(c *Client) {
	if c != nil {
		defaultClient = c
	}
}

var defaultClient = C()

// Client is the req's http client.
type Client struct {
	BaseURL               string
	PathParams            map[string]string
	QueryParams           urlpkg.Values
	Headers               http.Header
	Cookies               []*http.Cookie
	FormData              urlpkg.Values
	DebugLog              bool
	AllowGetMethodPayload bool

	trace                   bool
	disableAutoReadResponse bool
	commonErrorType         reflect.Type
	retryOption             *retryOption
	jsonMarshal             func(v interface{}) ([]byte, error)
	jsonUnmarshal           func(data []byte, v interface{}) error
	xmlMarshal              func(v interface{}) ([]byte, error)
	xmlUnmarshal            func(data []byte, v interface{}) error
	outputDirectory         string
	scheme                  string
	log                     Logger
	t                       *Transport
	dumpOptions             *DumpOptions
	httpClient              *http.Client
	beforeRequest           []RequestMiddleware
	udBeforeRequest         []RequestMiddleware
	afterResponse           []ResponseMiddleware
	wrappedRoundTrip        RoundTripper
	roundTripWrappers       []RoundTripWrapper
	responseBodyTransformer func(rawBody []byte, req *Request, resp *Response) (transformedBody []byte, err error)
	resultStateCheckFunc    func(resp *Response) ResultState
}

// R create a new request.
func (c *Client) R() *Request {
	return &Request{
		client:      c,
		retryOption: c.retryOption.Clone(),
	}
}

// Get create a new GET request, accepts 0 or 1 url.
func (c *Client) Get(url ...string) *Request {
	r := c.R()
	if len(url) > 0 {
		r.RawURL = url[0]
	}
	r.Method = http.MethodGet
	return r
}

// Post create a new POST request.
func (c *Client) Post(url ...string) *Request {
	r := c.R()
	if len(url) > 0 {
		r.RawURL = url[0]
	}
	r.Method = http.MethodPost
	return r
}

// Patch create a new PATCH request.
func (c *Client) Patch(url ...string) *Request {
	r := c.R()
	if len(url) > 0 {
		r.RawURL = url[0]
	}
	r.Method = http.MethodPatch
	return r
}

// Delete create a new DELETE request.
func (c *Client) Delete(url ...string) *Request {
	r := c.R()
	if len(url) > 0 {
		r.RawURL = url[0]
	}
	r.Method = http.MethodDelete
	return r
}

// Put create a new PUT request.
func (c *Client) Put(url ...string) *Request {
	r := c.R()
	if len(url) > 0 {
		r.RawURL = url[0]
	}
	r.Method = http.MethodPut
	return r
}

// Head create a new HEAD request.
func (c *Client) Head(url ...string) *Request {
	r := c.R()
	if len(url) > 0 {
		r.RawURL = url[0]
	}
	r.Method = http.MethodHead
	return r
}

// Options create a new OPTIONS request.
func (c *Client) Options(url ...string) *Request {
	r := c.R()
	if len(url) > 0 {
		r.RawURL = url[0]
	}
	r.Method = http.MethodOptions
	return r
}

// GetTransport return the underlying transport.
func (c *Client) GetTransport() *Transport {
	return c.t
}

// SetResponseBodyTransformer set the response body transformer, which can modify the
// response body before unmarshalled if auto-read response body is not disabled.
func (c *Client) SetResponseBodyTransformer(fn func(rawBody []byte, req *Request, resp *Response) (transformedBody []byte, err error)) *Client {
	c.responseBodyTransformer = fn
	return c
}

// SetCommonError set the common result that response body will be unmarshalled to
// if no error occurs but Response.ResultState returns ErrorState, by default it
// is HTTP status `code >= 400`, you can also use SetCommonResultStateChecker
// to customize the result state check logic.
//
// Deprecated: Use SetCommonErrorResult instead.
func (c *Client) SetCommonError(err interface{}) *Client {
	return c.SetCommonErrorResult(err)
}

// SetCommonErrorResult set the common result that response body will be unmarshalled to
// if no error occurs but Response.ResultState returns ErrorState, by default it
// is HTTP status `code >= 400`, you can also use SetCommonResultStateChecker
// to customize the result state check logic.
func (c *Client) SetCommonErrorResult(err interface{}) *Client {
	if err != nil {
		c.commonErrorType = util.GetType(err)
	}
	return c
}

// ResultState represents the state of the result.
type ResultState int

const (
	// SuccessState indicates the response is in success state,
	// and result will be unmarshalled if Request.SetSuccessResult
	// is called.
	SuccessState ResultState = iota
	// ErrorState indicates the response is in error state,
	// and result will be unmarshalled if Request.SetErrorResult
	// or Client.SetCommonErrorResult is called.
	ErrorState
	// UnknownState indicates the response is in unknown state,
	// and handler will be invoked if Request.SetUnknownResultHandlerFunc
	// or Client.SetCommonUnknownResultHandlerFunc is called.
	UnknownState
)

// SetResultStateCheckFunc overrides the default result state checker with customized one,
// which returns SuccessState when HTTP status `code >= 200 and <= 299`, and returns
// ErrorState when HTTP status `code >= 400`, otherwise returns UnknownState.
func (c *Client) SetResultStateCheckFunc(fn func(resp *Response) ResultState) *Client {
	c.resultStateCheckFunc = fn
	return c
}

// SetCommonFormDataFromValues set the form data from url.Values for requests
// fired from the client which request method allows payload.
func (c *Client) SetCommonFormDataFromValues(data urlpkg.Values) *Client {
	if c.FormData == nil {
		c.FormData = urlpkg.Values{}
	}
	for k, v := range data {
		for _, kv := range v {
			c.FormData.Add(k, kv)
		}
	}
	return c
}

// SetCommonFormData set the form data from map for requests fired from the client
// which request method allows payload.
func (c *Client) SetCommonFormData(data map[string]string) *Client {
	if c.FormData == nil {
		c.FormData = urlpkg.Values{}
	}
	for k, v := range data {
		c.FormData.Set(k, v)
	}
	return c
}

// SetBaseURL set the default base URL, will be used if request URL is
// a relative URL.
func (c *Client) SetBaseURL(u string) *Client {
	c.BaseURL = strings.TrimRight(u, "/")
	return c
}

// SetOutputDirectory set output directory that response will
// be downloaded to.
func (c *Client) SetOutputDirectory(dir string) *Client {
	c.outputDirectory = dir
	return c
}

// SetCertFromFile helps to set client certificates from cert and key file.
func (c *Client) SetCertFromFile(certFile, keyFile string) *Client {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		c.log.Errorf("failed to load client cert: %v", err)
		return c
	}
	config := c.GetTLSClientConfig()
	config.Certificates = append(config.Certificates, cert)
	return c
}

// SetCerts set client certificates.
func (c *Client) SetCerts(certs ...tls.Certificate) *Client {
	config := c.GetTLSClientConfig()
	config.Certificates = append(config.Certificates, certs...)
	return c
}

func (c *Client) appendRootCertData(data []byte) {
	config := c.GetTLSClientConfig()
	if config.RootCAs == nil {
		config.RootCAs = x509.NewCertPool()
	}
	config.RootCAs.AppendCertsFromPEM(data)
	return
}

// SetRootCertFromString set root certificates from string.
func (c *Client) SetRootCertFromString(pemContent string) *Client {
	c.appendRootCertData([]byte(pemContent))
	return c
}

// SetRootCertsFromFile set root certificates from files.
func (c *Client) SetRootCertsFromFile(pemFiles ...string) *Client {
	for _, pemFile := range pemFiles {
		rootPemData, err := os.ReadFile(pemFile)
		if err != nil {
			c.log.Errorf("failed to read root cert file: %v", err)
			return c
		}
		c.appendRootCertData(rootPemData)
	}
	return c
}

// GetTLSClientConfig return the underlying tls.Config.
func (c *Client) GetTLSClientConfig() *tls.Config {
	if c.t.TLSClientConfig == nil {
		c.t.TLSClientConfig = &tls.Config{
			NextProtos: []string{"h2", "http/1.1"},
		}
	}
	return c.t.TLSClientConfig
}

func (c *Client) defaultCheckRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return errors.New("stopped after 10 redirects")
	}
	if c.DebugLog {
		c.log.Debugf("<redirect> %s %s", req.Method, req.URL.String())
	}
	return nil
}

// SetRedirectPolicy set the RedirectPolicy which controls the behavior of receiving redirect
// responses (usually responses with 301 and 302 status code), see the predefined
// AllowedDomainRedirectPolicy, AllowedHostRedirectPolicy, MaxRedirectPolicy, NoRedirectPolicy,
// SameDomainRedirectPolicy and SameHostRedirectPolicy.
func (c *Client) SetRedirectPolicy(policies ...RedirectPolicy) *Client {
	if len(policies) == 0 {
		return c
	}
	c.httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		for _, f := range policies {
			if f == nil {
				continue
			}
			err := f(req, via)
			if err != nil {
				return err
			}
		}
		if c.DebugLog {
			c.log.Debugf("<redirect> %s %s", req.Method, req.URL.String())
		}
		return nil
	}
	return c
}

// DisableKeepAlives disable the HTTP keep-alives (enabled by default)
// and will only use the connection to the server for a single
// HTTP request.
//
// This is unrelated to the similarly named TCP keep-alives.
func (c *Client) DisableKeepAlives() *Client {
	c.t.DisableKeepAlives = true
	return c
}

// EnableKeepAlives enables HTTP keep-alives (enabled by default).
func (c *Client) EnableKeepAlives() *Client {
	c.t.DisableKeepAlives = false
	return c
}

// DisableCompression disables the compression (enabled by default),
// which prevents the Transport from requesting compression
// with an "Accept-Encoding: gzip" request header when the
// Request contains no existing Accept-Encoding value. If
// the Transport requests gzip on its own and gets a gzipped
// response, it's transparently decoded in the Response.Body.
// However, if the user explicitly requested gzip it is not
// automatically uncompressed.
func (c *Client) DisableCompression() *Client {
	c.t.DisableCompression = true
	return c
}

// EnableCompression enables the compression (enabled by default).
func (c *Client) EnableCompression() *Client {
	c.t.DisableCompression = false
	return c
}

// SetTLSClientConfig set the TLS client config. Be careful! Usually
// you don't need this, you can directly set the tls configuration with
// methods like EnableInsecureSkipVerify, SetCerts etc. Or you can call
// GetTLSClientConfig to get the current tls configuration to avoid
// overwriting some important configurations, such as not setting NextProtos
// will not use http2 by default.
func (c *Client) SetTLSClientConfig(conf *tls.Config) *Client {
	c.t.TLSClientConfig = conf
	return c
}

// EnableInsecureSkipVerify enable send https without verifing
// the server's certificates (disabled by default).
func (c *Client) EnableInsecureSkipVerify() *Client {
	c.GetTLSClientConfig().InsecureSkipVerify = true
	return c
}

// DisableInsecureSkipVerify disable send https without verifing
// the server's certificates (disabled by default).
func (c *Client) DisableInsecureSkipVerify() *Client {
	c.GetTLSClientConfig().InsecureSkipVerify = false
	return c
}

// SetCommonQueryParams set URL query parameters with a map
// for requests fired from the client.
func (c *Client) SetCommonQueryParams(params map[string]string) *Client {
	for k, v := range params {
		c.SetCommonQueryParam(k, v)
	}
	return c
}

// AddCommonQueryParam add a URL query parameter with a key-value
// pair for requests fired from the client.
func (c *Client) AddCommonQueryParam(key, value string) *Client {
	if c.QueryParams == nil {
		c.QueryParams = make(urlpkg.Values)
	}
	c.QueryParams.Add(key, value)
	return c
}

// AddCommonQueryParams add one or more values of specified URL query parameter
// for requests fired from the client.
func (c *Client) AddCommonQueryParams(key string, values ...string) *Client {
	if c.QueryParams == nil {
		c.QueryParams = make(urlpkg.Values)
	}
	vs := c.QueryParams[key]
	vs = append(vs, values...)
	c.QueryParams[key] = vs
	return c
}

func (c *Client) pathParams() map[string]string {
	if c.PathParams == nil {
		c.PathParams = make(map[string]string)
	}
	return c.PathParams
}

// SetCommonPathParam set a path parameter for requests fired from the client.
func (c *Client) SetCommonPathParam(key, value string) *Client {
	c.pathParams()[key] = value
	return c
}

// SetCommonPathParams set path parameters for requests fired from the client.
func (c *Client) SetCommonPathParams(pathParams map[string]string) *Client {
	m := c.pathParams()
	for k, v := range pathParams {
		m[k] = v
	}
	return c
}

// SetCommonQueryParam set a URL query parameter with a key-value
// pair for requests fired from the client.
func (c *Client) SetCommonQueryParam(key, value string) *Client {
	if c.QueryParams == nil {
		c.QueryParams = make(urlpkg.Values)
	}
	c.QueryParams.Set(key, value)
	return c
}

// SetCommonQueryString set URL query parameters with a raw query string
// for requests fired from the client.
func (c *Client) SetCommonQueryString(query string) *Client {
	params, err := urlpkg.ParseQuery(strings.TrimSpace(query))
	if err != nil {
		c.log.Warnf("failed to parse query string (%s): %v", query, err)
		return c
	}
	if c.QueryParams == nil {
		c.QueryParams = make(urlpkg.Values)
	}
	for p, v := range params {
		for _, pv := range v {
			c.QueryParams.Add(p, pv)
		}
	}
	return c
}

// SetCommonCookies set HTTP cookies for requests fired from the client.
func (c *Client) SetCommonCookies(cookies ...*http.Cookie) *Client {
	c.Cookies = append(c.Cookies, cookies...)
	return c
}

// DisableDebugLog disable debug level log (disabled by default).
func (c *Client) DisableDebugLog() *Client {
	c.DebugLog = false
	return c
}

// EnableDebugLog enable debug level log (disabled by default).
func (c *Client) EnableDebugLog() *Client {
	c.DebugLog = true
	return c
}

// DevMode enables:
// 1. Dump content of all requests and responses to see details.
// 2. Output debug level log for deeper insights.
// 3. Trace all requests, so you can get trace info to analyze performance.
func (c *Client) DevMode() *Client {
	return c.EnableDumpAll().
		EnableDebugLog().
		EnableTraceAll()
}

// SetScheme set the default scheme for client, will be used when
// there is no scheme in the request URL (e.g. "github.com/imroc/req").
func (c *Client) SetScheme(scheme string) *Client {
	if !util.IsStringEmpty(scheme) {
		c.scheme = strings.TrimSpace(scheme)
	}
	return c
}

// GetLogger return the internal logger, usually used in middleware.
func (c *Client) GetLogger() Logger {
	if c.log != nil {
		return c.log
	}
	c.log = createDefaultLogger()
	return c.log
}

// SetLogger set the customized logger for client, will disable log if set to nil.
func (c *Client) SetLogger(log Logger) *Client {
	if log == nil {
		c.log = &disableLogger{}
		return c
	}
	c.log = log
	return c
}

// SetTimeout set timeout for requests fired from the client.
func (c *Client) SetTimeout(d time.Duration) *Client {
	c.httpClient.Timeout = d
	return c
}

func (c *Client) getDumpOptions() *DumpOptions {
	if c.dumpOptions == nil {
		c.dumpOptions = newDefaultDumpOptions()
	}
	return c.dumpOptions
}

// EnableDumpAll enable dump for requests fired from the client, including
// all content for the request and response by default.
func (c *Client) EnableDumpAll() *Client {
	if c.t.Dump != nil { // dump already started
		return c
	}
	c.t.EnableDump(c.getDumpOptions())
	return c
}

// EnableDumpAllToFile enable dump for requests fired from the
// client and output to the specified file.
func (c *Client) EnableDumpAllToFile(filename string) *Client {
	file, err := os.Create(filename)
	if err != nil {
		c.log.Errorf("create dump file error: %v", err)
		return c
	}
	c.getDumpOptions().Output = file
	c.EnableDumpAll()
	return c
}

// EnableDumpAllTo enable dump for requests fired from the
// client and output to the specified io.Writer.
func (c *Client) EnableDumpAllTo(output io.Writer) *Client {
	c.getDumpOptions().Output = output
	c.EnableDumpAll()
	return c
}

// EnableDumpAllAsync enable dump for requests fired from the
// client and output asynchronously, can be used for debugging
// in production environment without affecting performance.
func (c *Client) EnableDumpAllAsync() *Client {
	o := c.getDumpOptions()
	o.Async = true
	c.EnableDumpAll()
	return c
}

// EnableDumpAllWithoutRequestBody enable dump for requests fired
// from the client without request body, can be used in the upload
// request to avoid dumping the unreadable binary content.
func (c *Client) EnableDumpAllWithoutRequestBody() *Client {
	o := c.getDumpOptions()
	o.RequestBody = false
	c.EnableDumpAll()
	return c
}

// EnableDumpAllWithoutResponseBody enable dump for requests fired
// from the client without response body, can be used in the download
// request to avoid dumping the unreadable binary content.
func (c *Client) EnableDumpAllWithoutResponseBody() *Client {
	o := c.getDumpOptions()
	o.ResponseBody = false
	c.EnableDumpAll()
	return c
}

// EnableDumpAllWithoutResponse enable dump for requests fired from
// the client without response, can be used if you only care about
// the request.
func (c *Client) EnableDumpAllWithoutResponse() *Client {
	o := c.getDumpOptions()
	o.ResponseBody = false
	o.ResponseHeader = false
	c.EnableDumpAll()
	return c
}

// EnableDumpAllWithoutRequest enables dump for requests fired from
// the client without request, can be used if you only care about
// the response.
func (c *Client) EnableDumpAllWithoutRequest() *Client {
	o := c.getDumpOptions()
	o.RequestHeader = false
	o.RequestBody = false
	c.EnableDumpAll()
	return c
}

// EnableDumpAllWithoutHeader enable dump for requests fired from
// the client without header, can be used if you only care about
// the body.
func (c *Client) EnableDumpAllWithoutHeader() *Client {
	o := c.getDumpOptions()
	o.RequestHeader = false
	o.ResponseHeader = false
	c.EnableDumpAll()
	return c
}

// EnableDumpAllWithoutBody enable dump for requests fired from
// the client without body, can be used if you only care about
// the header.
func (c *Client) EnableDumpAllWithoutBody() *Client {
	o := c.getDumpOptions()
	o.RequestBody = false
	o.ResponseBody = false
	c.EnableDumpAll()
	return c
}

// EnableDumpEachRequest enable dump at the request-level for each request, and only
// temporarily stores the dump content in memory, call Response.Dump() to get the
// dump content when needed.
func (c *Client) EnableDumpEachRequest() *Client {
	return c.OnBeforeRequest(func(client *Client, req *Request) error {
		if req.RetryAttempt == 0 { // Ignore on retry, no need to repeat enable dump.
			req.EnableDump()
		}
		return nil
	})
}

// EnableDumpEachRequestWithoutBody enable dump without body at the request-level for
// each request, and only temporarily stores the dump content in memory, call
// Response.Dump() to get the dump content when needed.
func (c *Client) EnableDumpEachRequestWithoutBody() *Client {
	return c.OnBeforeRequest(func(client *Client, req *Request) error {
		if req.RetryAttempt == 0 { // Ignore on retry, no need to repeat enable dump.
			req.EnableDumpWithoutBody()
		}
		return nil
	})
}

// EnableDumpEachRequestWithoutHeader enable dump without header at the request-level for
// each request, and only temporarily stores the dump content in memory, call
// Response.Dump() to get the dump content when needed.
func (c *Client) EnableDumpEachRequestWithoutHeader() *Client {
	return c.OnBeforeRequest(func(client *Client, req *Request) error {
		if req.RetryAttempt == 0 { // Ignore on retry, no need to repeat enable dump.
			req.EnableDumpWithoutHeader()
		}
		return nil
	})
}

// EnableDumpEachRequestWithoutRequest enable dump without request at the request-level for
// each request, and only temporarily stores the dump content in memory, call
// Response.Dump() to get the dump content when needed.
func (c *Client) EnableDumpEachRequestWithoutRequest() *Client {
	return c.OnBeforeRequest(func(client *Client, req *Request) error {
		if req.RetryAttempt == 0 { // Ignore on retry, no need to repeat enable dump.
			req.EnableDumpWithoutRequest()
		}
		return nil
	})
}

// EnableDumpEachRequestWithoutResponse enable dump without response at the request-level for
// each request, and only temporarily stores the dump content in memory, call
// Response.Dump() to get the dump content when needed.
func (c *Client) EnableDumpEachRequestWithoutResponse() *Client {
	return c.OnBeforeRequest(func(client *Client, req *Request) error {
		if req.RetryAttempt == 0 { // Ignore on retry, no need to repeat enable dump.
			req.EnableDumpWithoutResponse()
		}
		return nil
	})
}

// EnableDumpEachRequestWithoutResponseBody enable dump without response body at the
// request-level for each request, and only temporarily stores the dump content in memory,
// call Response.Dump() to get the dump content when needed.
func (c *Client) EnableDumpEachRequestWithoutResponseBody() *Client {
	return c.OnBeforeRequest(func(client *Client, req *Request) error {
		if req.RetryAttempt == 0 { // Ignore on retry, no need to repeat enable dump.
			req.EnableDumpWithoutResponseBody()
		}
		return nil
	})
}

// EnableDumpEachRequestWithoutRequestBody enable dump without request body at the
// request-level for each request, and only temporarily stores the dump content in memory,
// call Response.Dump() to get the dump content when needed.
func (c *Client) EnableDumpEachRequestWithoutRequestBody() *Client {
	return c.OnBeforeRequest(func(client *Client, req *Request) error {
		if req.RetryAttempt == 0 { // Ignore on retry, no need to repeat enable dump.
			req.EnableDumpWithoutRequestBody()
		}
		return nil
	})
}

// NewRequest is the alias of R()
func (c *Client) NewRequest() *Request {
	return c.R()
}

func (c *Client) NewParallelDownload(url string) *ParallelDownload {
	return &ParallelDownload{
		url:    url,
		client: c,
	}
}

// DisableAutoReadResponse disable read response body automatically (enabled by default).
func (c *Client) DisableAutoReadResponse() *Client {
	c.disableAutoReadResponse = true
	return c
}

// EnableAutoReadResponse enable read response body automatically (enabled by default).
func (c *Client) EnableAutoReadResponse() *Client {
	c.disableAutoReadResponse = false
	return c
}

// SetAutoDecodeContentType set the content types that will be auto-detected and decode to utf-8
// (e.g. "json", "xml", "html", "text").
func (c *Client) SetAutoDecodeContentType(contentTypes ...string) *Client {
	c.t.SetAutoDecodeContentType(contentTypes...)
	return c
}

// SetAutoDecodeContentTypeFunc set the function that determines whether the specified `Content-Type` should be auto-detected and decode to utf-8.
func (c *Client) SetAutoDecodeContentTypeFunc(fn func(contentType string) bool) *Client {
	c.t.SetAutoDecodeContentTypeFunc(fn)
	return c
}

// SetAutoDecodeAllContentType enable try auto-detect charset and decode all content type to utf-8.
func (c *Client) SetAutoDecodeAllContentType() *Client {
	c.t.SetAutoDecodeAllContentType()
	return c
}

// DisableAutoDecode disable auto-detect charset and decode to utf-8 (enabled by default).
func (c *Client) DisableAutoDecode() *Client {
	c.t.DisableAutoDecode()
	return c
}

// EnableAutoDecode enable auto-detect charset and decode to utf-8 (enabled by default).
func (c *Client) EnableAutoDecode() *Client {
	c.t.EnableAutoDecode()
	return c
}

// SetUserAgent set the "User-Agent" header for requests fired from the client.
func (c *Client) SetUserAgent(userAgent string) *Client {
	return c.SetCommonHeader(header.UserAgent, userAgent)
}

// SetCommonBearerAuthToken set the bearer auth token for requests fired from the client.
func (c *Client) SetCommonBearerAuthToken(token string) *Client {
	return c.SetCommonHeader(header.Authorization, "Bearer "+token)
}

// SetCommonBasicAuth set the basic auth for requests fired from
// the client.
func (c *Client) SetCommonBasicAuth(username, password string) *Client {
	c.SetCommonHeader(header.Authorization, util.BasicAuthHeaderValue(username, password))
	return c
}

// SetCommonDigestAuth sets the Digest Access auth scheme for requests fired from the client. If a server responds with
// 401 and sends a Digest challenge in the WWW-Authenticate Header, requests will be resent with the appropriate
// Authorization Header.
//
// For Example: To set the Digest scheme with user "roc" and password "123456"
//
//	client.SetCommonDigestAuth("roc", "123456")
//
// Information about Digest Access Authentication can be found in RFC7616:
//
//	https://datatracker.ietf.org/doc/html/rfc7616
//
// See `Request.SetDigestAuth`
func (c *Client) SetCommonDigestAuth(username, password string) *Client {
	c.OnAfterResponse(handleDigestAuthFunc(username, password))
	return c
}

// SetCommonHeaders set headers for requests fired from the client.
func (c *Client) SetCommonHeaders(hdrs map[string]string) *Client {
	for k, v := range hdrs {
		c.SetCommonHeader(k, v)
	}
	return c
}

// SetCommonHeader set a header for requests fired from the client.
func (c *Client) SetCommonHeader(key, value string) *Client {
	if c.Headers == nil {
		c.Headers = make(http.Header)
	}
	c.Headers.Set(key, value)
	return c
}

// SetCommonHeaderNonCanonical set a header for requests fired from
// the client which key is a non-canonical key (keep case unchanged),
// only valid for HTTP/1.1.
func (c *Client) SetCommonHeaderNonCanonical(key, value string) *Client {
	if c.Headers == nil {
		c.Headers = make(http.Header)
	}
	c.Headers[key] = append(c.Headers[key], value)
	return c
}

// SetCommonHeadersNonCanonical set headers for requests fired from the
// client which key is a non-canonical key (keep case unchanged), only
// valid for HTTP/1.1.
func (c *Client) SetCommonHeadersNonCanonical(hdrs map[string]string) *Client {
	for k, v := range hdrs {
		c.SetCommonHeaderNonCanonical(k, v)
	}
	return c
}

// SetCommonContentType set the `Content-Type` header for requests fired
// from the client.
func (c *Client) SetCommonContentType(ct string) *Client {
	c.SetCommonHeader(header.ContentType, ct)
	return c
}

// DisableDumpAll disable dump for requests fired from the client.
func (c *Client) DisableDumpAll() *Client {
	c.t.DisableDump()
	return c
}

// SetCommonDumpOptions configures the underlying Transport's DumpOptions
// for requests fired from the client.
func (c *Client) SetCommonDumpOptions(opt *DumpOptions) *Client {
	if opt == nil {
		return c
	}
	if opt.Output == nil {
		if c.dumpOptions != nil {
			opt.Output = c.dumpOptions.Output
		} else {
			opt.Output = os.Stdout
		}
	}
	c.dumpOptions = opt
	if c.t.Dump != nil {
		c.t.Dump.SetOptions(dumpOptions{opt})
	}
	return c
}

// SetProxy set the proxy function.
func (c *Client) SetProxy(proxy func(*http.Request) (*urlpkg.URL, error)) *Client {
	c.t.SetProxy(proxy)
	return c
}

// OnBeforeRequest add a request middleware which hooks before request sent.
func (c *Client) OnBeforeRequest(m RequestMiddleware) *Client {
	c.udBeforeRequest = append(c.udBeforeRequest, m)
	return c
}

// OnAfterResponse add a response middleware which hooks after response received.
func (c *Client) OnAfterResponse(m ResponseMiddleware) *Client {
	c.afterResponse = append(c.afterResponse, m)
	return c
}

// SetProxyURL set proxy from the proxy URL.
func (c *Client) SetProxyURL(proxyUrl string) *Client {
	if proxyUrl == "" {
		c.log.Warnf("ignore empty proxy url in SetProxyURL")
		return c
	}
	u, err := urlpkg.Parse(proxyUrl)
	if err != nil {
		c.log.Errorf("failed to parse proxy url %s: %v", proxyUrl, err)
		return c
	}
	proxy := http.ProxyURL(u)
	c.t.SetProxy(proxy)
	return c
}

// DisableTraceAll disable trace for requests fired from the client.
func (c *Client) DisableTraceAll() *Client {
	c.trace = false
	return c
}

// EnableTraceAll enable trace for requests fired from the client (http3
// currently does not support trace).
func (c *Client) EnableTraceAll() *Client {
	c.trace = true
	return c
}

// SetCookieJar set the `CookeJar` to the underlying `http.Client`, set to nil if you
// want to disable cookie.
func (c *Client) SetCookieJar(jar http.CookieJar) *Client {
	c.httpClient.Jar = jar
	return c
}

// GetCookies get cookies from the underlying `http.Client`'s `CookieJar`.
func (c *Client) GetCookies(url string) ([]*http.Cookie, error) {
	if c.httpClient.Jar == nil {
		return nil, errors.New("cookie jar is not enabled")
	}
	u, err := urlpkg.Parse(url)
	if err != nil {
		return nil, err
	}
	return c.httpClient.Jar.Cookies(u), nil
}

// ClearCookies clears all cookies if cookie is enabled.
func (c *Client) ClearCookies() *Client {
	if c.httpClient.Jar != nil {
		jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		c.httpClient.Jar = jar
	}
	return c
}

// SetJsonMarshal set the JSON marshal function which will be used
// to marshal request body.
func (c *Client) SetJsonMarshal(fn func(v interface{}) ([]byte, error)) *Client {
	c.jsonMarshal = fn
	return c
}

// SetJsonUnmarshal set the JSON unmarshal function which will be used
// to unmarshal response body.
func (c *Client) SetJsonUnmarshal(fn func(data []byte, v interface{}) error) *Client {
	c.jsonUnmarshal = fn
	return c
}

// SetXmlMarshal set the XML marshal function which will be used
// to marshal request body.
func (c *Client) SetXmlMarshal(fn func(v interface{}) ([]byte, error)) *Client {
	c.xmlMarshal = fn
	return c
}

// SetXmlUnmarshal set the XML unmarshal function which will be used
// to unmarshal response body.
func (c *Client) SetXmlUnmarshal(fn func(data []byte, v interface{}) error) *Client {
	c.xmlUnmarshal = fn
	return c
}

// SetDialTLS set the customized `DialTLSContext` function to Transport.
// Make sure the returned `conn` implements pkg/tls.Conn if you want your
// customized `conn` supports HTTP2.
func (c *Client) SetDialTLS(fn func(ctx context.Context, network, addr string) (net.Conn, error)) *Client {
	c.t.SetDialTLS(fn)
	return c
}

// SetDial set the customized `DialContext` function to Transport.
func (c *Client) SetDial(fn func(ctx context.Context, network, addr string) (net.Conn, error)) *Client {
	c.t.SetDial(fn)
	return c
}

// SetTLSFingerprintChrome uses tls fingerprint of Chrome browser.
func (c *Client) SetTLSFingerprintChrome() *Client {
	return c.SetTLSFingerprint(utls.HelloChrome_Auto)
}

// SetTLSFingerprintFirefox uses tls fingerprint of Firefox browser.
func (c *Client) SetTLSFingerprintFirefox() *Client {
	return c.SetTLSFingerprint(utls.HelloFirefox_Auto)
}

// SetTLSFingerprintEdge uses tls fingerprint of Edge browser.
func (c *Client) SetTLSFingerprintEdge() *Client {
	return c.SetTLSFingerprint(utls.HelloEdge_Auto)
}

// SetTLSFingerprintQQ uses tls fingerprint of QQ browser.
func (c *Client) SetTLSFingerprintQQ() *Client {
	return c.SetTLSFingerprint(utls.HelloQQ_Auto)
}

// SetTLSFingerprintSafari uses tls fingerprint of Safari browser.
func (c *Client) SetTLSFingerprintSafari() *Client {
	return c.SetTLSFingerprint(utls.HelloSafari_Auto)
}

// SetTLSFingerprint360 uses tls fingerprint of 360 browser.
func (c *Client) SetTLSFingerprint360() *Client {
	return c.SetTLSFingerprint(utls.Hello360_Auto)
}

// SetTLSFingerprintIOS uses tls fingerprint of IOS.
func (c *Client) SetTLSFingerprintIOS() *Client {
	return c.SetTLSFingerprint(utls.HelloIOS_Auto)
}

// SetTLSFingerprintAndroid uses tls fingerprint of Android.
func (c *Client) SetTLSFingerprintAndroid() *Client {
	return c.SetTLSFingerprint(utls.HelloAndroid_11_OkHttp)
}

// SetTLSFingerprintRandomized uses randomized tls fingerprint.
func (c *Client) SetTLSFingerprintRandomized() *Client {
	return c.SetTLSFingerprint(utls.HelloRandomized)
}

// uTLSConn is wrapper of UConn which implements the net.Conn interface.
type uTLSConn struct {
	*utls.UConn
}

func (conn *uTLSConn) ConnectionState() tls.ConnectionState {
	cs := conn.Conn.ConnectionState()
	return tls.ConnectionState{
		Version:                     cs.Version,
		HandshakeComplete:           cs.HandshakeComplete,
		DidResume:                   cs.DidResume,
		CipherSuite:                 cs.CipherSuite,
		NegotiatedProtocol:          cs.NegotiatedProtocol,
		NegotiatedProtocolIsMutual:  cs.NegotiatedProtocolIsMutual,
		ServerName:                  cs.ServerName,
		PeerCertificates:            cs.PeerCertificates,
		VerifiedChains:              cs.VerifiedChains,
		SignedCertificateTimestamps: cs.SignedCertificateTimestamps,
		OCSPResponse:                cs.OCSPResponse,
		TLSUnique:                   cs.TLSUnique,
	}
}

// SetTLSFingerprint set the tls fingerprint for tls handshake, will use utls
// (https://github.com/refraction-networking/utls) to perform the tls handshake,
// which uses the specified clientHelloID to simulate the tls fingerprint.
// Note this is valid for HTTP1 and HTTP2, not HTTP3.
func (c *Client) SetTLSFingerprint(clientHelloID utls.ClientHelloID) *Client {
	fn := func(ctx context.Context, addr string, plainConn net.Conn) (conn net.Conn, tlsState *tls.ConnectionState, err error) {
		colonPos := strings.LastIndex(addr, ":")
		if colonPos == -1 {
			colonPos = len(addr)
		}
		hostname := addr[:colonPos]
		utlsConfig := &utls.Config{ServerName: hostname, NextProtos: c.GetTLSClientConfig().NextProtos}
		uconn := &uTLSConn{utls.UClient(plainConn, utlsConfig, clientHelloID)}
		err = uconn.HandshakeContext(ctx)
		if err != nil {
			return
		}
		cs := uconn.Conn.ConnectionState()
		conn = uconn
		tlsState = &tls.ConnectionState{
			Version:                     cs.Version,
			HandshakeComplete:           cs.HandshakeComplete,
			DidResume:                   cs.DidResume,
			CipherSuite:                 cs.CipherSuite,
			NegotiatedProtocol:          cs.NegotiatedProtocol,
			NegotiatedProtocolIsMutual:  cs.NegotiatedProtocolIsMutual,
			ServerName:                  cs.ServerName,
			PeerCertificates:            cs.PeerCertificates,
			VerifiedChains:              cs.VerifiedChains,
			SignedCertificateTimestamps: cs.SignedCertificateTimestamps,
			OCSPResponse:                cs.OCSPResponse,
			TLSUnique:                   cs.TLSUnique,
		}
		return
	}
	c.SetTLSHandshake(fn)
	return c
}

// SetTLSHandshake set the custom tls handshake function, only valid for HTTP1 and HTTP2, not HTTP3,
// it specifies an optional dial function for tls handshake, it works even if a proxy is set, can be
// used to customize the tls fingerprint.
func (c *Client) SetTLSHandshake(fn func(ctx context.Context, addr string, plainConn net.Conn) (conn net.Conn, tlsState *tls.ConnectionState, err error)) *Client {
	c.t.SetTLSHandshake(fn)
	return c
}

// SetTLSHandshakeTimeout set the TLS handshake timeout.
func (c *Client) SetTLSHandshakeTimeout(timeout time.Duration) *Client {
	c.t.SetTLSHandshakeTimeout(timeout)
	return c
}

// EnableForceHTTP1 enable force using HTTP1 (disabled by default).
func (c *Client) EnableForceHTTP1() *Client {
	c.t.EnableForceHTTP1()
	return c
}

// EnableForceHTTP2 enable force using HTTP2 for https requests
// (disabled by default).
func (c *Client) EnableForceHTTP2() *Client {
	c.t.EnableForceHTTP2()
	return c
}

// EnableForceHTTP3 enable force using HTTP3 for https requests
// (disabled by default).
func (c *Client) EnableForceHTTP3() *Client {
	c.t.EnableForceHTTP3()
	return c
}

// DisableForceHttpVersion disable force using specified http
// version (disabled by default).
func (c *Client) DisableForceHttpVersion() *Client {
	c.t.DisableForceHttpVersion()
	return c
}

// EnableH2C enables HTTP/2 over TCP without TLS.
func (c *Client) EnableH2C() *Client {
	c.t.EnableH2C()
	return c
}

// DisableH2C disables HTTP/2 over TCP without TLS.
func (c *Client) DisableH2C() *Client {
	c.t.DisableH2C()
	return c
}

// DisableAllowGetMethodPayload disable sending GET method requests with body.
func (c *Client) DisableAllowGetMethodPayload() *Client {
	c.AllowGetMethodPayload = false
	return c
}

// EnableAllowGetMethodPayload allows sending GET method requests with body.
func (c *Client) EnableAllowGetMethodPayload() *Client {
	c.AllowGetMethodPayload = true
	return c
}

func (c *Client) isPayloadForbid(m string) bool {
	return (m == http.MethodGet && !c.AllowGetMethodPayload) || m == http.MethodHead || m == http.MethodOptions
}

// GetClient returns the underlying `http.Client`.
func (c *Client) GetClient() *http.Client {
	return c.httpClient
}

func (c *Client) getRetryOption() *retryOption {
	if c.retryOption == nil {
		c.retryOption = newDefaultRetryOption()
	}
	return c.retryOption
}

// SetCommonRetryCount enables retry and set the maximum retry count for requests
// fired from the client.
// It will retry infinitely if count is negative.
func (c *Client) SetCommonRetryCount(count int) *Client {
	c.getRetryOption().MaxRetries = count
	return c
}

// SetCommonRetryInterval sets the custom GetRetryIntervalFunc for requests fired
// from the client, you can use this to implement your own backoff retry algorithm.
// For example:
//
//		 req.SetCommonRetryInterval(func(resp *req.Response, attempt int) time.Duration {
//	     sleep := 0.01 * math.Exp2(float64(attempt))
//	     return time.Duration(math.Min(2, sleep)) * time.Second
//		 })
func (c *Client) SetCommonRetryInterval(getRetryIntervalFunc GetRetryIntervalFunc) *Client {
	c.getRetryOption().GetRetryInterval = getRetryIntervalFunc
	return c
}

// SetCommonRetryFixedInterval set retry to use a fixed interval for requests
// fired from the client.
func (c *Client) SetCommonRetryFixedInterval(interval time.Duration) *Client {
	c.getRetryOption().GetRetryInterval = func(resp *Response, attempt int) time.Duration {
		return interval
	}
	return c
}

// SetCommonRetryBackoffInterval set retry to use a capped exponential backoff
// with jitter for requests fired from the client.
// https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/
func (c *Client) SetCommonRetryBackoffInterval(min, max time.Duration) *Client {
	c.getRetryOption().GetRetryInterval = backoffInterval(min, max)
	return c
}

// SetCommonRetryHook set the retry hook which will be executed before a retry.
// It will override other retry hooks if any been added before.
func (c *Client) SetCommonRetryHook(hook RetryHookFunc) *Client {
	c.getRetryOption().RetryHooks = []RetryHookFunc{hook}
	return c
}

// AddCommonRetryHook adds a retry hook for requests fired from the client,
// which will be executed before a retry.
func (c *Client) AddCommonRetryHook(hook RetryHookFunc) *Client {
	ro := c.getRetryOption()
	ro.RetryHooks = append(ro.RetryHooks, hook)
	return c
}

// SetCommonRetryCondition sets the retry condition, which determines whether the
// request should retry.
// It will override other retry conditions if any been added before.
func (c *Client) SetCommonRetryCondition(condition RetryConditionFunc) *Client {
	c.getRetryOption().RetryConditions = []RetryConditionFunc{condition}
	return c
}

// AddCommonRetryCondition adds a retry condition, which determines whether the
// request should retry.
func (c *Client) AddCommonRetryCondition(condition RetryConditionFunc) *Client {
	ro := c.getRetryOption()
	ro.RetryConditions = append(ro.RetryConditions, condition)
	return c
}

// SetUnixSocket set client to dial connection use unix socket.
// For example:
//
// client.SetUnixSocket("/var/run/custom.sock")
func (c *Client) SetUnixSocket(file string) *Client {
	return c.SetDial(func(ctx context.Context, network, addr string) (net.Conn, error) {
		var d net.Dialer
		return d.DialContext(ctx, "unix", file)
	})
}

// DisableHTTP3 disables the http3 protocol.
func (c *Client) DisableHTTP3() *Client {
	c.t.DisableHTTP3()
	return c
}

// EnableHTTP3 enables the http3 protocol.
func (c *Client) EnableHTTP3() *Client {
	c.t.EnableHTTP3()
	return c
}

// SetHTTP2MaxHeaderListSize set the http2 MaxHeaderListSize,
// which is the http2 SETTINGS_MAX_HEADER_LIST_SIZE to
// send in the initial settings frame. It is how many bytes
// of response headers are allowed. Unlike the http2 spec, zero here
// means to use a default limit (currently 10MB). If you actually
// want to advertise an unlimited value to the peer, Transport
// interprets the highest possible value here (0xffffffff or 1<<32-1)
// to mean no limit.
func (c *Client) SetHTTP2MaxHeaderListSize(max uint32) *Client {
	c.t.SetHTTP2MaxHeaderListSize(max)
	return c
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
func (c *Client) SetHTTP2StrictMaxConcurrentStreams(strict bool) *Client {
	c.t.SetHTTP2StrictMaxConcurrentStreams(strict)
	return c
}

// SetHTTP2ReadIdleTimeout set the http2 ReadIdleTimeout,
// which is the timeout after which a health check using ping
// frame will be carried out if no frame is received on the connection.
// Note that a ping response will is considered a received frame, so if
// there is no other traffic on the connection, the health check will
// be performed every ReadIdleTimeout interval.
// If zero, no health check is performed.
func (c *Client) SetHTTP2ReadIdleTimeout(timeout time.Duration) *Client {
	c.t.SetHTTP2ReadIdleTimeout(timeout)
	return c
}

// SetHTTP2PingTimeout set the http2 PingTimeout, which is the timeout
// after which the connection will be closed if a response to Ping is
// not received.
// Defaults to 15s
func (c *Client) SetHTTP2PingTimeout(timeout time.Duration) *Client {
	c.t.SetHTTP2PingTimeout(timeout)
	return c
}

// SetHTTP2WriteByteTimeout set the http2 WriteByteTimeout, which is the
// timeout after which the connection will be closed no data can be written
// to it. The timeout begins when data is available to write, and is
// extended whenever any bytes are written.
func (c *Client) SetHTTP2WriteByteTimeout(timeout time.Duration) *Client {
	c.t.SetHTTP2WriteByteTimeout(timeout)
	return c
}

// NewClient is the alias of C
func NewClient() *Client {
	return C()
}

// Clone copy and returns the Client
func (c *Client) Clone() *Client {
	cc := *c

	// clone Transport
	cc.t = c.t.Clone()
	cc.initTransport()

	// clone http.Client
	client := *c.httpClient
	client.Transport = cc.t
	cc.httpClient = &client

	// clone client middleware
	if len(cc.roundTripWrappers) > 0 {
		cc.wrappedRoundTrip = roundTripImpl{&cc}
		for _, w := range cc.roundTripWrappers {
			cc.wrappedRoundTrip = w(cc.wrappedRoundTrip)
		}
	}

	// clone other fields that may need to be cloned
	cc.Headers = cloneHeaders(c.Headers)
	cc.Cookies = cloneCookies(c.Cookies)
	cc.PathParams = cloneMap(c.PathParams)
	cc.QueryParams = cloneUrlValues(c.QueryParams)
	cc.FormData = cloneUrlValues(c.FormData)
	cc.beforeRequest = cloneRequestMiddleware(c.beforeRequest)
	cc.udBeforeRequest = cloneRequestMiddleware(c.udBeforeRequest)
	cc.afterResponse = cloneResponseMiddleware(c.afterResponse)
	cc.dumpOptions = c.dumpOptions.Clone()
	cc.retryOption = c.retryOption.Clone()
	return &cc
}

// C create a new client.
func C() *Client {
	t := T()

	jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	httpClient := &http.Client{
		Transport: t,
		Jar:       jar,
		Timeout:   2 * time.Minute,
	}
	beforeRequest := []RequestMiddleware{
		parseRequestHeader,
		parseRequestURL,
		parseRequestBody,
		parseRequestCookie,
	}
	afterResponse := []ResponseMiddleware{
		parseResponseBody,
		handleDownload,
	}
	c := &Client{
		AllowGetMethodPayload: true,
		beforeRequest:         beforeRequest,
		afterResponse:         afterResponse,
		log:                   createDefaultLogger(),
		httpClient:            httpClient,
		t:                     t,
		jsonMarshal:           json.Marshal,
		jsonUnmarshal:         json.Unmarshal,
		xmlMarshal:            xml.Marshal,
		xmlUnmarshal:          xml.Unmarshal,
	}
	httpClient.CheckRedirect = c.defaultCheckRedirect

	c.initTransport()
	return c
}

func (c *Client) initTransport() {
	c.t.Debugf = func(format string, v ...interface{}) {
		if c.DebugLog {
			c.log.Debugf(format, v...)
		}
	}
}

// RoundTripper is the interface of req's Client.
type RoundTripper interface {
	RoundTrip(*Request) (*Response, error)
}

// RoundTripFunc is a RoundTripper implementation, which is a simple function.
type RoundTripFunc func(req *Request) (resp *Response, err error)

// RoundTrip implements RoundTripper.
func (fn RoundTripFunc) RoundTrip(req *Request) (*Response, error) {
	return fn(req)
}

// RoundTripWrapper is client middleware function.
type RoundTripWrapper func(rt RoundTripper) RoundTripper

// RoundTripWrapperFunc is client middleware function, more convenient than RoundTripWrapper.
type RoundTripWrapperFunc func(rt RoundTripper) RoundTripFunc

func (f RoundTripWrapperFunc) wrapper() RoundTripWrapper {
	return func(rt RoundTripper) RoundTripper {
		return f(rt)
	}
}

// WrapRoundTripFunc adds a client middleware function that will give the caller
// an opportunity to wrap the underlying http.RoundTripper.
func (c *Client) WrapRoundTripFunc(funcs ...RoundTripWrapperFunc) *Client {
	var wrappers []RoundTripWrapper
	for _, fn := range funcs {
		wrappers = append(wrappers, fn.wrapper())
	}
	return c.WrapRoundTrip(wrappers...)
}

type roundTripImpl struct {
	*Client
}

func (r roundTripImpl) RoundTrip(req *Request) (resp *Response, err error) {
	return r.roundTrip(req)
}

// WrapRoundTrip adds a client middleware function that will give the caller
// an opportunity to wrap the underlying http.RoundTripper.
func (c *Client) WrapRoundTrip(wrappers ...RoundTripWrapper) *Client {
	if len(wrappers) == 0 {
		return c
	}
	if c.wrappedRoundTrip == nil {
		c.roundTripWrappers = wrappers
		c.wrappedRoundTrip = roundTripImpl{c}
	} else {
		c.roundTripWrappers = append(c.roundTripWrappers, wrappers...)
	}
	for _, w := range wrappers {
		c.wrappedRoundTrip = w(c.wrappedRoundTrip)
	}
	return c
}

// RoundTrip implements RoundTripper
func (c *Client) roundTrip(r *Request) (resp *Response, err error) {
	resp = &Response{Request: r}
	defer func() {
		if err != nil {
			resp.Err = err
		} else {
			err = resp.Err
		}
	}()

	// setup trace
	if r.trace == nil && r.client.trace {
		r.trace = &clientTrace{}
	}

	ctx := r.ctx

	if r.trace != nil {
		ctx = r.trace.createContext(r.Context())
	}

	// setup url and host
	var host string
	if h := r.getHeader("Host"); h != "" {
		host = h // Host header override
	} else {
		host = r.URL.Host
	}

	// setup header
	contentLength := int64(len(r.Body))

	var reqBody io.ReadCloser
	if r.GetBody != nil {
		reqBody, resp.Err = r.GetBody()
		if resp.Err != nil {
			return
		}
	}
	req := &http.Request{
		Method:        r.Method,
		Header:        r.Headers.Clone(),
		URL:           r.URL,
		Host:          host,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		ContentLength: contentLength,
		Body:          reqBody,
		GetBody:       r.GetBody,
		Close:         r.close,
	}
	for _, cookie := range r.Cookies {
		req.AddCookie(cookie)
	}
	if r.isSaveResponse && r.downloadCallback != nil {
		var wrap wrapResponseBodyFunc = func(rc io.ReadCloser) io.ReadCloser {
			return &callbackReader{
				ReadCloser: rc,
				callback: func(read int64) {
					r.downloadCallback(DownloadInfo{
						Response:       resp,
						DownloadedSize: read,
					})
				},
				lastTime: time.Now(),
				interval: r.downloadCallbackInterval,
			}
		}
		if ctx == nil {
			ctx = context.Background()
		}
		ctx = context.WithValue(ctx, wrapResponseBodyKey, wrap)
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}
	r.RawRequest = req
	r.StartTime = time.Now()

	var httpResponse *http.Response
	httpResponse, resp.Err = c.httpClient.Do(r.RawRequest)
	resp.Response = httpResponse

	// auto-read response body if possible
	if resp.Err == nil && !c.disableAutoReadResponse && !r.isSaveResponse && !r.disableAutoReadResponse && resp.StatusCode > 199 {
		resp.ToBytes()
		// restore body for re-reads
		resp.Body = io.NopCloser(bytes.NewReader(resp.body))
	}

	for _, f := range c.afterResponse {
		if e := f(c, resp); e != nil {
			resp.Err = e
		}
	}
	return
}
