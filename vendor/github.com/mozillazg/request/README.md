request
=======
[![Build Status](https://travis-ci.org/mozillazg/request.svg?branch=master)](https://travis-ci.org/mozillazg/request)
[![Coverage Status](https://coveralls.io/repos/mozillazg/request/badge.png?branch=master)](https://coveralls.io/r/mozillazg/request?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/mozillazg/request)](https://goreportcard.com/report/github.com/mozillazg/request)
[![GoDoc](https://godoc.org/github.com/mozillazg/request?status.svg)](https://godoc.org/github.com/mozillazg/request)

A developer-friendly HTTP request library for Gopher. Inspired by [Python-Requests](https://github.com/kennethreitz/requests).


Installation
------------

```
go get -u github.com/mozillazg/request
```


Documentation
--------------

API documentation can be found here:
https://godoc.org/github.com/mozillazg/request


Usage
-------

```go
import (
    "github.com/mozillazg/request"
)
```

**GET**:

```go
c := new(http.Client)
req := request.NewRequest(c)
resp, err := req.Get("http://httpbin.org/get")
j, err := resp.Json()
defer resp.Body.Close()  // Don't forget close the response body
```

**POST**:

```go
req := request.NewRequest(c)
req.Data = map[string]string{
    "key": "value",
    "a":   "123",
}
resp, err := req.Post("http://httpbin.org/post")
```

**Cookies**:

```go
req := request.NewRequest(c)
req.Cookies = map[string]string{
    "key": "value",
    "a":   "123",
}
resp, err := req.Get("http://httpbin.org/cookies")
```

**Headers**:

```go
req := request.NewRequest(c)
req.Headers = map[string]string{
    "Accept-Encoding": "gzip,deflate,sdch",
    "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
}
resp, err := req.Get("http://httpbin.org/get")
```

**Files**:

```go
req := request.NewRequest(c)
f, err := os.Open("test.txt")
req.Files = []request.FileField{
    request.FileField{"file", "test.txt", f},
}
resp, err := req.Post("http://httpbin.org/post")
```

**Json**:

```go
req := request.NewRequest(c)
req.Json = map[string]string{
    "a": "A",
    "b": "B",
}
resp, err := req.Post("http://httpbin.org/post")
req.Json = []int{1, 2, 3}
resp, err = req.Post("http://httpbin.org/post")
```

**Proxy**:
```go
req := request.NewRequest(c)
req.Proxy = "http://127.0.0.1:8080"
// req.Proxy = "https://127.0.0.1:8080"
// req.Proxy = "socks5://127.0.0.1:57341"
resp, err := req.Get("http://httpbin.org/get")
```
or https://github.com/mozillazg/request/tree/develop/_example/proxy

**HTTP Basic Authentication**:
```go
req := request.NewRequest(c)
req.BasicAuth = request.BasicAuth{"user", "passwd"}
resp, err := req.Get("http://httpbin.org/basic-auth/user/passwd")
```


License
---------

Under the MIT License.
