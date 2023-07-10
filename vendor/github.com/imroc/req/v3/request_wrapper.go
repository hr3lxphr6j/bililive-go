package req

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"time"
)

// SetURL is a global wrapper methods which delegated
// to the default client, create a request and SetURL for request.
func SetURL(url string) *Request {
	return defaultClient.R().SetURL(url)
}

// SetFormDataFromValues is a global wrapper methods which delegated
// to the default client, create a request and SetFormDataFromValues for request.
func SetFormDataFromValues(data url.Values) *Request {
	return defaultClient.R().SetFormDataFromValues(data)
}

// SetFormData is a global wrapper methods which delegated
// to the default client, create a request and SetFormData for request.
func SetFormData(data map[string]string) *Request {
	return defaultClient.R().SetFormData(data)
}

// SetFormDataAnyType is a global wrapper methods which delegated
// to the default client, create a request and SetFormDataAnyType for request.
func SetFormDataAnyType(data map[string]interface{}) *Request {
	return defaultClient.R().SetFormDataAnyType(data)
}

// SetCookies is a global wrapper methods which delegated
// to the default client, create a request and SetCookies for request.
func SetCookies(cookies ...*http.Cookie) *Request {
	return defaultClient.R().SetCookies(cookies...)
}

// SetQueryString is a global wrapper methods which delegated
// to the default client, create a request and SetQueryString for request.
func SetQueryString(query string) *Request {
	return defaultClient.R().SetQueryString(query)
}

// SetFileReader is a global wrapper methods which delegated
// to the default client, create a request and SetFileReader for request.
func SetFileReader(paramName, filePath string, reader io.Reader) *Request {
	return defaultClient.R().SetFileReader(paramName, filePath, reader)
}

// SetFileBytes is a global wrapper methods which delegated
// to the default client, create a request and SetFileBytes for request.
func SetFileBytes(paramName, filename string, content []byte) *Request {
	return defaultClient.R().SetFileBytes(paramName, filename, content)
}

// SetFiles is a global wrapper methods which delegated
// to the default client, create a request and SetFiles for request.
func SetFiles(files map[string]string) *Request {
	return defaultClient.R().SetFiles(files)
}

// SetFile is a global wrapper methods which delegated
// to the default client, create a request and SetFile for request.
func SetFile(paramName, filePath string) *Request {
	return defaultClient.R().SetFile(paramName, filePath)
}

// SetFileUpload is a global wrapper methods which delegated
// to the default client, create a request and SetFileUpload for request.
func SetFileUpload(f ...FileUpload) *Request {
	return defaultClient.R().SetFileUpload(f...)
}

// SetResult is a global wrapper methods which delegated
// to the default client, create a request and SetSuccessResult for request.
//
// Deprecated: Use SetSuccessResult instead.
func SetResult(result interface{}) *Request {
	return defaultClient.R().SetSuccessResult(result)
}

// SetSuccessResult is a global wrapper methods which delegated
// to the default client, create a request and SetSuccessResult for request.
func SetSuccessResult(result interface{}) *Request {
	return defaultClient.R().SetSuccessResult(result)
}

// SetError is a global wrapper methods which delegated
// to the default client, create a request and SetErrorResult for request.
//
// Deprecated: Use SetErrorResult instead.
func SetError(error interface{}) *Request {
	return defaultClient.R().SetErrorResult(error)
}

// SetErrorResult is a global wrapper methods which delegated
// to the default client, create a request and SetErrorResult for request.
func SetErrorResult(error interface{}) *Request {
	return defaultClient.R().SetErrorResult(error)
}

// SetBearerAuthToken is a global wrapper methods which delegated
// to the default client, create a request and SetBearerAuthToken for request.
func SetBearerAuthToken(token string) *Request {
	return defaultClient.R().SetBearerAuthToken(token)
}

// SetBasicAuth is a global wrapper methods which delegated
// to the default client, create a request and SetBasicAuth for request.
func SetBasicAuth(username, password string) *Request {
	return defaultClient.R().SetBasicAuth(username, password)
}

// SetDigestAuth is a global wrapper methods which delegated
// to the default client, create a request and SetDigestAuth for request.
func SetDigestAuth(username, password string) *Request {
	return defaultClient.R().SetDigestAuth(username, password)
}

// SetHeaders is a global wrapper methods which delegated
// to the default client, create a request and SetHeaders for request.
func SetHeaders(hdrs map[string]string) *Request {
	return defaultClient.R().SetHeaders(hdrs)
}

// SetHeader is a global wrapper methods which delegated
// to the default client, create a request and SetHeader for request.
func SetHeader(key, value string) *Request {
	return defaultClient.R().SetHeader(key, value)
}

// SetOutputFile is a global wrapper methods which delegated
// to the default client, create a request and SetOutputFile for request.
func SetOutputFile(file string) *Request {
	return defaultClient.R().SetOutputFile(file)
}

// SetOutput is a global wrapper methods which delegated
// to the default client, create a request and SetOutput for request.
func SetOutput(output io.Writer) *Request {
	return defaultClient.R().SetOutput(output)
}

// SetQueryParams is a global wrapper methods which delegated
// to the default client, create a request and SetQueryParams for request.
func SetQueryParams(params map[string]string) *Request {
	return defaultClient.R().SetQueryParams(params)
}

// SetQueryParamsAnyType is a global wrapper methods which delegated
// to the default client, create a request and SetQueryParamsAnyType for request.
func SetQueryParamsAnyType(params map[string]interface{}) *Request {
	return defaultClient.R().SetQueryParamsAnyType(params)
}

// SetQueryParam is a global wrapper methods which delegated
// to the default client, create a request and SetQueryParam for request.
func SetQueryParam(key, value string) *Request {
	return defaultClient.R().SetQueryParam(key, value)
}

// AddQueryParam is a global wrapper methods which delegated
// to the default client, create a request and AddQueryParam for request.
func AddQueryParam(key, value string) *Request {
	return defaultClient.R().AddQueryParam(key, value)
}

// AddQueryParams is a global wrapper methods which delegated
// to the default client, create a request and AddQueryParams for request.
func AddQueryParams(key string, values ...string) *Request {
	return defaultClient.R().AddQueryParams(key, values...)
}

// SetPathParams is a global wrapper methods which delegated
// to the default client, create a request and SetPathParams for request.
func SetPathParams(params map[string]string) *Request {
	return defaultClient.R().SetPathParams(params)
}

// SetPathParam is a global wrapper methods which delegated
// to the default client, create a request and SetPathParam for request.
func SetPathParam(key, value string) *Request {
	return defaultClient.R().SetPathParam(key, value)
}

// MustGet is a global wrapper methods which delegated
// to the default client, create a request and MustGet for request.
func MustGet(url string) *Response {
	return defaultClient.R().MustGet(url)
}

// Get is a global wrapper methods which delegated
// to the default client, create a request and Get for request.
func Get(url string) (*Response, error) {
	return defaultClient.R().Get(url)
}

// MustPost is a global wrapper methods which delegated
// to the default client, create a request and Get for request.
func MustPost(url string) *Response {
	return defaultClient.R().MustPost(url)
}

// Post is a global wrapper methods which delegated
// to the default client, create a request and Post for request.
func Post(url string) (*Response, error) {
	return defaultClient.R().Post(url)
}

// MustPut is a global wrapper methods which delegated
// to the default client, create a request and MustPut for request.
func MustPut(url string) *Response {
	return defaultClient.R().MustPut(url)
}

// Put is a global wrapper methods which delegated
// to the default client, create a request and Put for request.
func Put(url string) (*Response, error) {
	return defaultClient.R().Put(url)
}

// MustPatch is a global wrapper methods which delegated
// to the default client, create a request and MustPatch for request.
func MustPatch(url string) *Response {
	return defaultClient.R().MustPatch(url)
}

// Patch is a global wrapper methods which delegated
// to the default client, create a request and Patch for request.
func Patch(url string) (*Response, error) {
	return defaultClient.R().Patch(url)
}

// MustDelete is a global wrapper methods which delegated
// to the default client, create a request and MustDelete for request.
func MustDelete(url string) *Response {
	return defaultClient.R().MustDelete(url)
}

// Delete is a global wrapper methods which delegated
// to the default client, create a request and Delete for request.
func Delete(url string) (*Response, error) {
	return defaultClient.R().Delete(url)
}

// MustOptions is a global wrapper methods which delegated
// to the default client, create a request and MustOptions for request.
func MustOptions(url string) *Response {
	return defaultClient.R().MustOptions(url)
}

// Options is a global wrapper methods which delegated
// to the default client, create a request and Options for request.
func Options(url string) (*Response, error) {
	return defaultClient.R().Options(url)
}

// MustHead is a global wrapper methods which delegated
// to the default client, create a request and MustHead for request.
func MustHead(url string) *Response {
	return defaultClient.R().MustHead(url)
}

// Head is a global wrapper methods which delegated
// to the default client, create a request and Head for request.
func Head(url string) (*Response, error) {
	return defaultClient.R().Head(url)
}

// SetBody is a global wrapper methods which delegated
// to the default client, create a request and SetBody for request.
func SetBody(body interface{}) *Request {
	return defaultClient.R().SetBody(body)
}

// SetBodyBytes is a global wrapper methods which delegated
// to the default client, create a request and SetBodyBytes for request.
func SetBodyBytes(body []byte) *Request {
	return defaultClient.R().SetBodyBytes(body)
}

// SetBodyString is a global wrapper methods which delegated
// to the default client, create a request and SetBodyString for request.
func SetBodyString(body string) *Request {
	return defaultClient.R().SetBodyString(body)
}

// SetBodyJsonString is a global wrapper methods which delegated
// to the default client, create a request and SetBodyJsonString for request.
func SetBodyJsonString(body string) *Request {
	return defaultClient.R().SetBodyJsonString(body)
}

// SetBodyJsonBytes is a global wrapper methods which delegated
// to the default client, create a request and SetBodyJsonBytes for request.
func SetBodyJsonBytes(body []byte) *Request {
	return defaultClient.R().SetBodyJsonBytes(body)
}

// SetBodyJsonMarshal is a global wrapper methods which delegated
// to the default client, create a request and SetBodyJsonMarshal for request.
func SetBodyJsonMarshal(v interface{}) *Request {
	return defaultClient.R().SetBodyJsonMarshal(v)
}

// SetBodyXmlString is a global wrapper methods which delegated
// to the default client, create a request and SetBodyXmlString for request.
func SetBodyXmlString(body string) *Request {
	return defaultClient.R().SetBodyXmlString(body)
}

// SetBodyXmlBytes is a global wrapper methods which delegated
// to the default client, create a request and SetBodyXmlBytes for request.
func SetBodyXmlBytes(body []byte) *Request {
	return defaultClient.R().SetBodyXmlBytes(body)
}

// SetBodyXmlMarshal is a global wrapper methods which delegated
// to the default client, create a request and SetBodyXmlMarshal for request.
func SetBodyXmlMarshal(v interface{}) *Request {
	return defaultClient.R().SetBodyXmlMarshal(v)
}

// SetContentType is a global wrapper methods which delegated
// to the default client, create a request and SetContentType for request.
func SetContentType(contentType string) *Request {
	return defaultClient.R().SetContentType(contentType)
}

// SetContext is a global wrapper methods which delegated
// to the default client, create a request and SetContext for request.
func SetContext(ctx context.Context) *Request {
	return defaultClient.R().SetContext(ctx)
}

// DisableTrace is a global wrapper methods which delegated
// to the default client, create a request and DisableTrace for request.
func DisableTrace() *Request {
	return defaultClient.R().DisableTrace()
}

// EnableTrace is a global wrapper methods which delegated
// to the default client, create a request and EnableTrace for request.
func EnableTrace() *Request {
	return defaultClient.R().EnableTrace()
}

// EnableForceChunkedEncoding is a global wrapper methods which delegated
// to the default client, create a request and EnableForceChunkedEncoding for request.
func EnableForceChunkedEncoding() *Request {
	return defaultClient.R().EnableForceChunkedEncoding()
}

// DisableForceChunkedEncoding is a global wrapper methods which delegated
// to the default client, create a request and DisableForceChunkedEncoding for request.
func DisableForceChunkedEncoding() *Request {
	return defaultClient.R().DisableForceChunkedEncoding()
}

// EnableForceMultipart is a global wrapper methods which delegated
// to the default client, create a request and EnableForceMultipart for request.
func EnableForceMultipart() *Request {
	return defaultClient.R().EnableForceMultipart()
}

// DisableForceMultipart is a global wrapper methods which delegated
// to the default client, create a request and DisableForceMultipart for request.
func DisableForceMultipart() *Request {
	return defaultClient.R().DisableForceMultipart()
}

// EnableDumpTo is a global wrapper methods which delegated
// to the default client, create a request and EnableDumpTo for request.
func EnableDumpTo(output io.Writer) *Request {
	return defaultClient.R().EnableDumpTo(output)
}

// EnableDumpToFile is a global wrapper methods which delegated
// to the default client, create a request and EnableDumpToFile for request.
func EnableDumpToFile(filename string) *Request {
	return defaultClient.R().EnableDumpToFile(filename)
}

// SetDumpOptions is a global wrapper methods which delegated
// to the default client, create a request and SetDumpOptions for request.
func SetDumpOptions(opt *DumpOptions) *Request {
	return defaultClient.R().SetDumpOptions(opt)
}

// EnableDump is a global wrapper methods which delegated
// to the default client, create a request and EnableDump for request.
func EnableDump() *Request {
	return defaultClient.R().EnableDump()
}

// EnableDumpWithoutBody is a global wrapper methods which delegated
// to the default client, create a request and EnableDumpWithoutBody for request.
func EnableDumpWithoutBody() *Request {
	return defaultClient.R().EnableDumpWithoutBody()
}

// EnableDumpWithoutHeader is a global wrapper methods which delegated
// to the default client, create a request and EnableDumpWithoutHeader for request.
func EnableDumpWithoutHeader() *Request {
	return defaultClient.R().EnableDumpWithoutHeader()
}

// EnableDumpWithoutResponse is a global wrapper methods which delegated
// to the default client, create a request and EnableDumpWithoutResponse for request.
func EnableDumpWithoutResponse() *Request {
	return defaultClient.R().EnableDumpWithoutResponse()
}

// EnableDumpWithoutRequest is a global wrapper methods which delegated
// to the default client, create a request and EnableDumpWithoutRequest for request.
func EnableDumpWithoutRequest() *Request {
	return defaultClient.R().EnableDumpWithoutRequest()
}

// EnableDumpWithoutRequestBody is a global wrapper methods which delegated
// to the default client, create a request and EnableDumpWithoutRequestBody for request.
func EnableDumpWithoutRequestBody() *Request {
	return defaultClient.R().EnableDumpWithoutRequestBody()
}

// EnableDumpWithoutResponseBody is a global wrapper methods which delegated
// to the default client, create a request and EnableDumpWithoutResponseBody for request.
func EnableDumpWithoutResponseBody() *Request {
	return defaultClient.R().EnableDumpWithoutResponseBody()
}

// SetRetryCount is a global wrapper methods which delegated
// to the default client, create a request and SetRetryCount for request.
func SetRetryCount(count int) *Request {
	return defaultClient.R().SetRetryCount(count)
}

// SetRetryInterval is a global wrapper methods which delegated
// to the default client, create a request and SetRetryInterval for request.
func SetRetryInterval(getRetryIntervalFunc GetRetryIntervalFunc) *Request {
	return defaultClient.R().SetRetryInterval(getRetryIntervalFunc)
}

// SetRetryFixedInterval is a global wrapper methods which delegated
// to the default client, create a request and SetRetryFixedInterval for request.
func SetRetryFixedInterval(interval time.Duration) *Request {
	return defaultClient.R().SetRetryFixedInterval(interval)
}

// SetRetryBackoffInterval is a global wrapper methods which delegated
// to the default client, create a request and SetRetryBackoffInterval for request.
func SetRetryBackoffInterval(min, max time.Duration) *Request {
	return defaultClient.R().SetRetryBackoffInterval(min, max)
}

// SetRetryHook is a global wrapper methods which delegated
// to the default client, create a request and SetRetryHook for request.
func SetRetryHook(hook RetryHookFunc) *Request {
	return defaultClient.R().SetRetryHook(hook)
}

// AddRetryHook is a global wrapper methods which delegated
// to the default client, create a request and AddRetryHook for request.
func AddRetryHook(hook RetryHookFunc) *Request {
	return defaultClient.R().AddRetryHook(hook)
}

// SetRetryCondition is a global wrapper methods which delegated
// to the default client, create a request and SetRetryCondition for request.
func SetRetryCondition(condition RetryConditionFunc) *Request {
	return defaultClient.R().SetRetryCondition(condition)
}

// AddRetryCondition is a global wrapper methods which delegated
// to the default client, create a request and AddRetryCondition for request.
func AddRetryCondition(condition RetryConditionFunc) *Request {
	return defaultClient.R().AddRetryCondition(condition)
}

// SetUploadCallback is a global wrapper methods which delegated
// to the default client, create a request and SetUploadCallback for request.
func SetUploadCallback(callback UploadCallback) *Request {
	return defaultClient.R().SetUploadCallback(callback)
}

// SetUploadCallbackWithInterval is a global wrapper methods which delegated
// to the default client, create a request and SetUploadCallbackWithInterval for request.
func SetUploadCallbackWithInterval(callback UploadCallback, minInterval time.Duration) *Request {
	return defaultClient.R().SetUploadCallbackWithInterval(callback, minInterval)
}

// SetDownloadCallback is a global wrapper methods which delegated
// to the default client, create a request and SetDownloadCallback for request.
func SetDownloadCallback(callback DownloadCallback) *Request {
	return defaultClient.R().SetDownloadCallback(callback)
}

// SetDownloadCallbackWithInterval is a global wrapper methods which delegated
// to the default client, create a request and SetDownloadCallbackWithInterval for request.
func SetDownloadCallbackWithInterval(callback DownloadCallback, minInterval time.Duration) *Request {
	return defaultClient.R().SetDownloadCallbackWithInterval(callback, minInterval)
}

// EnableCloseConnection is a global wrapper methods which delegated
// to the default client, create a request and EnableCloseConnection for request.
func EnableCloseConnection() *Request {
	return defaultClient.R().EnableCloseConnection()
}
