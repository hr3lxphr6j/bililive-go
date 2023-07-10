package requests

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
)

// NewRequest return a new *http.Request
func NewRequest(method, url string, opts ...RequestOption) (*http.Request, error) {
	return NewRequestWithContext(context.Background(), method, url, opts...)
}

// NewRequestWithContext return a new *http.Request
func NewRequestWithContext(ctx context.Context, method, url string, opts ...RequestOption) (*http.Request, error) {
	options := NewOptions()
	for _, opt := range opts {
		opt(options)
		if options.Err != nil {
			return nil, options.Err
		}
	}
	if options.multipartWriter != nil {
		if err := options.multipartWriter.Close(); err != nil {
			return nil, fmt.Errorf("write multipart error %w", err)
		}
	}

	if !options.Deadline.IsZero() {
		ctx, _ = context.WithDeadline(ctx, options.Deadline)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, options.Body)
	if err != nil {
		return nil, err
	}

	// set headers
	for k, v := range options.Headers {
		switch val := v.(type) {
		case string:
			req.Header.Set(k, val)
		case fmt.Stringer:
			req.Header.Set(k, val.String())
		case nil:
			req.Header.Del(k)
		default:
			return nil, fmt.Errorf("value of header [%s] must be string or nil, but %s", k, reflect.TypeOf(v))
		}
	}

	// set queries
	values := req.URL.Query()
	for k, vs := range options.Queries {
		for _, v := range vs {
			values.Add(k, v)
		}
	}
	req.URL.RawQuery = values.Encode()

	// set cookies
	for k, v := range options.Cookies {
		req.AddCookie(&http.Cookie{Name: k, Value: v})
	}

	return req, nil
}

// Request sends an HTTP request and returns an HTTP response.
func Request(method, url string, opts ...RequestOption) (*Response, error) {
	return DefaultSession.Request(method, url, opts...)
}

// Get sends a GET request.
func Get(url string, opts ...RequestOption) (*Response, error) {
	return DefaultSession.Get(url, opts...)
}

// Head sends a HEAT request.
func Head(url string, opts ...RequestOption) (*Response, error) {
	return DefaultSession.Head(url, opts...)
}

// Post sends a POST request.
func Post(url string, opts ...RequestOption) (*Response, error) {
	return DefaultSession.Post(url, opts...)
}

// Put sends a PUT request.
func Put(url string, opts ...RequestOption) (*Response, error) {
	return DefaultSession.Put(url, opts...)
}

// Patch sends a PATCH request.
func Patch(url string, opts ...RequestOption) (*Response, error) {
	return DefaultSession.Patch(url, opts...)
}

// Delete sends a DELETE request.
func Delete(url string, opts ...RequestOption) (*Response, error) {
	return DefaultSession.Delete(url, opts...)
}

// Connect sends a CONNECT request.
func Connect(url string, opts ...RequestOption) (*Response, error) {
	return DefaultSession.Connect(url, opts...)
}

// Options sends a OPTIONS request.
func Options(url string, opts ...RequestOption) (*Response, error) {
	return DefaultSession.Options(url, opts...)
}

// Trace sends a TRACE request.
func Trace(url string, opts ...RequestOption) (*Response, error) {
	return DefaultSession.Trace(url, opts...)
}
