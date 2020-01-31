package http

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

var commonHeader = map[string]string{
	"Accept":          "application/json, text/javascript, */*; q=0.01",
	"Accept-Encoding": "gzip, deflate",
	"Accept-Language": "zh-CN,zh;q=0.8,en-US;q=0.6,en;q=0.4,zh-TW;q=0.2",
	"Connection":      "keep-alive",
	"User-Agent":      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Safari/537.36",
}

var client = new(http.Client)

func Get(path string, header, query map[string]string) ([]byte, error) {
	return do(http.MethodGet, path, header, query, nil)
}

func Post(path string, header, query map[string]string, body []byte) ([]byte, error) {
	return do(http.MethodPost, path, header, query, body)
}

func parseResponse(resp *http.Response) ([]byte, error) {
	var reader io.ReadCloser
	defer func() {
		if reader != nil {
			reader.Close()
		}
	}()

	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		r, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		reader = r
	default:
		reader = resp.Body
	}

	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func do(method string, path string, header, query map[string]string, body []byte) ([]byte, error) {
	var r io.Reader = nil
	if body != nil {
		r = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, path, r)
	if err != nil {
		return nil, err
	}
	if header == nil || len(header) == 0 {
		header = commonHeader
	}
	for k, v := range header {
		req.Header.Set(k, v)
	}
	q := url.Values{}
	for k, v := range query {
		q.Set(k, v)
	}
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}
	return parseResponse(resp)
}
