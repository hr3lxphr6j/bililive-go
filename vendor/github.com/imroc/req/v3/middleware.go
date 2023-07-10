package req

import (
	"bytes"
	"github.com/imroc/req/v3/internal/header"
	"github.com/imroc/req/v3/internal/util"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"
)

type (
	// RequestMiddleware type is for request middleware, called before a request is sent
	RequestMiddleware func(client *Client, req *Request) error

	// ResponseMiddleware type is for response middleware, called after a response has been received
	ResponseMiddleware func(client *Client, resp *Response) error
)

func createMultipartHeader(file *FileUpload, contentType string) textproto.MIMEHeader {
	hdr := make(textproto.MIMEHeader)

	contentDispositionValue := "form-data"
	cd := new(ContentDisposition)
	if file.ParamName != "" {
		cd.Add("name", file.ParamName)
	}
	if file.FileName != "" {
		cd.Add("filename", file.FileName)
	}
	if file.ExtraContentDisposition != nil {
		for _, kv := range file.ExtraContentDisposition.kv {
			cd.Add(kv.Key, kv.Value)
		}
	}
	if c := cd.string(); c != "" {
		contentDispositionValue += c
	}
	hdr.Set("Content-Disposition", contentDispositionValue)

	if !util.IsStringEmpty(contentType) {
		hdr.Set(header.ContentType, contentType)
	}
	return hdr
}

func closeq(v interface{}) {
	if c, ok := v.(io.Closer); ok {
		c.Close()
	}
}

func writeMultipartFormFile(w *multipart.Writer, file *FileUpload, r *Request) error {
	content, err := file.GetFileContent()
	if err != nil {
		return err
	}
	defer content.Close()
	if r.RetryAttempt > 0 { // reset file reader when retry a multipart file upload
		if rs, ok := content.(io.ReadSeeker); ok {
			_, err = rs.Seek(0, io.SeekStart)
			if err != nil {
				return err
			}
		}
	}
	// Auto detect actual multipart content type
	cbuf := make([]byte, 512)
	seeEOF := false
	lastTime := time.Now()
	size, err := content.Read(cbuf)
	if err != nil {
		if err == io.EOF {
			seeEOF = true
		} else {
			return err
		}
	}

	ct := file.ContentType
	if ct == "" {
		ct = http.DetectContentType(cbuf)
	}
	pw, err := w.CreatePart(createMultipartHeader(file, ct))
	if err != nil {
		return err
	}

	if r.forceChunkedEncoding && r.uploadCallback != nil {
		pw = &callbackWriter{
			Writer:    pw,
			lastTime:  lastTime,
			interval:  r.uploadCallbackInterval,
			totalSize: file.FileSize,
			callback: func(written int64) {
				r.uploadCallback(UploadInfo{
					ParamName:    file.ParamName,
					FileName:     file.FileName,
					FileSize:     file.FileSize,
					UploadedSize: written,
				})
			},
		}
	}

	if _, err = pw.Write(cbuf[:size]); err != nil {
		return err
	}
	if seeEOF {
		return nil
	}

	_, err = io.Copy(pw, content)
	return err
}

func writeMultiPart(r *Request, w *multipart.Writer) {
	defer w.Close() // close multipart to write tailer boundary
	for k, vs := range r.FormData {
		for _, v := range vs {
			w.WriteField(k, v)
		}
	}
	for _, file := range r.uploadFiles {
		writeMultipartFormFile(w, file, r)
	}
}

func handleMultiPart(c *Client, r *Request) (err error) {
	if r.forceChunkedEncoding {
		pr, pw := io.Pipe()
		r.GetBody = func() (io.ReadCloser, error) {
			return pr, nil
		}
		w := multipart.NewWriter(pw)
		r.SetContentType(w.FormDataContentType())
		go func() {
			writeMultiPart(r, w)
			pw.Close() // close pipe writer so that pipe reader could get EOF, and stop upload
		}()
	} else {
		buf := new(bytes.Buffer)
		w := multipart.NewWriter(buf)
		writeMultiPart(r, w)
		r.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(buf.Bytes())), nil
		}
		r.Body = buf.Bytes()
		r.SetContentType(w.FormDataContentType())
	}
	return
}

func handleFormData(r *Request) {
	r.SetContentType(header.FormContentType)
	r.SetBodyBytes([]byte(r.FormData.Encode()))
}

func handleMarshalBody(c *Client, r *Request) error {
	ct := ""
	if r.Headers != nil {
		ct = r.Headers.Get(header.ContentType)
	}
	if ct == "" {
		ct = c.Headers.Get(header.ContentType)
	}
	if ct != "" {
		if util.IsXMLType(ct) {
			body, err := c.xmlMarshal(r.marshalBody)
			if err != nil {
				return err
			}
			r.SetBodyBytes(body)
		} else {
			body, err := c.jsonMarshal(r.marshalBody)
			if err != nil {
				return err
			}
			r.SetBodyBytes(body)
		}
		return nil
	}
	body, err := c.jsonMarshal(r.marshalBody)
	if err != nil {
		return err
	}
	r.SetBodyJsonBytes(body)
	return nil
}

func parseRequestBody(c *Client, r *Request) (err error) {
	if c.isPayloadForbid(r.Method) {
		r.marshalBody = nil
		r.Body = nil
		r.GetBody = nil
		return
	}
	// handle multipart
	if r.isMultiPart {
		return handleMultiPart(c, r)
	}

	// handle form data
	if len(c.FormData) > 0 {
		r.SetFormDataFromValues(c.FormData)
	}
	if len(r.FormData) > 0 {
		handleFormData(r)
		return
	}

	// handle marshal body
	if r.marshalBody != nil {
		err = handleMarshalBody(c, r)
		if err != nil {
			return
		}
	}

	if r.Body == nil {
		return
	}
	// body is in-memory []byte, so we can guess content type
	if r.getHeader(header.ContentType) == "" {
		r.SetContentType(http.DetectContentType(r.Body))
	}
	return
}

func unmarshalBody(c *Client, r *Response, v interface{}) (err error) {
	body, err := r.ToBytes() // in case req.SetResult or req.SetError with cient.DisalbeAutoReadResponse(true)
	if err != nil {
		return
	}
	ct := r.GetContentType()
	if util.IsJSONType(ct) {
		return c.jsonUnmarshal(body, v)
	} else if util.IsXMLType(ct) {
		return c.xmlUnmarshal(body, v)
	} else {
		if c.DebugLog {
			c.log.Debugf("cannot determine the unmarshal function with %q Content-Type, default to json", ct)
		}
		return c.jsonUnmarshal(body, v)
	}
	return
}

func defaultResultStateChecker(resp *Response) ResultState {
	if code := resp.StatusCode; code > 199 && code < 300 {
		return SuccessState
	} else if code > 399 {
		return ErrorState
	} else {
		return UnknownState
	}
}

func parseResponseBody(c *Client, r *Response) (err error) {
	if r.Response == nil {
		return
	}
	req := r.Request
	switch r.ResultState() {
	case SuccessState:
		if req.Result != nil && r.StatusCode != http.StatusNoContent {
			err = unmarshalBody(c, r, r.Request.Result)
			if err == nil {
				r.result = r.Request.Result
			}
		}
	case ErrorState:
		if r.StatusCode == http.StatusNoContent {
			return
		}
		if req.Error != nil {
			err = unmarshalBody(c, r, req.Error)
			if err == nil {
				r.error = req.Error
			}
		} else if c.commonErrorType != nil {
			e := reflect.New(c.commonErrorType).Interface()
			err = unmarshalBody(c, r, e)
			if err == nil {
				r.error = e
			}
		}
	}
	return
}

type callbackWriter struct {
	io.Writer
	written   int64
	totalSize int64
	lastTime  time.Time
	interval  time.Duration
	callback  func(written int64)
}

func (w *callbackWriter) Write(p []byte) (n int, err error) {
	n, err = w.Writer.Write(p)
	if n <= 0 {
		return
	}
	w.written += int64(n)
	if w.written == w.totalSize {
		w.callback(w.written)
	} else if now := time.Now(); now.Sub(w.lastTime) >= w.interval {
		w.lastTime = now
		w.callback(w.written)
	}
	return
}

type callbackReader struct {
	io.ReadCloser
	read     int64
	lastRead int64
	callback func(read int64)
	lastTime time.Time
	interval time.Duration
}

func (r *callbackReader) Read(p []byte) (n int, err error) {
	n, err = r.ReadCloser.Read(p)
	if n <= 0 {
		if err == io.EOF && r.read > r.lastRead {
			r.callback(r.read)
			r.lastRead = r.read
		}
		return
	}
	r.read += int64(n)
	if err == io.EOF {
		r.callback(r.read)
		r.lastRead = r.read
	} else if now := time.Now(); now.Sub(r.lastTime) >= r.interval {
		r.lastTime = now
		r.callback(r.read)
		r.lastRead = r.read
	}
	return
}

func handleDownload(c *Client, r *Response) (err error) {
	if r.Response == nil || !r.Request.isSaveResponse {
		return nil
	}
	var body io.ReadCloser

	if r.body != nil { // already read
		body = io.NopCloser(bytes.NewReader(r.body))
	} else {
		body = r.Body
	}

	var output io.Writer
	if r.Request.outputFile != "" {
		file := r.Request.outputFile
		if c.outputDirectory != "" && !filepath.IsAbs(file) {
			file = c.outputDirectory + string(filepath.Separator) + file
		}

		file = filepath.Clean(file)

		if err = util.CreateDirectory(filepath.Dir(file)); err != nil {
			return err
		}
		output, err = os.Create(file)
		if err != nil {
			return
		}
	} else {
		output = r.Request.output // must not nil
	}

	defer func() {
		body.Close()
		closeq(output)
	}()

	_, err = io.Copy(output, body)
	r.setReceivedAt()
	return
}

func parseRequestHeader(c *Client, r *Request) error {
	if c.Headers == nil {
		return nil
	}
	if r.Headers == nil {
		r.Headers = make(http.Header)
	}
	for k, vs := range c.Headers {
		for _, v := range vs {
			if len(r.Headers[k]) == 0 {
				r.Headers[k] = append(r.Headers[k], v)
			}
		}
	}
	return nil
}

func parseRequestCookie(c *Client, r *Request) error {
	if len(c.Cookies) == 0 {
		return nil
	}
	for _, ck := range c.Cookies {
		r.Cookies = append(r.Cookies, ck)
	}
	return nil
}

// generate URL
func parseRequestURL(c *Client, r *Request) error {
	tempURL := r.RawURL
	if len(r.PathParams) > 0 {
		for p, v := range r.PathParams {
			tempURL = strings.Replace(tempURL, "{"+p+"}", url.PathEscape(v), -1)
		}
	}
	if len(c.PathParams) > 0 {
		for p, v := range c.PathParams {
			tempURL = strings.Replace(tempURL, "{"+p+"}", url.PathEscape(v), -1)
		}
	}

	// Parsing request URL
	reqURL, err := url.Parse(tempURL)
	if err != nil {
		return err
	}

	if reqURL.Scheme == "" && len(c.scheme) > 0 { // set scheme if missing
		reqURL, err = url.Parse(c.scheme + "://" + tempURL)
		if err != nil {
			return err
		}
	}

	// If RawURL is relative path then added c.BaseURL into
	// the request URL otherwise Request.URL will be used as-is
	if !reqURL.IsAbs() {
		tempURL = reqURL.String()
		if len(tempURL) > 0 && tempURL[0] != '/' {
			tempURL = "/" + tempURL
		}

		reqURL, err = url.Parse(c.BaseURL + tempURL)
		if err != nil {
			return err
		}
	}

	// Adding Query Param
	query := make(url.Values)
	for k, v := range c.QueryParams {
		for _, iv := range v {
			query.Add(k, iv)
		}
	}

	for k, v := range r.QueryParams {
		// remove query param from client level by key
		// since overrides happens for that key in the request
		query.Del(k)

		for _, iv := range v {
			query.Add(k, iv)
		}
	}

	// Preserve query string order partially.
	// Since not feasible in `SetQuery*` resty methods, because
	// standard package `url.Encode(...)` sorts the query params
	// alphabetically
	if len(query) > 0 {
		if util.IsStringEmpty(reqURL.RawQuery) {
			reqURL.RawQuery = query.Encode()
		} else {
			reqURL.RawQuery = reqURL.RawQuery + "&" + query.Encode()
		}
	}

	reqURL.Host = removeEmptyPort(reqURL.Host)
	r.URL = reqURL
	return nil
}
