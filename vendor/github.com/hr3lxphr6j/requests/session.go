package requests

import (
	"net/http"
)

// Session is a wrapper of *http.Client.
type Session struct {
	*http.Client
}

// DefaultSession is a wrapper of http.DefaultClient.
var DefaultSession = NewSession(http.DefaultClient)

// NewSession return a new *Session
func NewSession(c *http.Client) *Session {
	return &Session{Client: c}
}

// Do sends an HTTP request and returns an HTTP response.
func (s *Session) Do(r *http.Request) (*Response, error) {
	resp, err := s.Client.Do(r)
	if err != nil {
		return nil, err
	}
	return NewResponse(resp), nil
}

// Request sends an HTTP request and returns an HTTP response.
func (s *Session) Request(method, url string, opts ...RequestOption) (*Response, error) {
	req, err := NewRequest(method, url, opts...)
	if err != nil {
		return nil, err
	}
	return s.Do(req)
}

// Get sends a GET request.
func (s *Session) Get(url string, opts ...RequestOption) (*Response, error) {
	return s.Request(http.MethodGet, url, opts...)
}

// Head sends a HEAT request.
func (s *Session) Head(url string, opts ...RequestOption) (*Response, error) {
	return s.Request(http.MethodHead, url, opts...)
}

// Post sends a POST request.
func (s *Session) Post(url string, opts ...RequestOption) (*Response, error) {
	return s.Request(http.MethodPost, url, opts...)
}

// Put sends a PUT request.
func (s *Session) Put(url string, opts ...RequestOption) (*Response, error) {
	return s.Request(http.MethodPut, url, opts...)
}

// Patch sends a PATCH request.
func (s *Session) Patch(url string, opts ...RequestOption) (*Response, error) {
	return s.Request(http.MethodPatch, url, opts...)
}

// Delete sends a DELETE request.
func (s *Session) Delete(url string, opts ...RequestOption) (*Response, error) {
	return s.Request(http.MethodDelete, url, opts...)
}

// Connect sends a CONNECT request.
func (s *Session) Connect(url string, opts ...RequestOption) (*Response, error) {
	return s.Request(http.MethodConnect, url, opts...)
}

// Options sends a OPTIONS request.
func (s *Session) Options(url string, opts ...RequestOption) (*Response, error) {
	return s.Request(http.MethodOptions, url, opts...)
}

// Trace sends a TRACE request.
func (s *Session) Trace(url string, opts ...RequestOption) (*Response, error) {
	return s.Request(http.MethodTrace, url, opts...)
}
