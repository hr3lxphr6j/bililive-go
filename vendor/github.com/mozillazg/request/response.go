package request

import (
	"compress/gzip"
	"compress/zlib"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/bitly/go-simplejson"
)

// Response ...
type Response struct {
	*http.Response
	content []byte
}

// Json return Response Body as simplejson.Json
func (resp *Response) Json() (*simplejson.Json, error) {
	b, err := resp.Content()
	if err != nil {
		return nil, err
	}
	return simplejson.NewJson(b)
}

// Content return Response Body as []byte
func (resp *Response) Content() (b []byte, err error) {
	if resp.content != nil {
		return resp.content, nil
	}

	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		if reader, err = gzip.NewReader(resp.Body); err != nil {
			return nil, err
		}
	case "deflate":
		if reader, err = zlib.NewReader(resp.Body); err != nil {
			return nil, err
		}
	default:
		reader = resp.Body
	}

	defer reader.Close()
	if b, err = ioutil.ReadAll(reader); err != nil {
		return nil, err
	}

	resp.content = b
	return b, err
}

// Text return Response Body as string
func (resp *Response) Text() (string, error) {
	b, err := resp.Content()
	s := string(b)
	return s, err
}

// OK check Response StatusCode < 400 ?
func (resp *Response) OK() bool {
	return resp.StatusCode < 400
}

// Ok check Response StatusCode < 400 ?
func (resp *Response) Ok() bool {
	return resp.OK()
}

// Reason return Response Status
func (resp *Response) Reason() string {
	return resp.Status
}

// URL return finally request url
func (resp *Response) URL() (*url.URL, error) {
	u := resp.Request.URL
	switch resp.StatusCode {
	case http.StatusMovedPermanently, http.StatusFound,
		http.StatusSeeOther, http.StatusTemporaryRedirect:
		location, err := resp.Location()
		if err != nil {
			return nil, err
		}
		u = u.ResolveReference(location)
	}
	return u, nil
}
