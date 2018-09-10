package http

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/mozillazg/request"
)

var commonHeader = map[string]string{
	"Accept":          "application/json, text/javascript, */*; q=0.01",
	"Accept-Encoding": "gzip, deflate",
	"Accept-Language": "zh-CN,zh;q=0.8,en-US;q=0.6,en;q=0.4,zh-TW;q=0.2",
	"Connection":      "keep-alive",
	"User-Agent":      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Safari/537.36",
}

var client = new(http.Client)

func parseResponse(resp *request.Response) ([]byte, error) {
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

func Get(url string, query map[string]string, header map[string]string) ([]byte, error) {
	req := request.NewRequest(client)
	if header != nil {
		req.Headers = header
	} else {
		req.Headers = commonHeader
	}
	req.Params = query

	if resp, err := req.Get(url); err == nil {
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return nil, errors.New(resp.Status)
		}
		return parseResponse(resp)
	} else {
		return nil, err
	}
}

func Post(url string, query map[string]string, body []byte, header map[string]string) ([]byte, error) {
	req := request.NewRequest(client)
	if header != nil {
		req.Headers = header
	} else {
		req.Headers = commonHeader
	}
	req.Params = query
	req.Body = bytes.NewReader(body)

	if resp, err := req.Post(url); err == nil {
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return nil, errors.New(resp.Status)
		}
		return parseResponse(resp)
	} else {
		return nil, err
	}
}
