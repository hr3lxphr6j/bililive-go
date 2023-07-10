package requests

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// RequestOptions is a collection of request options.
type RequestOptions struct {
	Headers  map[string]interface{}
	Queries  url.Values
	Cookies  map[string]string
	Body     io.Reader
	Deadline time.Time

	multipartWriter *multipart.Writer

	Err error
}

func (o *RequestOptions) getMultipartWriter() *multipart.Writer {
	if o.multipartWriter == nil {
		Body(new(bytes.Buffer))(o)
		o.multipartWriter = multipart.NewWriter(o.Body.(*bytes.Buffer))
	}
	return o.multipartWriter
}

// NewOptions return a new *RequestOptions.
func NewOptions() *RequestOptions {
	return &RequestOptions{
		Headers: map[string]interface{}{},
		Queries: url.Values{},
		Cookies: map[string]string{},
	}
}

// RequestOption is used to update the fields in RequestOptions.
type RequestOption func(o *RequestOptions)

// Deadline set the deadline of this request.
func Deadline(t time.Time) RequestOption {
	return func(o *RequestOptions) {
		o.Deadline = t
	}
}

// Timeout set the timeout of this request.
func Timeout(d time.Duration) RequestOption {
	return func(o *RequestOptions) {
		o.Deadline = time.Now().Add(d)
	}
}

// Header sets the request's header, v can be a string, fmt.Stringer or nil,
// when v is nil this header will be removed.
func Header(k string, v interface{}) RequestOption {
	return func(o *RequestOptions) {
		o.Headers[k] = v
	}
}

// Headers sets the request's headers, value of map be a string, fmt.Stringer or nil,
// when v is nil this header will be removed.
// If the replace flag is true, the existing header will be removed.
func Headers(m map[string]interface{}, replace ...bool) RequestOption {
	return func(o *RequestOptions) {
		if len(replace) == 1 && replace[0] {
			o.Headers = m
			return
		}
		for k, v := range m {
			Header(k, v)(o)
		}
	}
}

// UserAgent sets the request's UserAgent header.
func UserAgent(ua string) RequestOption {
	return func(o *RequestOptions) {
		Header(HeaderUserAgent, ua)(o)
	}
}

// ContentType sets the request's ContentType header.
func ContentType(ct string) RequestOption {
	return func(o *RequestOptions) {
		Header(HeaderContentType, ct)(o)
	}
}

// Referer sets the request's Referer header.
func Referer(r string) RequestOption {
	return func(o *RequestOptions) {
		Header(HeaderReferer, r)(o)
	}
}

// Authorization sets the request's Authorization header.
func Authorization(a string) RequestOption {
	return func(o *RequestOptions) {
		Header(HeaderAuthorization, a)(o)
	}
}

// BasicAuth sets the request's Authorization header to use HTTP
// Basic Authentication with the provided username and password.
func BasicAuth(username, password string) RequestOption {
	return func(o *RequestOptions) {
		Authorization("Basic " + base64.StdEncoding.EncodeToString(
			[]byte(fmt.Sprintf("%s:%s", username, password))))(o)
	}
}

// Cookie sets the request's cookie.
func Cookie(k, v string) RequestOption {
	return func(o *RequestOptions) {
		o.Cookies[k] = v
	}
}

// Cookies sets the request's cookies,
// if the replace flag is true, the existing cookies will be removed.
func Cookies(m map[string]string, replace ...bool) RequestOption {
	return func(o *RequestOptions) {
		if len(replace) == 1 && replace[0] {
			o.Cookies = m
			return
		}
		for k, v := range m {
			o.Cookies[k] = v
		}
	}
}

// Query sets the request's query.
func Query(k, v string) RequestOption {
	return func(o *RequestOptions) {
		o.Queries.Set(k, v)
	}
}

// Queries sets the request's queries.
func Queries(m map[string]string, replace ...bool) RequestOption {
	return func(o *RequestOptions) {
		if len(replace) == 1 && replace[0] {
			o.Queries = url.Values{}
		}
		for k, v := range m {
			o.Queries.Set(k, v)
		}
	}
}

// QueriesFromValue sets the request's queries.
func QueriesFromValue(v url.Values) RequestOption {
	return func(o *RequestOptions) {
		o.Queries = v
	}
}

// Body sets the request's body.
func Body(r io.Reader) RequestOption {
	return func(o *RequestOptions) {
		o.Body = r
	}
}

// JSON marshal i to JSON format and sets it to request's body.
// ContentType will set to ContentTypeJSON.
func JSON(i interface{}) RequestOption {
	return func(o *RequestOptions) {
		ContentType(ContentTypeJSON)(o)
		Body(new(bytes.Buffer))(o)
		if err := json.NewEncoder(o.Body.(*bytes.Buffer)).Encode(i); err != nil {
			o.Err = fmt.Errorf("json option error %w", err)
		}
	}
}

// Form sets m as form format of request's body.
// ContentType will set to ContentTypeForm.
func Form(m map[string]string) RequestOption {
	return func(o *RequestOptions) {
		ContentType(ContentTypeForm)(o)
		value := url.Values{}
		for k, v := range m {
			value.Set(k, v)
		}
		Body(strings.NewReader(value.Encode()))(o)
	}
}

// FileFromReader build the body for a multipart/form-data request.
func FileFromReader(fieldName, fileName string, r io.Reader) RequestOption {
	return func(o *RequestOptions) {
		mw := o.getMultipartWriter()
		w, err := mw.CreateFormFile(fieldName, fileName)
		if err != nil {
			o.Err = fmt.Errorf("fileFromReader option error %w", err)
			return
		}
		_, err = io.Copy(w, r)
		if err != nil {
			o.Err = fmt.Errorf("fileFromReader option error %w", err)
			return
		}
		ContentType(mw.FormDataContentType())(o)
	}
}

// File build the body for a multipart/form-data request.
func File(fieldName string, file *os.File) RequestOption {
	return FileFromReader(fieldName, filepath.Base(file.Name()), file)
}

// MultipartField build the body for a multipart/form-data request.
func MultipartField(fieldName string, r io.Reader) RequestOption {
	return func(o *RequestOptions) {
		mw := o.getMultipartWriter()
		w, err := mw.CreateFormField(fieldName)
		if err != nil {
			o.Err = fmt.Errorf("multipartField option error %w", err)
			return
		}
		_, err = io.Copy(w, r)
		if err != nil {
			o.Err = fmt.Errorf("file option error %w", err)
			return
		}
		ContentType(mw.FormDataContentType())(o)
	}
}

// MultipartFieldString build the body for a multipart/form-data request.
func MultipartFieldString(fieldName, value string) RequestOption {
	return MultipartField(fieldName, strings.NewReader(value))
}

// MultipartFieldBytes build the body for a multipart/form-data request.
func MultipartFieldBytes(fieldName string, value []byte) RequestOption {
	return MultipartField(fieldName, bytes.NewReader(value))
}
