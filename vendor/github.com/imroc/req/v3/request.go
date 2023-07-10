package req

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	urlpkg "net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/imroc/req/v3/internal/dump"
	"github.com/imroc/req/v3/internal/header"
	"github.com/imroc/req/v3/internal/util"
)

// Request struct is used to compose and fire individual request from
// req client. Request provides lots of chainable settings which can
// override client level settings.
type Request struct {
	PathParams   map[string]string
	QueryParams  urlpkg.Values
	FormData     urlpkg.Values
	Headers      http.Header
	Cookies      []*http.Cookie
	Result       interface{}
	Error        interface{}
	RawRequest   *http.Request
	StartTime    time.Time
	RetryAttempt int
	RawURL       string // read only
	Method       string
	Body         []byte
	GetBody      GetContentFunc
	// URL is an auto-generated field, and is nil in request middleware (OnBeforeRequest),
	// consider using RawURL if you want, it's not nil in client middleware (WrapRoundTripFunc)
	URL *urlpkg.URL

	isMultiPart              bool
	disableAutoReadResponse  bool
	forceChunkedEncoding     bool
	isSaveResponse           bool
	close                    bool
	error                    error
	client                   *Client
	uploadCallback           UploadCallback
	uploadCallbackInterval   time.Duration
	downloadCallback         DownloadCallback
	downloadCallbackInterval time.Duration
	unReplayableBody         io.ReadCloser
	retryOption              *retryOption
	bodyReadCloser           io.ReadCloser
	dumpOptions              *DumpOptions
	marshalBody              interface{}
	ctx                      context.Context
	uploadFiles              []*FileUpload
	uploadReader             []io.ReadCloser
	outputFile               string
	output                   io.Writer
	trace                    *clientTrace
	dumpBuffer               *bytes.Buffer
	responseReturnTime       time.Time
	afterResponse            []ResponseMiddleware
}

type GetContentFunc func() (io.ReadCloser, error)

func (r *Request) getHeader(key string) string {
	if r.Headers == nil {
		return ""
	}
	return r.Headers.Get(key)
}

// TraceInfo returns the trace information, only available if trace is enabled
// (see Request.EnableTrace and Client.EnableTraceAll).
func (r *Request) TraceInfo() TraceInfo {
	ct := r.trace

	if ct == nil {
		return TraceInfo{}
	}

	ti := TraceInfo{
		IsConnReused:  ct.gotConnInfo.Reused,
		IsConnWasIdle: ct.gotConnInfo.WasIdle,
		ConnIdleTime:  ct.gotConnInfo.IdleTime,
	}

	endTime := ct.endTime
	if endTime.IsZero() { // in case timeout
		endTime = r.responseReturnTime
	}

	if !ct.tlsHandshakeStart.IsZero() {
		if !ct.tlsHandshakeDone.IsZero() {
			ti.TLSHandshakeTime = ct.tlsHandshakeDone.Sub(ct.tlsHandshakeStart)
		} else {
			ti.TLSHandshakeTime = endTime.Sub(ct.tlsHandshakeStart)
		}
	}

	if ct.gotConnInfo.Reused {
		ti.TotalTime = endTime.Sub(ct.getConn)
	} else {
		if ct.dnsStart.IsZero() {
			ti.TotalTime = endTime.Sub(r.StartTime)
		} else {
			ti.TotalTime = endTime.Sub(ct.dnsStart)
		}
	}

	dnsDone := ct.dnsDone
	if dnsDone.IsZero() {
		dnsDone = endTime
	}

	if !ct.dnsStart.IsZero() {
		ti.DNSLookupTime = dnsDone.Sub(ct.dnsStart)
	}

	// Only calculate on successful connections
	if !ct.connectDone.IsZero() {
		ti.TCPConnectTime = ct.connectDone.Sub(dnsDone)
	}

	// Only calculate on successful connections
	if !ct.gotConn.IsZero() {
		ti.ConnectTime = ct.gotConn.Sub(ct.getConn)
	}

	// Only calculate on successful connections
	if !ct.gotFirstResponseByte.IsZero() {
		ti.FirstResponseTime = ct.gotFirstResponseByte.Sub(ct.gotConn)
		ti.ResponseTime = endTime.Sub(ct.gotFirstResponseByte)
	}

	// Capture remote address info when connection is non-nil
	if ct.gotConnInfo.Conn != nil {
		ti.RemoteAddr = ct.gotConnInfo.Conn.RemoteAddr()
	}

	return ti
}

// HeaderToString get all header as string.
func (r *Request) HeaderToString() string {
	return convertHeaderToString(r.Headers)
}

// SetURL set the url for request.
func (r *Request) SetURL(url string) *Request {
	r.RawURL = url
	return r
}

// SetFormDataFromValues set the form data from url.Values, will not
// been used if request method does not allow payload.
func (r *Request) SetFormDataFromValues(data urlpkg.Values) *Request {
	if r.FormData == nil {
		r.FormData = urlpkg.Values{}
	}
	for k, v := range data {
		for _, kv := range v {
			r.FormData.Add(k, kv)
		}
	}
	return r
}

// SetFormData set the form data from a map, will not been used
// if request method does not allow payload.
func (r *Request) SetFormData(data map[string]string) *Request {
	if r.FormData == nil {
		r.FormData = urlpkg.Values{}
	}
	for k, v := range data {
		r.FormData.Set(k, v)
	}
	return r
}

// SetFormDataAnyType set the form data from a map, which value could be any type,
// will convert to string automatically.
// It will not been used if request method does not allow payload.
func (r *Request) SetFormDataAnyType(data map[string]interface{}) *Request {
	if r.FormData == nil {
		r.FormData = urlpkg.Values{}
	}
	for k, v := range data {
		r.FormData.Set(k, fmt.Sprint(v))
	}
	return r
}

// SetCookies set http cookies for the request.
func (r *Request) SetCookies(cookies ...*http.Cookie) *Request {
	r.Cookies = append(r.Cookies, cookies...)
	return r
}

// SetQueryString set URL query parameters for the request using
// raw query string.
func (r *Request) SetQueryString(query string) *Request {
	params, err := urlpkg.ParseQuery(strings.TrimSpace(query))
	if err != nil {
		r.client.log.Warnf("failed to parse query string (%s): %v", query, err)
		return r
	}
	if r.QueryParams == nil {
		r.QueryParams = make(urlpkg.Values)
	}
	for p, v := range params {
		for _, pv := range v {
			r.QueryParams.Add(p, pv)
		}
	}
	return r
}

// SetFileReader set up a multipart form with a reader to upload file.
func (r *Request) SetFileReader(paramName, filename string, reader io.Reader) *Request {
	r.SetFileUpload(FileUpload{
		ParamName: paramName,
		FileName:  filename,
		GetFileContent: func() (io.ReadCloser, error) {
			if rc, ok := reader.(io.ReadCloser); ok {
				return rc, nil
			}
			return io.NopCloser(reader), nil
		},
	})
	return r
}

// SetFileBytes set up a multipart form with given []byte to upload.
func (r *Request) SetFileBytes(paramName, filename string, content []byte) *Request {
	return r.SetFileReader(paramName, filename, bytes.NewReader(content))
}

// SetFiles set up a multipart form from a map to upload, which
// key is the parameter name, and value is the file path.
func (r *Request) SetFiles(files map[string]string) *Request {
	for k, v := range files {
		r.SetFile(k, v)
	}
	return r
}

// SetFile set up a multipart form from file path to upload,
// which read file from filePath automatically to upload.
func (r *Request) SetFile(paramName, filePath string) *Request {
	file, err := os.Open(filePath)
	if err != nil {
		r.client.log.Errorf("failed to open %s: %v", filePath, err)
		r.appendError(err)
		return r
	}
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		r.client.log.Errorf("failed to stat file %s: %v", filePath, err)
		r.appendError(err)
		return r
	}
	r.isMultiPart = true
	return r.SetFileUpload(FileUpload{
		ParamName: paramName,
		FileName:  filepath.Base(filePath),
		GetFileContent: func() (io.ReadCloser, error) {
			if r.RetryAttempt > 0 {
				file, err = os.Open(filePath)
				if err != nil {
					return nil, err
				}
			}
			return file, nil
		},
		FileSize: fileInfo.Size(),
	})
}

var errMissingParamName = errors.New("missing param name in multipart file upload")
var errMissingFileName = errors.New("missing filename in multipart file upload")
var errMissingFileContent = errors.New("missing file content in multipart file upload")

// SetFileUpload set the fully custimized multipart file upload options.
func (r *Request) SetFileUpload(uploads ...FileUpload) *Request {
	r.isMultiPart = true
	for _, upload := range uploads {
		shouldAppend := true
		if upload.ParamName == "" {
			r.appendError(errMissingParamName)
			shouldAppend = false
		}
		if upload.FileName == "" {
			r.appendError(errMissingFileName)
			shouldAppend = false
		}
		if upload.GetFileContent == nil {
			r.appendError(errMissingFileContent)
			shouldAppend = false
		}
		if shouldAppend {
			r.uploadFiles = append(r.uploadFiles, &upload)
		}
	}
	return r
}

// SetUploadCallback set the UploadCallback which will be invoked at least
// every 200ms during file upload, usually used to show upload progress.
func (r *Request) SetUploadCallback(callback UploadCallback) *Request {
	return r.SetUploadCallbackWithInterval(callback, 200*time.Millisecond)
}

// SetUploadCallbackWithInterval set the UploadCallback which will be invoked at least
// every `minInterval` during file upload, usually used to show upload progress.
func (r *Request) SetUploadCallbackWithInterval(callback UploadCallback, minInterval time.Duration) *Request {
	if callback == nil {
		return r
	}
	r.forceChunkedEncoding = true
	r.uploadCallback = callback
	r.uploadCallbackInterval = minInterval
	return r
}

// SetDownloadCallback set the DownloadCallback which will be invoked at least
// every 200ms during file upload, usually used to show download progress.
func (r *Request) SetDownloadCallback(callback DownloadCallback) *Request {
	return r.SetDownloadCallbackWithInterval(callback, 200*time.Millisecond)
}

// SetDownloadCallbackWithInterval set the DownloadCallback which will be invoked at least
// every `minInterval` during file upload, usually used to show download progress.
func (r *Request) SetDownloadCallbackWithInterval(callback DownloadCallback, minInterval time.Duration) *Request {
	if callback == nil {
		return r
	}
	r.downloadCallback = callback
	r.downloadCallbackInterval = minInterval
	return r
}

// SetResult set the result that response Body will be unmarshalled to if
// no error occurs and Response.ResultState() returns SuccessState, by default
// it requires HTTP status `code >= 200 && code <= 299`, you can also use
// Request.SetResultStateCheckFunc or Client.SetResultStateCheckFunc to customize
// the result state check logic.
//
// Deprecated: Use SetSuccessResult instead.
func (r *Request) SetResult(result interface{}) *Request {
	return r.SetSuccessResult(result)
}

// SetSuccessResult set the result that response Body will be unmarshalled to if
// no error occurs and Response.ResultState() returns SuccessState, by default
// it requires HTTP status `code >= 200 && code <= 299`, you can also use
// Request.SetResultStateCheckFunc or Client.SetResultStateCheckFunc to customize
// the result state check logic.
func (r *Request) SetSuccessResult(result interface{}) *Request {
	if result == nil {
		return r
	}
	r.Result = util.GetPointer(result)
	return r
}

// SetError set the result that response body will be unmarshalled to if
// no error occurs and Response.ResultState() returns ErrorState, by default
// it requires HTTP status `code >= 400`, you can also use Request.SetResultStateCheckFunc
// or Client.SetResultStateCheckFunc to customize the result state check logic.
//
// Deprecated: Use SetErrorResult result.
func (r *Request) SetError(err interface{}) *Request {
	return r.SetErrorResult(err)
}

// SetErrorResult set the result that response body will be unmarshalled to if
// no error occurs and Response.ResultState() returns ErrorState, by default
// it requires HTTP status `code >= 400`, you can also use Request.SetResultStateCheckFunc
// or Client.SetResultStateCheckFunc to customize the result state check logic.
func (r *Request) SetErrorResult(err interface{}) *Request {
	if err == nil {
		return r
	}
	r.Error = util.GetPointer(err)
	return r
}

// SetBearerAuthToken set bearer auth token for the request.
func (r *Request) SetBearerAuthToken(token string) *Request {
	return r.SetHeader(header.Authorization, "Bearer "+token)
}

// SetBasicAuth set basic auth for the request.
func (r *Request) SetBasicAuth(username, password string) *Request {
	return r.SetHeader(header.Authorization, util.BasicAuthHeaderValue(username, password))
}

// SetDigestAuth sets the Digest Access auth scheme for the HTTP request. If a server responds with 401 and sends a
// Digest challenge in the WWW-Authenticate Header, the request will be resent with the appropriate Authorization Header.
//
// For Example: To set the Digest scheme with username "roc" and password "123456"
//
//	client.R().SetDigestAuth("roc", "123456")
//
// Information about Digest Access Authentication can be found in RFC7616:
//
//	https://datatracker.ietf.org/doc/html/rfc7616
//
// This method overrides the username and password set by method `Client.SetCommonDigestAuth`.
func (r *Request) SetDigestAuth(username, password string) *Request {
	r.OnAfterResponse(handleDigestAuthFunc(username, password))
	return r
}

// OnAfterResponse add a response middleware which hooks after response received.
func (r *Request) OnAfterResponse(m ResponseMiddleware) *Request {
	r.afterResponse = append(r.afterResponse, m)
	return r
}

// SetHeaders set headers from a map for the request.
func (r *Request) SetHeaders(hdrs map[string]string) *Request {
	for k, v := range hdrs {
		r.SetHeader(k, v)
	}
	return r
}

// SetHeader set a header for the request.
func (r *Request) SetHeader(key, value string) *Request {
	if r.Headers == nil {
		r.Headers = make(http.Header)
	}
	r.Headers.Set(key, value)
	return r
}

// SetHeadersNonCanonical set headers from a map for the request which key is a
// non-canonical key (keep case unchanged), only valid for HTTP/1.1.
func (r *Request) SetHeadersNonCanonical(hdrs map[string]string) *Request {
	for k, v := range hdrs {
		r.SetHeaderNonCanonical(k, v)
	}
	return r
}

// SetHeaderNonCanonical set a header for the request which key is a
// non-canonical key (keep case unchanged), only valid for HTTP/1.1.
func (r *Request) SetHeaderNonCanonical(key, value string) *Request {
	if r.Headers == nil {
		r.Headers = make(http.Header)
	}
	r.Headers[key] = append(r.Headers[key], value)
	return r
}

// SetOutputFile set the file that response Body will be downloaded to.
func (r *Request) SetOutputFile(file string) *Request {
	r.isSaveResponse = true
	r.outputFile = file
	return r
}

// SetOutput set the io.Writer that response Body will be downloaded to.
func (r *Request) SetOutput(output io.Writer) *Request {
	if output == nil {
		r.client.log.Warnf("nil io.Writer is not allowed in SetOutput")
		return r
	}
	r.output = output
	r.isSaveResponse = true
	return r
}

// SetQueryParams set URL query parameters from a map for the request.
func (r *Request) SetQueryParams(params map[string]string) *Request {
	for k, v := range params {
		r.SetQueryParam(k, v)
	}
	return r
}

// SetQueryParamsAnyType set URL query parameters from a map for the request.
// The value of map is any type, will be convert to string automatically.
func (r *Request) SetQueryParamsAnyType(params map[string]interface{}) *Request {
	for k, v := range params {
		r.SetQueryParam(k, fmt.Sprint(v))
	}
	return r
}

// SetQueryParam set an URL query parameter for the request.
func (r *Request) SetQueryParam(key, value string) *Request {
	if r.QueryParams == nil {
		r.QueryParams = make(urlpkg.Values)
	}
	r.QueryParams.Set(key, value)
	return r
}

// AddQueryParam add a URL query parameter for the request.
func (r *Request) AddQueryParam(key, value string) *Request {
	if r.QueryParams == nil {
		r.QueryParams = make(urlpkg.Values)
	}
	r.QueryParams.Add(key, value)
	return r
}

// AddQueryParams add one or more values of specified URL query parameter for the request.
func (r *Request) AddQueryParams(key string, values ...string) *Request {
	if r.QueryParams == nil {
		r.QueryParams = make(urlpkg.Values)
	}
	vs := r.QueryParams[key]
	vs = append(vs, values...)
	r.QueryParams[key] = vs
	return r
}

// SetPathParams set URL path parameters from a map for the request.
func (r *Request) SetPathParams(params map[string]string) *Request {
	for key, value := range params {
		r.SetPathParam(key, value)
	}
	return r
}

// SetPathParam set a URL path parameter for the request.
func (r *Request) SetPathParam(key, value string) *Request {
	if r.PathParams == nil {
		r.PathParams = make(map[string]string)
	}
	r.PathParams[key] = value
	return r
}

func (r *Request) appendError(err error) {
	r.error = multierror.Append(r.error, err)
}

var errRetryableWithUnReplayableBody = errors.New("retryable request should not have unreplayable Body (io.Reader)")

func (r *Request) newErrorResponse(err error) *Response {
	resp := &Response{Request: r}
	resp.Err = err
	return resp
}

// Do fires http request, 0 or 1 context is allowed, and returns the *Response which
// is always not nil, and Response.Err is not nil if error occurs.
func (r *Request) Do(ctx ...context.Context) *Response {
	if len(ctx) > 0 && ctx[0] != nil {
		r.ctx = ctx[0]
	}

	defer func() {
		r.responseReturnTime = time.Now()
	}()
	if r.error != nil {
		return r.newErrorResponse(r.error)
	}
	if r.retryOption != nil && r.retryOption.MaxRetries != 0 && r.unReplayableBody != nil { // retryable request should not have unreplayable Body
		return r.newErrorResponse(errRetryableWithUnReplayableBody)
	}
	resp, _ := r.do()
	return resp
}

func (r *Request) do() (resp *Response, err error) {
	defer func() {
		if resp == nil {
			resp = &Response{Request: r}
		}
		if err != nil {
			resp.Err = err
		}
	}()

	for {
		if r.Headers == nil {
			r.Headers = make(http.Header)
		}
		for _, f := range r.client.udBeforeRequest {
			if err = f(r.client, r); err != nil {
				return
			}
		}
		for _, f := range r.client.beforeRequest {
			if err = f(r.client, r); err != nil {
				return
			}
		}

		if r.client.wrappedRoundTrip != nil {
			resp, err = r.client.wrappedRoundTrip.RoundTrip(r)
		} else {
			resp, err = r.client.roundTrip(r)
		}

		for _, f := range r.afterResponse {
			if err = f(r.client, resp); err != nil {
				return
			}
		}

		if r.retryOption == nil || (r.RetryAttempt >= r.retryOption.MaxRetries && r.retryOption.MaxRetries >= 0) { // absolutely cannot retry.
			return
		}

		// check retry whether is needed.
		needRetry := err != nil                                   // default behaviour: retry if error occurs
		for _, condition := range r.retryOption.RetryConditions { // override default behaviour if custom RetryConditions has been set.
			needRetry = condition(resp, err)
			if needRetry {
				break
			}
		}
		if !needRetry { // no retry is needed.
			return
		}

		// need retry, attempt to retry
		r.RetryAttempt++
		for _, hook := range r.retryOption.RetryHooks { // run retry hooks
			hook(resp, err)
		}
		time.Sleep(r.retryOption.GetRetryInterval(resp, r.RetryAttempt))

		// clean up before retry
		if r.dumpBuffer != nil {
			r.dumpBuffer.Reset()
		}
		if r.trace != nil {
			r.trace = &clientTrace{}
		}
		resp.body = nil
		resp.result = nil
		resp.error = nil
	}

	return
}

// Send fires http request with specified method and url, returns the
// *Response which is always not nil, and the error is not nil if error occurs.
func (r *Request) Send(method, url string) (*Response, error) {
	r.Method = method
	r.RawURL = url
	resp := r.Do()
	return resp, resp.Err
}

// MustGet like Get, panic if error happens, should only be used to
// test without error handling.
func (r *Request) MustGet(url string) *Response {
	resp, err := r.Get(url)
	if err != nil {
		panic(err)
	}
	return resp
}

// Get fires http request with GET method and the specified URL.
func (r *Request) Get(url string) (*Response, error) {
	return r.Send(http.MethodGet, url)
}

// MustPost like Post, panic if error happens. should only be used to
// test without error handling.
func (r *Request) MustPost(url string) *Response {
	resp, err := r.Post(url)
	if err != nil {
		panic(err)
	}
	return resp
}

// Post fires http request with POST method and the specified URL.
func (r *Request) Post(url string) (*Response, error) {
	return r.Send(http.MethodPost, url)
}

// MustPut like Put, panic if error happens, should only be used to
// test without error handling.
func (r *Request) MustPut(url string) *Response {
	resp, err := r.Put(url)
	if err != nil {
		panic(err)
	}
	return resp
}

// Put fires http request with PUT method and the specified URL.
func (r *Request) Put(url string) (*Response, error) {
	return r.Send(http.MethodPut, url)
}

// MustPatch like Patch, panic if error happens, should only be used
// to test without error handling.
func (r *Request) MustPatch(url string) *Response {
	resp, err := r.Patch(url)
	if err != nil {
		panic(err)
	}
	return resp
}

// Patch fires http request with PATCH method and the specified URL.
func (r *Request) Patch(url string) (*Response, error) {
	return r.Send(http.MethodPatch, url)
}

// MustDelete like Delete, panic if error happens, should only be used
// to test without error handling.
func (r *Request) MustDelete(url string) *Response {
	resp, err := r.Delete(url)
	if err != nil {
		panic(err)
	}
	return resp
}

// Delete fires http request with DELETE method and the specified URL.
func (r *Request) Delete(url string) (*Response, error) {
	return r.Send(http.MethodDelete, url)
}

// MustOptions like Options, panic if error happens, should only be
// used to test without error handling.
func (r *Request) MustOptions(url string) *Response {
	resp, err := r.Options(url)
	if err != nil {
		panic(err)
	}
	return resp
}

// Options fires http request with OPTIONS method and the specified URL.
func (r *Request) Options(url string) (*Response, error) {
	return r.Send(http.MethodOptions, url)
}

// MustHead like Head, panic if error happens, should only be used
// to test without error handling.
func (r *Request) MustHead(url string) *Response {
	resp, err := r.Send(http.MethodHead, url)
	if err != nil {
		panic(err)
	}
	return resp
}

// Head fires http request with HEAD method and the specified URL.
func (r *Request) Head(url string) (*Response, error) {
	return r.Send(http.MethodHead, url)
}

// SetBody set the request Body, accepts string, []byte, io.Reader, map and struct.
func (r *Request) SetBody(body interface{}) *Request {
	if body == nil {
		return r
	}
	switch b := body.(type) {
	case io.ReadCloser:
		r.unReplayableBody = b
		r.GetBody = func() (io.ReadCloser, error) {
			return r.unReplayableBody, nil
		}
	case io.Reader:
		r.unReplayableBody = io.NopCloser(b)
		r.GetBody = func() (io.ReadCloser, error) {
			return r.unReplayableBody, nil
		}
	case []byte:
		r.SetBodyBytes(b)
	case string:
		r.SetBodyString(b)
	case func() (io.ReadCloser, error):
		r.GetBody = b
	case GetContentFunc:
		r.GetBody = b
	default:
		t := reflect.TypeOf(body)
		switch t.Kind() {
		case reflect.Ptr, reflect.Struct, reflect.Map, reflect.Slice, reflect.Array:
			r.marshalBody = body
		default:
			r.SetBodyString(fmt.Sprint(body))
		}
	}
	return r
}

// SetBodyBytes set the request Body as []byte.
func (r *Request) SetBodyBytes(body []byte) *Request {
	r.Body = body
	r.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(body)), nil
	}
	return r
}

// SetBodyString set the request Body as string.
func (r *Request) SetBodyString(body string) *Request {
	return r.SetBodyBytes([]byte(body))
}

// SetBodyJsonString set the request Body as string and set Content-Type header
// as "application/json; charset=utf-8"
func (r *Request) SetBodyJsonString(body string) *Request {
	return r.SetBodyJsonBytes([]byte(body))
}

// SetBodyJsonBytes set the request Body as []byte and set Content-Type header
// as "application/json; charset=utf-8"
func (r *Request) SetBodyJsonBytes(body []byte) *Request {
	r.SetContentType(header.JsonContentType)
	return r.SetBodyBytes(body)
}

// SetBodyJsonMarshal set the request Body that marshaled from object, and
// set Content-Type header as "application/json; charset=utf-8"
func (r *Request) SetBodyJsonMarshal(v interface{}) *Request {
	b, err := r.client.jsonMarshal(v)
	if err != nil {
		r.appendError(err)
		return r
	}
	return r.SetBodyJsonBytes(b)
}

// SetBodyXmlString set the request Body as string and set Content-Type header
// as "text/xml; charset=utf-8"
func (r *Request) SetBodyXmlString(body string) *Request {
	return r.SetBodyXmlBytes([]byte(body))
}

// SetBodyXmlBytes set the request Body as []byte and set Content-Type header
// as "text/xml; charset=utf-8"
func (r *Request) SetBodyXmlBytes(body []byte) *Request {
	r.SetContentType(header.XmlContentType)
	return r.SetBodyBytes(body)
}

// SetBodyXmlMarshal set the request Body that marshaled from object, and
// set Content-Type header as "text/xml; charset=utf-8"
func (r *Request) SetBodyXmlMarshal(v interface{}) *Request {
	b, err := r.client.xmlMarshal(v)
	if err != nil {
		r.appendError(err)
		return r
	}
	return r.SetBodyXmlBytes(b)
}

// SetContentType set the `Content-Type` for the request.
func (r *Request) SetContentType(contentType string) *Request {
	return r.SetHeader(header.ContentType, contentType)
}

// Context method returns the Context if its already set in request
// otherwise it creates new one using `context.Background()`.
func (r *Request) Context() context.Context {
	if r.ctx == nil {
		r.ctx = context.Background()
	}
	return r.ctx
}

// SetContext method sets the context.Context for current Request. It allows
// to interrupt the request execution if ctx.Done() channel is closed.
// See https://blog.golang.org/context article and the "context" package
// documentation.
//
// Attention: make sure call SetContext before EnableDumpXXX if you want to
// dump at the request level.
func (r *Request) SetContext(ctx context.Context) *Request {
	if ctx != nil {
		r.ctx = ctx
	}
	return r
}

// DisableAutoReadResponse disable read response body automatically (enabled by default).
func (r *Request) DisableAutoReadResponse() *Request {
	r.disableAutoReadResponse = true
	return r
}

// EnableAutoReadResponse enable read response body automatically (enabled by default).
func (r *Request) EnableAutoReadResponse() *Request {
	r.disableAutoReadResponse = false
	return r
}

// DisableTrace disables trace.
func (r *Request) DisableTrace() *Request {
	r.trace = nil
	return r
}

// EnableTrace enables trace (http3 currently does not support trace).
func (r *Request) EnableTrace() *Request {
	if r.trace == nil {
		r.trace = &clientTrace{}
	}
	return r
}

func (r *Request) getDumpBuffer() *bytes.Buffer {
	if r.dumpBuffer == nil {
		r.dumpBuffer = new(bytes.Buffer)
	}
	return r.dumpBuffer
}

func (r *Request) getDumpOptions() *DumpOptions {
	if r.dumpOptions == nil {
		r.dumpOptions = &DumpOptions{
			RequestHeader:  true,
			RequestBody:    true,
			ResponseHeader: true,
			ResponseBody:   true,
			Output:         r.getDumpBuffer(),
		}
	}
	return r.dumpOptions
}

// EnableDumpTo enables dump and save to the specified io.Writer.
func (r *Request) EnableDumpTo(output io.Writer) *Request {
	r.getDumpOptions().Output = output
	return r.EnableDump()
}

// EnableDumpToFile enables dump and save to the specified filename.
func (r *Request) EnableDumpToFile(filename string) *Request {
	file, err := os.Create(filename)
	if err != nil {
		r.appendError(err)
		return r
	}
	r.getDumpOptions().Output = file
	return r.EnableDump()
}

// SetDumpOptions sets DumpOptions at request level.
func (r *Request) SetDumpOptions(opt *DumpOptions) *Request {
	if opt == nil {
		return r
	}
	if opt.Output == nil {
		opt.Output = r.getDumpBuffer()
	}
	if r.dumpOptions != nil {
		*r.dumpOptions = *opt
	} else {
		r.dumpOptions = opt
	}
	return r
}

// EnableDump enables dump, including all content for the request and response by default.
func (r *Request) EnableDump() *Request {
	return r.SetContext(context.WithValue(r.Context(), dump.DumperKey, newDumper(r.getDumpOptions())))
}

// EnableDumpWithoutBody enables dump only header for the request and response.
func (r *Request) EnableDumpWithoutBody() *Request {
	o := r.getDumpOptions()
	o.RequestBody = false
	o.ResponseBody = false
	return r.EnableDump()
}

// EnableDumpWithoutHeader enables dump only Body for the request and response.
func (r *Request) EnableDumpWithoutHeader() *Request {
	o := r.getDumpOptions()
	o.RequestHeader = false
	o.ResponseHeader = false
	return r.EnableDump()
}

// EnableDumpWithoutResponse enables dump only request.
func (r *Request) EnableDumpWithoutResponse() *Request {
	o := r.getDumpOptions()
	o.ResponseHeader = false
	o.ResponseBody = false
	return r.EnableDump()
}

// EnableDumpWithoutRequest enables dump only response.
func (r *Request) EnableDumpWithoutRequest() *Request {
	o := r.getDumpOptions()
	o.RequestHeader = false
	o.RequestBody = false
	return r.EnableDump()
}

// EnableDumpWithoutRequestBody enables dump with request Body excluded,
// can be used in upload request to avoid dump the unreadable binary content.
func (r *Request) EnableDumpWithoutRequestBody() *Request {
	o := r.getDumpOptions()
	o.RequestBody = false
	return r.EnableDump()
}

// EnableDumpWithoutResponseBody enables dump with response Body excluded,
// can be used in download request to avoid dump the unreadable binary content.
func (r *Request) EnableDumpWithoutResponseBody() *Request {
	o := r.getDumpOptions()
	o.ResponseBody = false
	return r.EnableDump()
}

// EnableForceChunkedEncoding enables force using chunked encoding when uploading.
func (r *Request) EnableForceChunkedEncoding() *Request {
	r.forceChunkedEncoding = true
	return r
}

// DisableForceChunkedEncoding disables force using chunked encoding when uploading.
func (r *Request) DisableForceChunkedEncoding() *Request {
	r.forceChunkedEncoding = false
	return r
}

// EnableForceMultipart enables force using multipart to upload form data.
func (r *Request) EnableForceMultipart() *Request {
	r.isMultiPart = true
	return r
}

// DisableForceMultipart disables force using multipart to upload form data.
func (r *Request) DisableForceMultipart() *Request {
	r.isMultiPart = true
	return r
}

func (r *Request) getRetryOption() *retryOption {
	if r.retryOption == nil {
		r.retryOption = newDefaultRetryOption()
	}
	return r.retryOption
}

// SetRetryCount enables retry and set the maximum retry count.
// It will retry infinitely if count is negative.
func (r *Request) SetRetryCount(count int) *Request {
	r.getRetryOption().MaxRetries = count
	return r
}

// SetRetryInterval sets the custom GetRetryIntervalFunc, you can use this to
// implement your own backoff retry algorithm.
// For example:
//
//		 req.SetRetryInterval(func(resp *req.Response, attempt int) time.Duration {
//	     sleep := 0.01 * math.Exp2(float64(attempt))
//	     return time.Duration(math.Min(2, sleep)) * time.Second
//		 })
func (r *Request) SetRetryInterval(getRetryIntervalFunc GetRetryIntervalFunc) *Request {
	r.getRetryOption().GetRetryInterval = getRetryIntervalFunc
	return r
}

// SetRetryFixedInterval set retry to use a fixed interval.
func (r *Request) SetRetryFixedInterval(interval time.Duration) *Request {
	r.getRetryOption().GetRetryInterval = func(resp *Response, attempt int) time.Duration {
		return interval
	}
	return r
}

// SetRetryBackoffInterval set retry to use a capped exponential backoff with jitter.
// https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/
func (r *Request) SetRetryBackoffInterval(min, max time.Duration) *Request {
	r.getRetryOption().GetRetryInterval = backoffInterval(min, max)
	return r
}

// SetRetryHook set the retry hook which will be executed before a retry.
// It will override other retry hooks if any been added before (including
// client-level retry hooks).
func (r *Request) SetRetryHook(hook RetryHookFunc) *Request {
	r.getRetryOption().RetryHooks = []RetryHookFunc{hook}
	return r
}

// AddRetryHook adds a retry hook which will be executed before a retry.
func (r *Request) AddRetryHook(hook RetryHookFunc) *Request {
	ro := r.getRetryOption()
	ro.RetryHooks = append(ro.RetryHooks, hook)
	return r
}

// SetRetryCondition sets the retry condition, which determines whether the
// request should retry.
// It will override other retry conditions if any been added before (including
// client-level retry conditions).
func (r *Request) SetRetryCondition(condition RetryConditionFunc) *Request {
	r.getRetryOption().RetryConditions = []RetryConditionFunc{condition}
	return r
}

// AddRetryCondition adds a retry condition, which determines whether the
// request should retry.
func (r *Request) AddRetryCondition(condition RetryConditionFunc) *Request {
	ro := r.getRetryOption()
	ro.RetryConditions = append(ro.RetryConditions, condition)
	return r
}

// SetClient change the client of request dynamically.
func (r *Request) SetClient(client *Client) *Request {
	if client != nil {
		r.client = client
	}
	return r
}

// GetClient returns the current client used by request.
func (r *Request) GetClient() *Client {
	return r.client
}

// EnableCloseConnection closes the connection after sending this
// request and reading its response if set to true in HTTP/1.1 and
// HTTP/2.
//
// Setting this field prevents re-use of TCP connections between
// requests to the same hosts event if EnableKeepAlives() were called.
func (r *Request) EnableCloseConnection() *Request {
	r.close = true
	return r
}
