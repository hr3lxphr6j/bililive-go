package req

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
)

type kv struct {
	Key   string
	Value string
}

// ContentDisposition represents parameters in `Content-Disposition`
// MIME header of multipart request.
type ContentDisposition struct {
	kv []kv
}

// Add adds a new key-value pair of Content-Disposition
func (c *ContentDisposition) Add(key, value string) *ContentDisposition {
	c.kv = append(c.kv, kv{Key: key, Value: value})
	return c
}

func (c *ContentDisposition) string() string {
	if c == nil {
		return ""
	}
	s := ""
	for _, kv := range c.kv {
		s += fmt.Sprintf("; %s=%q", kv.Key, kv.Value)
	}
	return s
}

// FileUpload represents a "form-data" multipart
type FileUpload struct {
	// "name" parameter in `Content-Disposition`
	ParamName string
	// "filename" parameter in `Content-Disposition`
	FileName string
	// The file to be uploaded.
	GetFileContent GetContentFunc
	// Optional file length in bytes.
	FileSize int64
	// Optional Content-Type
	ContentType string

	// Optional extra ContentDisposition parameters.
	// According to the HTTP specification, this should be nil,
	// but some servers may not follow the specification and
	// requires `Content-Disposition` parameters more than just
	// "name" and "filename".
	ExtraContentDisposition *ContentDisposition
}

// UploadInfo is the information for each UploadCallback call.
type UploadInfo struct {
	// parameter name in multipart upload
	ParamName string
	// filename in multipart upload
	FileName string
	// total file length in bytes.
	FileSize int64
	// uploaded file length in bytes.
	UploadedSize int64
}

// UploadCallback is the callback which will be invoked during
// multipart upload.
type UploadCallback func(info UploadInfo)

// DownloadInfo is the information for each DownloadCallback call.
type DownloadInfo struct {
	// Response is the corresponding Response during download.
	Response *Response
	// downloaded body length in bytes.
	DownloadedSize int64
}

// DownloadCallback is the callback which will be invoked during
// response body download.
type DownloadCallback func(info DownloadInfo)

func cloneCookies(cookies []*http.Cookie) []*http.Cookie {
	if len(cookies) == 0 {
		return nil
	}
	c := make([]*http.Cookie, len(cookies))
	copy(c, cookies)
	return c
}

func cloneHeaders(hdrs http.Header) http.Header {
	if hdrs == nil {
		return nil
	}
	h := make(http.Header)
	for k, vs := range hdrs {
		for _, v := range vs {
			h.Add(k, v)
		}
	}
	return h
}

// TODO: change to generics function when generics are commonly used.
func cloneRequestMiddleware(m []RequestMiddleware) []RequestMiddleware {
	if len(m) == 0 {
		return nil
	}
	mm := make([]RequestMiddleware, len(m))
	copy(mm, m)
	return mm
}

func cloneResponseMiddleware(m []ResponseMiddleware) []ResponseMiddleware {
	if len(m) == 0 {
		return nil
	}
	mm := make([]ResponseMiddleware, len(m))
	copy(mm, m)
	return mm
}

func cloneUrlValues(v url.Values) url.Values {
	if v == nil {
		return nil
	}
	vv := make(url.Values)
	for key, values := range v {
		for _, value := range values {
			vv.Add(key, value)
		}
	}
	return vv
}

func cloneMap(h map[string]string) map[string]string {
	if h == nil {
		return nil
	}
	m := make(map[string]string)
	for k, v := range h {
		m[k] = v
	}
	return m
}

// convertHeaderToString converts http header to a string.
func convertHeaderToString(h http.Header) string {
	if h == nil {
		return ""
	}
	buf := new(bytes.Buffer)
	h.Write(buf)
	return buf.String()
}
