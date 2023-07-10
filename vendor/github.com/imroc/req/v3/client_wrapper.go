package req

import (
	"context"
	"crypto/tls"
	utls "github.com/refraction-networking/utls"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

// WrapRoundTrip is a global wrapper methods which delegated
// to the default client's Client.WrapRoundTrip.
func WrapRoundTrip(wrappers ...RoundTripWrapper) *Client {
	return defaultClient.WrapRoundTrip(wrappers...)
}

// WrapRoundTripFunc is a global wrapper methods which delegated
// to the default client's Client.WrapRoundTripFunc.
func WrapRoundTripFunc(funcs ...RoundTripWrapperFunc) *Client {
	return defaultClient.WrapRoundTripFunc(funcs...)
}

// SetCommonError is a global wrapper methods which delegated
// to the default client's Client.SetCommonErrorResult.
//
// Deprecated: Use SetCommonErrorResult instead.
func SetCommonError(err interface{}) *Client {
	return defaultClient.SetCommonErrorResult(err)
}

// SetCommonErrorResult is a global wrapper methods which delegated
// to the default client's Client.SetCommonError.
func SetCommonErrorResult(err interface{}) *Client {
	return defaultClient.SetCommonErrorResult(err)
}

// SetResultStateCheckFunc is a global wrapper methods which delegated
// to the default client's Client.SetCommonResultStateCheckFunc.
func SetResultStateCheckFunc(fn func(resp *Response) ResultState) *Client {
	return defaultClient.SetResultStateCheckFunc(fn)
}

// SetCommonFormDataFromValues is a global wrapper methods which delegated
// to the default client's Client.SetCommonFormDataFromValues.
func SetCommonFormDataFromValues(data url.Values) *Client {
	return defaultClient.SetCommonFormDataFromValues(data)
}

// SetCommonFormData is a global wrapper methods which delegated
// to the default client's Client.SetCommonFormData.
func SetCommonFormData(data map[string]string) *Client {
	return defaultClient.SetCommonFormData(data)
}

// SetBaseURL is a global wrapper methods which delegated
// to the default client's Client.SetBaseURL.
func SetBaseURL(u string) *Client {
	return defaultClient.SetBaseURL(u)
}

// SetOutputDirectory is a global wrapper methods which delegated
// to the default client's Client.SetOutputDirectory.
func SetOutputDirectory(dir string) *Client {
	return defaultClient.SetOutputDirectory(dir)
}

// SetCertFromFile is a global wrapper methods which delegated
// to the default client's Client.SetCertFromFile.
func SetCertFromFile(certFile, keyFile string) *Client {
	return defaultClient.SetCertFromFile(certFile, keyFile)
}

// SetCerts is a global wrapper methods which delegated
// to the default client's Client.SetCerts.
func SetCerts(certs ...tls.Certificate) *Client {
	return defaultClient.SetCerts(certs...)
}

// SetRootCertFromString is a global wrapper methods which delegated
// to the default client's Client.SetRootCertFromString.
func SetRootCertFromString(pemContent string) *Client {
	return defaultClient.SetRootCertFromString(pemContent)
}

// SetRootCertsFromFile is a global wrapper methods which delegated
// to the default client's Client.SetRootCertsFromFile.
func SetRootCertsFromFile(pemFiles ...string) *Client {
	return defaultClient.SetRootCertsFromFile(pemFiles...)
}

// GetTLSClientConfig is a global wrapper methods which delegated
// to the default client's Client.GetTLSClientConfig.
func GetTLSClientConfig() *tls.Config {
	return defaultClient.GetTLSClientConfig()
}

// SetRedirectPolicy is a global wrapper methods which delegated
// to the default client's Client.SetRedirectPolicy.
func SetRedirectPolicy(policies ...RedirectPolicy) *Client {
	return defaultClient.SetRedirectPolicy(policies...)
}

// DisableKeepAlives is a global wrapper methods which delegated
// to the default client's Client.DisableKeepAlives.
func DisableKeepAlives() *Client {
	return defaultClient.DisableKeepAlives()
}

// EnableKeepAlives is a global wrapper methods which delegated
// to the default client's Client.EnableKeepAlives.
func EnableKeepAlives() *Client {
	return defaultClient.EnableKeepAlives()
}

// DisableCompression is a global wrapper methods which delegated
// to the default client's Client.DisableCompression.
func DisableCompression() *Client {
	return defaultClient.DisableCompression()
}

// EnableCompression is a global wrapper methods which delegated
// to the default client's Client.EnableCompression.
func EnableCompression() *Client {
	return defaultClient.EnableCompression()
}

// SetTLSClientConfig is a global wrapper methods which delegated
// to the default client's Client.SetTLSClientConfig.
func SetTLSClientConfig(conf *tls.Config) *Client {
	return defaultClient.SetTLSClientConfig(conf)
}

// EnableInsecureSkipVerify is a global wrapper methods which delegated
// to the default client's Client.EnableInsecureSkipVerify.
func EnableInsecureSkipVerify() *Client {
	return defaultClient.EnableInsecureSkipVerify()
}

// DisableInsecureSkipVerify is a global wrapper methods which delegated
// to the default client's Client.DisableInsecureSkipVerify.
func DisableInsecureSkipVerify() *Client {
	return defaultClient.DisableInsecureSkipVerify()
}

// SetCommonQueryParams is a global wrapper methods which delegated
// to the default client's Client.SetCommonQueryParams.
func SetCommonQueryParams(params map[string]string) *Client {
	return defaultClient.SetCommonQueryParams(params)
}

// AddCommonQueryParam is a global wrapper methods which delegated
// to the default client's Client.AddCommonQueryParam.
func AddCommonQueryParam(key, value string) *Client {
	return defaultClient.AddCommonQueryParam(key, value)
}

// AddCommonQueryParams is a global wrapper methods which delegated
// to the default client's Client.AddCommonQueryParams.
func AddCommonQueryParams(key string, values ...string) *Client {
	return defaultClient.AddCommonQueryParams(key, values...)
}

// SetCommonPathParam is a global wrapper methods which delegated
// to the default client's Client.SetCommonPathParam.
func SetCommonPathParam(key, value string) *Client {
	return defaultClient.SetCommonPathParam(key, value)
}

// SetCommonPathParams is a global wrapper methods which delegated
// to the default client's Client.SetCommonPathParams.
func SetCommonPathParams(pathParams map[string]string) *Client {
	return defaultClient.SetCommonPathParams(pathParams)
}

// SetCommonQueryParam is a global wrapper methods which delegated
// to the default client's Client.SetCommonQueryParam.
func SetCommonQueryParam(key, value string) *Client {
	return defaultClient.SetCommonQueryParam(key, value)
}

// SetCommonQueryString is a global wrapper methods which delegated
// to the default client's Client.SetCommonQueryString.
func SetCommonQueryString(query string) *Client {
	return defaultClient.SetCommonQueryString(query)
}

// SetCommonCookies is a global wrapper methods which delegated
// to the default client's Client.SetCommonCookies.
func SetCommonCookies(cookies ...*http.Cookie) *Client {
	return defaultClient.SetCommonCookies(cookies...)
}

// DisableDebugLog is a global wrapper methods which delegated
// to the default client's Client.DisableDebugLog.
func DisableDebugLog() *Client {
	return defaultClient.DisableDebugLog()
}

// EnableDebugLog is a global wrapper methods which delegated
// to the default client's Client.EnableDebugLog.
func EnableDebugLog() *Client {
	return defaultClient.EnableDebugLog()
}

// DevMode is a global wrapper methods which delegated
// to the default client's Client.DevMode.
func DevMode() *Client {
	return defaultClient.DevMode()
}

// SetScheme is a global wrapper methods which delegated
// to the default client's Client.SetScheme.
func SetScheme(scheme string) *Client {
	return defaultClient.SetScheme(scheme)
}

// SetLogger is a global wrapper methods which delegated
// to the default client's Client.SetLogger.
func SetLogger(log Logger) *Client {
	return defaultClient.SetLogger(log)
}

// SetTimeout is a global wrapper methods which delegated
// to the default client's Client.SetTimeout.
func SetTimeout(d time.Duration) *Client {
	return defaultClient.SetTimeout(d)
}

// EnableDumpAll is a global wrapper methods which delegated
// to the default client's Client.EnableDumpAll.
func EnableDumpAll() *Client {
	return defaultClient.EnableDumpAll()
}

// EnableDumpAllToFile is a global wrapper methods which delegated
// to the default client's Client.EnableDumpAllToFile.
func EnableDumpAllToFile(filename string) *Client {
	return defaultClient.EnableDumpAllToFile(filename)
}

// EnableDumpAllTo is a global wrapper methods which delegated
// to the default client's Client.EnableDumpAllTo.
func EnableDumpAllTo(output io.Writer) *Client {
	return defaultClient.EnableDumpAllTo(output)
}

// EnableDumpAllAsync is a global wrapper methods which delegated
// to the default client's Client.EnableDumpAllAsync.
func EnableDumpAllAsync() *Client {
	return defaultClient.EnableDumpAllAsync()
}

// EnableDumpAllWithoutRequestBody is a global wrapper methods which delegated
// to the default client's Client.EnableDumpAllWithoutRequestBody.
func EnableDumpAllWithoutRequestBody() *Client {
	return defaultClient.EnableDumpAllWithoutRequestBody()
}

// EnableDumpAllWithoutResponseBody is a global wrapper methods which delegated
// to the default client's Client.EnableDumpAllWithoutResponseBody.
func EnableDumpAllWithoutResponseBody() *Client {
	return defaultClient.EnableDumpAllWithoutResponseBody()
}

// EnableDumpAllWithoutResponse is a global wrapper methods which delegated
// to the default client's Client.EnableDumpAllWithoutResponse.
func EnableDumpAllWithoutResponse() *Client {
	return defaultClient.EnableDumpAllWithoutResponse()
}

// EnableDumpAllWithoutRequest is a global wrapper methods which delegated
// to the default client's Client.EnableDumpAllWithoutRequest.
func EnableDumpAllWithoutRequest() *Client {
	return defaultClient.EnableDumpAllWithoutRequest()
}

// EnableDumpAllWithoutHeader is a global wrapper methods which delegated
// to the default client's Client.EnableDumpAllWithoutHeader.
func EnableDumpAllWithoutHeader() *Client {
	return defaultClient.EnableDumpAllWithoutHeader()
}

// EnableDumpAllWithoutBody is a global wrapper methods which delegated
// to the default client's Client.EnableDumpAllWithoutBody.
func EnableDumpAllWithoutBody() *Client {
	return defaultClient.EnableDumpAllWithoutBody()
}

// EnableDumpEachRequest is a global wrapper methods which delegated
// to the default client's Client.EnableDumpEachRequest.
func EnableDumpEachRequest() *Client {
	return defaultClient.EnableDumpEachRequest()
}

// EnableDumpEachRequestWithoutBody is a global wrapper methods which delegated
// to the default client's Client.EnableDumpEachRequestWithoutBody.
func EnableDumpEachRequestWithoutBody() *Client {
	return defaultClient.EnableDumpEachRequestWithoutBody()
}

// EnableDumpEachRequestWithoutHeader is a global wrapper methods which delegated
// to the default client's Client.EnableDumpEachRequestWithoutHeader.
func EnableDumpEachRequestWithoutHeader() *Client {
	return defaultClient.EnableDumpEachRequestWithoutHeader()
}

// EnableDumpEachRequestWithoutResponse is a global wrapper methods which delegated
// to the default client's Client.EnableDumpEachRequestWithoutResponse.
func EnableDumpEachRequestWithoutResponse() *Client {
	return defaultClient.EnableDumpEachRequestWithoutResponse()
}

// EnableDumpEachRequestWithoutRequest is a global wrapper methods which delegated
// to the default client's Client.EnableDumpEachRequestWithoutRequest.
func EnableDumpEachRequestWithoutRequest() *Client {
	return defaultClient.EnableDumpEachRequestWithoutRequest()
}

// EnableDumpEachRequestWithoutResponseBody is a global wrapper methods which delegated
// to the default client's Client.EnableDumpEachRequestWithoutResponseBody.
func EnableDumpEachRequestWithoutResponseBody() *Client {
	return defaultClient.EnableDumpEachRequestWithoutResponseBody()
}

// EnableDumpEachRequestWithoutRequestBody is a global wrapper methods which delegated
// to the default client's Client.EnableDumpEachRequestWithoutRequestBody.
func EnableDumpEachRequestWithoutRequestBody() *Client {
	return defaultClient.EnableDumpEachRequestWithoutRequestBody()
}

// DisableAutoReadResponse is a global wrapper methods which delegated
// to the default client's Client.DisableAutoReadResponse.
func DisableAutoReadResponse() *Client {
	return defaultClient.DisableAutoReadResponse()
}

// EnableAutoReadResponse is a global wrapper methods which delegated
// to the default client's Client.EnableAutoReadResponse.
func EnableAutoReadResponse() *Client {
	return defaultClient.EnableAutoReadResponse()
}

// SetAutoDecodeContentType is a global wrapper methods which delegated
// to the default client's Client.SetAutoDecodeContentType.
func SetAutoDecodeContentType(contentTypes ...string) *Client {
	return defaultClient.SetAutoDecodeContentType(contentTypes...)
}

// SetAutoDecodeContentTypeFunc is a global wrapper methods which delegated
// to the default client's Client.SetAutoDecodeAllTypeFunc.
func SetAutoDecodeContentTypeFunc(fn func(contentType string) bool) *Client {
	return defaultClient.SetAutoDecodeContentTypeFunc(fn)
}

// SetAutoDecodeAllContentType is a global wrapper methods which delegated
// to the default client's Client.SetAutoDecodeAllContentType.
func SetAutoDecodeAllContentType() *Client {
	return defaultClient.SetAutoDecodeAllContentType()
}

// DisableAutoDecode is a global wrapper methods which delegated
// to the default client's Client.DisableAutoDecode.
func DisableAutoDecode() *Client {
	return defaultClient.DisableAutoDecode()
}

// EnableAutoDecode is a global wrapper methods which delegated
// to the default client's Client.EnableAutoDecode.
func EnableAutoDecode() *Client {
	return defaultClient.EnableAutoDecode()
}

// SetUserAgent is a global wrapper methods which delegated
// to the default client's Client.SetUserAgent.
func SetUserAgent(userAgent string) *Client {
	return defaultClient.SetUserAgent(userAgent)
}

// SetCommonBearerAuthToken is a global wrapper methods which delegated
// to the default client's Client.SetCommonBearerAuthToken.
func SetCommonBearerAuthToken(token string) *Client {
	return defaultClient.SetCommonBearerAuthToken(token)
}

// SetCommonBasicAuth is a global wrapper methods which delegated
// to the default client's Client.SetCommonBasicAuth.
func SetCommonBasicAuth(username, password string) *Client {
	return defaultClient.SetCommonBasicAuth(username, password)
}

// SetCommonDigestAuth is a global wrapper methods which delegated
// to the default client's Client.SetCommonDigestAuth.
func SetCommonDigestAuth(username, password string) *Client {
	return defaultClient.SetCommonDigestAuth(username, password)
}

// SetCommonHeaders is a global wrapper methods which delegated
// to the default client's Client.SetCommonHeaders.
func SetCommonHeaders(hdrs map[string]string) *Client {
	return defaultClient.SetCommonHeaders(hdrs)
}

// SetCommonHeader is a global wrapper methods which delegated
// to the default client's Client.SetCommonHeader.
func SetCommonHeader(key, value string) *Client {
	return defaultClient.SetCommonHeader(key, value)
}

// SetCommonContentType is a global wrapper methods which delegated
// to the default client's Client.SetCommonContentType.
func SetCommonContentType(ct string) *Client {
	return defaultClient.SetCommonContentType(ct)
}

// DisableDumpAll is a global wrapper methods which delegated
// to the default client's Client.DisableDumpAll.
func DisableDumpAll() *Client {
	return defaultClient.DisableDumpAll()
}

// SetCommonDumpOptions is a global wrapper methods which delegated
// to the default client's Client.SetCommonDumpOptions.
func SetCommonDumpOptions(opt *DumpOptions) *Client {
	return defaultClient.SetCommonDumpOptions(opt)
}

// SetProxy is a global wrapper methods which delegated
// to the default client's Client.SetProxy.
func SetProxy(proxy func(*http.Request) (*url.URL, error)) *Client {
	return defaultClient.SetProxy(proxy)
}

// OnBeforeRequest is a global wrapper methods which delegated
// to the default client's Client.OnBeforeRequest.
func OnBeforeRequest(m RequestMiddleware) *Client {
	return defaultClient.OnBeforeRequest(m)
}

// OnAfterResponse is a global wrapper methods which delegated
// to the default client's Client.OnAfterResponse.
func OnAfterResponse(m ResponseMiddleware) *Client {
	return defaultClient.OnAfterResponse(m)
}

// SetProxyURL is a global wrapper methods which delegated
// to the default client's Client.SetProxyURL.
func SetProxyURL(proxyUrl string) *Client {
	return defaultClient.SetProxyURL(proxyUrl)
}

// DisableTraceAll is a global wrapper methods which delegated
// to the default client's Client.DisableTraceAll.
func DisableTraceAll() *Client {
	return defaultClient.DisableTraceAll()
}

// EnableTraceAll is a global wrapper methods which delegated
// to the default client's Client.EnableTraceAll.
func EnableTraceAll() *Client {
	return defaultClient.EnableTraceAll()
}

// SetCookieJar is a global wrapper methods which delegated
// to the default client's Client.SetCookieJar.
func SetCookieJar(jar http.CookieJar) *Client {
	return defaultClient.SetCookieJar(jar)
}

// GetCookies is a global wrapper methods which delegated
// to the default client's Client.GetCookies.
func GetCookies(url string) ([]*http.Cookie, error) {
	return defaultClient.GetCookies(url)
}

// ClearCookies is a global wrapper methods which delegated
// to the default client's Client.ClearCookies.
func ClearCookies() *Client {
	return defaultClient.ClearCookies()
}

// SetJsonMarshal is a global wrapper methods which delegated
// to the default client's Client.SetJsonMarshal.
func SetJsonMarshal(fn func(v interface{}) ([]byte, error)) *Client {
	return defaultClient.SetJsonMarshal(fn)
}

// SetJsonUnmarshal is a global wrapper methods which delegated
// to the default client's Client.SetJsonUnmarshal.
func SetJsonUnmarshal(fn func(data []byte, v interface{}) error) *Client {
	return defaultClient.SetJsonUnmarshal(fn)
}

// SetXmlMarshal is a global wrapper methods which delegated
// to the default client's Client.SetXmlMarshal.
func SetXmlMarshal(fn func(v interface{}) ([]byte, error)) *Client {
	return defaultClient.SetXmlMarshal(fn)
}

// SetXmlUnmarshal is a global wrapper methods which delegated
// to the default client's Client.SetXmlUnmarshal.
func SetXmlUnmarshal(fn func(data []byte, v interface{}) error) *Client {
	return defaultClient.SetXmlUnmarshal(fn)
}

// SetDialTLS is a global wrapper methods which delegated
// to the default client's Client.SetDialTLS.
func SetDialTLS(fn func(ctx context.Context, network, addr string) (net.Conn, error)) *Client {
	return defaultClient.SetDialTLS(fn)
}

// SetDial is a global wrapper methods which delegated
// to the default client's Client.SetDial.
func SetDial(fn func(ctx context.Context, network, addr string) (net.Conn, error)) *Client {
	return defaultClient.SetDial(fn)
}

// SetTLSHandshakeTimeout is a global wrapper methods which delegated
// to the default client's Client.SetTLSHandshakeTimeout.
func SetTLSHandshakeTimeout(timeout time.Duration) *Client {
	return defaultClient.SetTLSHandshakeTimeout(timeout)
}

// EnableForceHTTP1 is a global wrapper methods which delegated
// to the default client's Client.EnableForceHTTP1.
func EnableForceHTTP1() *Client {
	return defaultClient.EnableForceHTTP1()
}

// EnableForceHTTP2 is a global wrapper methods which delegated
// to the default client's Client.EnableForceHTTP2.
func EnableForceHTTP2() *Client {
	return defaultClient.EnableForceHTTP2()
}

// EnableForceHTTP3 is a global wrapper methods which delegated
// to the default client's Client.EnableForceHTTP3.
func EnableForceHTTP3() *Client {
	return defaultClient.EnableForceHTTP3()
}

// EnableHTTP3 is a global wrapper methods which delegated
// to the default client's Client.EnableHTTP3.
func EnableHTTP3() *Client {
	return defaultClient.EnableHTTP3()
}

// DisableForceHttpVersion is a global wrapper methods which delegated
// to the default client's Client.DisableForceHttpVersion.
func DisableForceHttpVersion() *Client {
	return defaultClient.DisableForceHttpVersion()
}

// EnableH2C is a global wrapper methods which delegated
// to the default client's Client.EnableH2C.
func EnableH2C() *Client {
	return defaultClient.EnableH2C()
}

// DisableH2C is a global wrapper methods which delegated
// to the default client's Client.DisableH2C.
func DisableH2C() *Client {
	return defaultClient.DisableH2C()
}

// DisableAllowGetMethodPayload is a global wrapper methods which delegated
// to the default client's Client.DisableAllowGetMethodPayload.
func DisableAllowGetMethodPayload() *Client {
	return defaultClient.DisableAllowGetMethodPayload()
}

// EnableAllowGetMethodPayload is a global wrapper methods which delegated
// to the default client's Client.EnableAllowGetMethodPayload.
func EnableAllowGetMethodPayload() *Client {
	return defaultClient.EnableAllowGetMethodPayload()
}

// SetCommonRetryCount is a global wrapper methods which delegated
// to the default client's Client.SetCommonRetryCount.
func SetCommonRetryCount(count int) *Client {
	return defaultClient.SetCommonRetryCount(count)
}

// SetCommonRetryInterval is a global wrapper methods which delegated
// to the default client's Client.SetCommonRetryInterval.
func SetCommonRetryInterval(getRetryIntervalFunc GetRetryIntervalFunc) *Client {
	return defaultClient.SetCommonRetryInterval(getRetryIntervalFunc)
}

// SetCommonRetryFixedInterval is a global wrapper methods which delegated
// to the default client's Client.SetCommonRetryFixedInterval.
func SetCommonRetryFixedInterval(interval time.Duration) *Client {
	return defaultClient.SetCommonRetryFixedInterval(interval)
}

// SetCommonRetryBackoffInterval is a global wrapper methods which delegated
// to the default client's Client.SetCommonRetryBackoffInterval.
func SetCommonRetryBackoffInterval(min, max time.Duration) *Client {
	return defaultClient.SetCommonRetryBackoffInterval(min, max)
}

// SetCommonRetryHook is a global wrapper methods which delegated
// to the default client's Client.SetCommonRetryHook.
func SetCommonRetryHook(hook RetryHookFunc) *Client {
	return defaultClient.SetCommonRetryHook(hook)
}

// AddCommonRetryHook is a global wrapper methods which delegated
// to the default client's Client.AddCommonRetryHook.
func AddCommonRetryHook(hook RetryHookFunc) *Client {
	return defaultClient.AddCommonRetryHook(hook)
}

// SetCommonRetryCondition is a global wrapper methods which delegated
// to the default client's Client.SetCommonRetryCondition.
func SetCommonRetryCondition(condition RetryConditionFunc) *Client {
	return defaultClient.SetCommonRetryCondition(condition)
}

// AddCommonRetryCondition is a global wrapper methods which delegated
// to the default client's Client.AddCommonRetryCondition.
func AddCommonRetryCondition(condition RetryConditionFunc) *Client {
	return defaultClient.AddCommonRetryCondition(condition)
}

// SetResponseBodyTransformer is a global wrapper methods which delegated
// to the default client's Client.SetResponseBodyTransformer.
func SetResponseBodyTransformer(fn func(rawBody []byte, req *Request, resp *Response) (transformedBody []byte, err error)) *Client {
	return defaultClient.SetResponseBodyTransformer(fn)
}

// SetUnixSocket is a global wrapper methods which delegated
// to the default client's Client.SetUnixSocket.
func SetUnixSocket(file string) *Client {
	return defaultClient.SetUnixSocket(file)
}

// SetTLSFingerprint is a global wrapper methods which delegated
// to the default client's Client.SetTLSFingerprint.
func SetTLSFingerprint(clientHelloID utls.ClientHelloID) *Client {
	return defaultClient.SetTLSFingerprint(clientHelloID)
}

// SetTLSFingerprintRandomized is a global wrapper methods which delegated
// to the default client's Client.SetTLSFingerprintRandomized.
func SetTLSFingerprintRandomized() *Client {
	return defaultClient.SetTLSFingerprintRandomized()
}

// SetTLSFingerprintChrome is a global wrapper methods which delegated
// to the default client's Client.SetTLSFingerprintChrome.
func SetTLSFingerprintChrome() *Client {
	return defaultClient.SetTLSFingerprintChrome()
}

// SetTLSFingerprintAndroid is a global wrapper methods which delegated
// to the default client's Client.SetTLSFingerprintAndroid.
func SetTLSFingerprintAndroid() *Client {
	return defaultClient.SetTLSFingerprintAndroid()
}

// SetTLSFingerprint360 is a global wrapper methods which delegated
// to the default client's Client.SetTLSFingerprint360.
func SetTLSFingerprint360() *Client {
	return defaultClient.SetTLSFingerprint360()
}

// SetTLSFingerprintEdge is a global wrapper methods which delegated
// to the default client's Client.SetTLSFingerprintEdge.
func SetTLSFingerprintEdge() *Client {
	return defaultClient.SetTLSFingerprintEdge()
}

// SetTLSFingerprintFirefox is a global wrapper methods which delegated
// to the default client's Client.SetTLSFingerprintFirefox.
func SetTLSFingerprintFirefox() *Client {
	return defaultClient.SetTLSFingerprintFirefox()
}

// SetTLSFingerprintQQ is a global wrapper methods which delegated
// to the default client's Client.SetTLSFingerprintQQ.
func SetTLSFingerprintQQ() *Client {
	return defaultClient.SetTLSFingerprintQQ()
}

// SetTLSFingerprintIOS is a global wrapper methods which delegated
// to the default client's Client.SetTLSFingerprintIOS.
func SetTLSFingerprintIOS() *Client {
	return defaultClient.SetTLSFingerprintIOS()
}

// SetTLSFingerprintSafari is a global wrapper methods which delegated
// to the default client's Client.SetTLSFingerprintSafari.
func SetTLSFingerprintSafari() *Client {
	return defaultClient.SetTLSFingerprintSafari()
}

// GetClient is a global wrapper methods which delegated
// to the default client's Client.GetClient.
func GetClient() *http.Client {
	return defaultClient.GetClient()
}

// NewRequest is a global wrapper methods which delegated
// to the default client's Client.NewRequest.
func NewRequest() *Request {
	return defaultClient.R()
}

// R is a global wrapper methods which delegated
// to the default client's Client.R().
func R() *Request {
	return defaultClient.R()
}
