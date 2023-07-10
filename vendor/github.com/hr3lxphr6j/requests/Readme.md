# Requests

[![Build Status](https://travis-ci.org/hr3lxphr6j/requests.svg?branch=master)](https://travis-ci.org/hr3lxphr6j/requests)
[![Go Report Card](https://goreportcard.com/badge/github.com/hr3lxphr6j/requests)](https://goreportcard.com/report/github.com/hr3lxphr6j/requests)
[![codecov](https://codecov.io/gh/hr3lxphr6j/requests/branch/master/graph/badge.svg)](https://codecov.io/gh/hr3lxphr6j/requests)

A "Requests" style HTTP client for golang

## Install

```shell
go get -u github.com/hr3lxphr6j/requests
```

## Example

```go
var (
	timeout     = time.Second * 3
	url         = "http://example.com"
	queryParams = map[string]string{"foo": "bar"}
	body        = map[string]interface{}{"a": "b", "nums": []int{1, 2, 3}}
)

// Do `curl --connect-timeout 3 -d '{"a": "b", "num": [1, 2, 3]}' -H 'Content-Type: application/json' http://example.com?foo=bar`

// With net/http
func UseNetHttp() map[string]interface{} {
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(body); err != nil {
		log.Panic(err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, buf)
	if err != nil {
		log.Panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	value := urlPkg.Values{}
	for k, v := range queryParams {
		value.Set(k, v)
	}
	req.URL.RawQuery = value.Encode()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Panic(err)
	}
	defer resp.Body.Close()
	data := map[string]interface{}{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Panic(err)
	}
	return data
}

// With Requests
func UseRequests() map[string]interface{} {
	resp, err := requests.Post(url,
		requests.Timeout(timeout),
		requests.Queries(queryParams),
		requests.JSON(body),
	)
	if err != nil {
		log.Panic(err)
	}
	data := map[string]interface{}{}
	if err := resp.JSON(&data); err != nil {
		log.Panic(err)
	}
	return data
}
```
