package req

import (
	"errors"
	"github.com/imroc/req/v3/internal/ascii"
	"golang.org/x/net/http/httpguts"
	"net/http"
	"strings"
)

// Given a string of the form "host", "host:port", or "[ipv6::address]:port",
// return true if the string includes a port.
func hasPort(s string) bool { return strings.LastIndex(s, ":") > strings.LastIndex(s, "]") }

// removeEmptyPort strips the empty port in ":port" to ""
// as mandated by RFC 3986 Section 6.2.3.
func removeEmptyPort(host string) string {
	if hasPort(host) {
		return strings.TrimSuffix(host, ":")
	}
	return host
}

func isNotToken(r rune) bool {
	return !httpguts.IsTokenRune(r)
}

func validMethod(method string) bool {
	/*
	     Method         = "OPTIONS"                ; Section 9.2
	                    | "GET"                    ; Section 9.3
	                    | "HEAD"                   ; Section 9.4
	                    | "POST"                   ; Section 9.5
	                    | "PUT"                    ; Section 9.6
	                    | "DELETE"                 ; Section 9.7
	                    | "TRACE"                  ; Section 9.8
	                    | "CONNECT"                ; Section 9.9
	                    | extension-method
	   extension-method = token
	     token          = 1*<any CHAR except CTLs or separators>
	*/
	return len(method) > 0 && strings.IndexFunc(method, isNotToken) == -1
}

func closeBody(r *http.Request) error {
	if r.Body == nil {
		return nil
	}
	return r.Body.Close()
}

// requestBodyReadError wraps an error from (*Request).write to indicate
// that the error came from a Read call on the Request.Body.
// This error type should not escape the net/http package to users.
type requestBodyReadError struct{ error }

// Return value if nonempty, def otherwise.
func valueOrDefault(value, def string) string {
	if value != "" {
		return value
	}
	return def
}

// outgoingLength reports the Content-Length of this outgoing (Client) request.
// It maps 0 into -1 (unknown) when the Body is non-nil.
func outgoingLength(r *http.Request) int64 {
	if r.Body == nil || r.Body == NoBody {
		return 0
	}
	if r.ContentLength != 0 {
		return r.ContentLength
	}
	return -1
}

// errMissingHost is returned by Write when there is no Host or URL present in
// the Request.
var errMissingHost = errors.New("http: Request.Write on Request with no Host or URL set")

func closeRequestBody(r *http.Request) error {
	if r.Body == nil {
		return nil
	}
	return r.Body.Close()
}

// Headers that Request.Write handles itself and should be skipped.
var reqWriteExcludeHeader = map[string]bool{
	"Host":              true, // not in Header map anyway
	"User-Agent":        true,
	"Content-Length":    true,
	"Transfer-Encoding": true,
	"Trailer":           true,
}

// requestMethodUsuallyLacksBody reports whether the given request
// method is one that typically does not involve a request body.
// This is used by the Transport (via
// transferWriter.shouldSendChunkedRequestBody) to determine whether
// we try to test-read a byte from a non-nil Request.Body when
// Request.outgoingLength() returns -1. See the comments in
// shouldSendChunkedRequestBody.
func requestMethodUsuallyLacksBody(method string) bool {
	switch method {
	case "GET", "HEAD", "DELETE", "OPTIONS", "PROPFIND", "SEARCH":
		return true
	}
	return false
}

// requiresHTTP1 reports whether this request requires being sent on
// an HTTP/1 connection.
func requestRequiresHTTP1(r *http.Request) bool {
	return hasToken(r.Header.Get("Connection"), "upgrade") &&
		ascii.EqualFold(r.Header.Get("Upgrade"), "websocket")
}

func isReplayable(r *http.Request) bool {
	if r.Body == nil || r.Body == NoBody || r.GetBody != nil {
		switch valueOrDefault(r.Method, "GET") {
		case "GET", "HEAD", "OPTIONS", "TRACE":
			return true
		}
		// The Idempotency-Key, while non-standard, is widely used to
		// mean a POST or other request is idempotent. See
		// https://golang.org/issue/19943#issuecomment-421092421
		if headerHas(r.Header, "Idempotency-Key") || headerHas(r.Header, "X-Idempotency-Key") {
			return true
		}
	}
	return false
}

func reqExpectsContinue(r *http.Request) bool {
	return hasToken(headerGet(r.Header, "Expect"), "100-continue")
}

func reqWantsClose(r *http.Request) bool {
	if r.Close {
		return true
	}
	return hasToken(headerGet(r.Header, "Connection"), "close")
}
