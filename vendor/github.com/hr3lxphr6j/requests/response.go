package requests

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// Response is wrapper of *http.Response.
type Response struct {
	*http.Response
}

// NewResponse return a new *Response.
func NewResponse(r *http.Response) *Response {
	return &Response{Response: r}
}

// StdResponse return a unwrapped *http.Response.
func (r *Response) StdResponse() *http.Response {
	return r.Response
}

// Bytes return the response`s body as []byte.
func (r *Response) Bytes() ([]byte, error) {
	defer r.Body.Close()
	return ioutil.ReadAll(r.Body)
}

// Text return the response`s body as string.
func (r *Response) Text() (string, error) {
	b, err := r.Bytes()
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// JSON unmarshal the response`s body as JSON.
func (r *Response) JSON(i interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(i)
}
