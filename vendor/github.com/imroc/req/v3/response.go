package req

import (
	"github.com/imroc/req/v3/internal/header"
	"github.com/imroc/req/v3/internal/util"
	"io"
	"net/http"
	"strings"
	"time"
)

// Response is the http response.
type Response struct {
	// The underlying http.Response is embed into Response.
	*http.Response
	// Err is the underlying error, not nil if some error occurs.
	// Usually used in the ResponseMiddleware, you can skip logic in
	// ResponseMiddleware that doesn't need to be executed when err occurs.
	Err error
	// Request is the Response's related Request.
	Request    *Request
	body       []byte
	receivedAt time.Time
	error      interface{}
	result     interface{}
}

// IsSuccess method returns true if no error occurs and HTTP status `code >= 200 and <= 299`
// by default, you can also use Request.SetResultStateCheckFunc to customize the result
// state check logic.
//
// Deprecated: Use IsSuccessState instead.
func (r *Response) IsSuccess() bool {
	return r.IsSuccessState()
}

// IsSuccessState method returns true if no error occurs and HTTP status `code >= 200 and <= 299`
// by default, you can also use Request.SetResultStateCheckFunc to customize the result state
// check logic.
func (r *Response) IsSuccessState() bool {
	if r.Response == nil {
		return false
	}
	return r.ResultState() == SuccessState
}

// IsError method returns true if no error occurs and HTTP status `code >= 400`
// by default, you can also use Request.SetResultStateCheckFunc to customize the result
// state check logic.
//
// Deprecated: Use IsErrorState instead.
func (r *Response) IsError() bool {
	return r.IsErrorState()
}

// IsErrorState method returns true if no error occurs and HTTP status `code >= 400`
// by default, you can also use Request.SetResultStateCheckFunc to customize the result
// state check logic.
func (r *Response) IsErrorState() bool {
	if r.Response == nil {
		return false
	}
	return r.ResultState() == ErrorState
}

// GetContentType return the `Content-Type` header value.
func (r *Response) GetContentType() string {
	if r.Response == nil {
		return ""
	}
	return r.Header.Get(header.ContentType)
}

// ResultState returns the result state.
// By default, it returns SuccessState if HTTP status `code >= 400`, and returns
// ErrorState if HTTP status `code >= 400`, otherwise returns UnknownState.
// You can also use Request.SetResultStateCheckFunc or Client.SetResultStateCheckFunc
// to customize the result state check logic.
func (r *Response) ResultState() ResultState {
	if r.Response == nil {
		return UnknownState
	}
	var resultStateCheckFunc func(resp *Response) ResultState
	if r.Request.client.resultStateCheckFunc != nil {
		resultStateCheckFunc = r.Request.client.resultStateCheckFunc
	} else {
		resultStateCheckFunc = defaultResultStateChecker
	}
	return resultStateCheckFunc(r)
}

// Result returns the automatically unmarshalled object if Request.SetSuccessResult
// is called and ResultState returns SuccessState.
// Otherwise, return nil.
//
// Deprecated: Use SuccessResult instead.
func (r *Response) Result() interface{} {
	return r.SuccessResult()
}

// SuccessResult returns the automatically unmarshalled object if Request.SetSuccessResult
// is called and ResultState returns SuccessState.
// Otherwise, return nil.
func (r *Response) SuccessResult() interface{} {
	return r.result
}

// Error returns the automatically unmarshalled object when Request.SetErrorResult
// or Client.SetCommonErrorResult is called, and ResultState returns ErrorState.
// Otherwise, return nil.
//
// Deprecated: Use ErrorResult instead.
func (r *Response) Error() interface{} {
	return r.error
}

// ErrorResult returns the automatically unmarshalled object when Request.SetErrorResult
// or Client.SetCommonErrorResult is called, and ResultState returns ErrorState.
// Otherwise, return nil.
func (r *Response) ErrorResult() interface{} {
	return r.error
}

// TraceInfo returns the TraceInfo from Request.
func (r *Response) TraceInfo() TraceInfo {
	return r.Request.TraceInfo()
}

// TotalTime returns the total time of the request, from request we sent to response we received.
func (r *Response) TotalTime() time.Duration {
	if r.Request.trace != nil {
		return r.Request.TraceInfo().TotalTime
	}
	if !r.receivedAt.IsZero() {
		return r.receivedAt.Sub(r.Request.StartTime)
	}
	return r.Request.responseReturnTime.Sub(r.Request.StartTime)
}

// ReceivedAt returns the timestamp that response we received.
func (r *Response) ReceivedAt() time.Time {
	return r.receivedAt
}

func (r *Response) setReceivedAt() {
	r.receivedAt = time.Now()
	if r.Request.trace != nil {
		r.Request.trace.endTime = r.receivedAt
	}
}

// UnmarshalJson unmarshalls JSON response body into the specified object.
func (r *Response) UnmarshalJson(v interface{}) error {
	if r.Err != nil {
		return r.Err
	}
	b, err := r.ToBytes()
	if err != nil {
		return err
	}
	return r.Request.client.jsonUnmarshal(b, v)
}

// UnmarshalXml unmarshalls XML response body into the specified object.
func (r *Response) UnmarshalXml(v interface{}) error {
	if r.Err != nil {
		return r.Err
	}
	b, err := r.ToBytes()
	if err != nil {
		return err
	}
	return r.Request.client.xmlUnmarshal(b, v)
}

// Unmarshal unmarshalls response body into the specified object according
// to response `Content-Type`.
func (r *Response) Unmarshal(v interface{}) error {
	if r.Err != nil {
		return r.Err
	}
	v = util.GetPointer(v)
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "json") {
		return r.UnmarshalJson(v)
	} else if strings.Contains(contentType, "xml") {
		return r.UnmarshalXml(v)
	}
	return r.UnmarshalJson(v)
}

// Into unmarshalls response body into the specified object according
// to response `Content-Type`.
func (r *Response) Into(v interface{}) error {
	return r.Unmarshal(v)
}

// Bytes return the response body as []bytes that hava already been read, could be
// nil if not read, the following cases are already read:
//  1. `Request.SetResult` or `Request.SetError` is called.
//  2. `Client.DisableAutoReadResponse` and `Request.DisableAutoReadResponse` is not
//     called, and also `Request.SetOutput` and `Request.SetOutputFile` is not called.
func (r *Response) Bytes() []byte {
	return r.body
}

// String returns the response body as string that hava already been read, could be
// nil if not read, the following cases are already read:
//  1. `Request.SetResult` or `Request.SetError` is called.
//  2. `Client.DisableAutoReadResponse` and `Request.DisableAutoReadResponse` is not
//     called, and also `Request.SetOutput` and `Request.SetOutputFile` is not called.
func (r *Response) String() string {
	return string(r.body)
}

// ToString returns the response body as string, read body if not have been read.
func (r *Response) ToString() (string, error) {
	b, err := r.ToBytes()
	return string(b), err
}

// ToBytes returns the response body as []byte, read body if not have been read.
func (r *Response) ToBytes() (body []byte, err error) {
	if r.Err != nil {
		return nil, r.Err
	}
	if r.body != nil {
		return r.body, nil
	}
	if r.Response == nil || r.Response.Body == nil {
		return []byte{}, nil
	}
	defer func() {
		r.Body.Close()
		if err != nil {
			r.Err = err
		}
		r.body = body
	}()
	body, err = io.ReadAll(r.Body)
	r.setReceivedAt()
	if err == nil && r.Request.client.responseBodyTransformer != nil {
		body, err = r.Request.client.responseBodyTransformer(body, r.Request, r)
	}
	return
}

// Dump return the string content that have been dumped for the request.
// `Request.Dump` or `Request.DumpXXX` MUST have been called.
func (r *Response) Dump() string {
	return r.Request.getDumpBuffer().String()
}

// GetStatus returns the response status.
func (r *Response) GetStatus() string {
	if r.Response == nil {
		return ""
	}
	return r.Status
}

// GetStatusCode returns the response status code.
func (r *Response) GetStatusCode() int {
	if r.Response == nil {
		return 0
	}
	return r.StatusCode
}

// GetHeader returns the response header value by key.
func (r *Response) GetHeader(key string) string {
	if r.Response == nil {
		return ""
	}
	return r.Header.Get(key)
}

// GetHeaderValues returns the response header values by key.
func (r *Response) GetHeaderValues(key string) []string {
	if r.Response == nil {
		return nil
	}
	return r.Header.Values(key)
}

// HeaderToString get all header as string.
func (r *Response) HeaderToString() string {
	if r.Response == nil {
		return ""
	}
	return convertHeaderToString(r.Header)
}
