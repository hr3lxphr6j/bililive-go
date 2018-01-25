package request

import "net/http"

// DefaultUserAgent define default User-Agent header
var DefaultUserAgent = "go-request/" + Version

// DefaultHeaders define default headers
var DefaultHeaders = map[string]string{
	"Connection":      "keep-alive",
	"Accept-Encoding": "gzip, deflate",
	"Accept":          "*/*",
	"User-Agent":      DefaultUserAgent,
}

// DefaultContentType define default Content-Type Header for form body
var DefaultContentType = "application/x-www-form-urlencoded; charset=utf-8"

// DefaultJsonType define default Content-Type Header for json body
var DefaultJsonType = "application/json; charset=utf-8"

func applyHeaders(a *Args, req *http.Request, contentType string) {
	// apply defaultHeaders
	for k, v := range DefaultHeaders {
		_, ok := a.Headers[k]
		if !ok {
			req.Header.Set(k, v)
		}
	}
	// apply custom Headers
	for k, v := range a.Headers {
		req.Header.Set(k, v)
	}
	// apply "Content-Type" Headers
	_, ok := a.Headers["Content-Type"]
	if !ok {
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}
	}
}
