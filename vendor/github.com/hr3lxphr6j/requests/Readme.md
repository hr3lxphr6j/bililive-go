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
package main

import (
	"fmt"

	"github.com/hr3lxphr6j/requests"
)

func main() {
	resp, err := requests.Post("http://example.com",
		requests.JSON(map[string]string{"foo": "bar"}),
		requests.Query("foo", "bar"),
	)
	if err != nil {
		panic(err)
	}
	m := make(map[string]interface{})
	if err := resp.JSON(m); err != nil {
		panic(err)
	}
	fmt.Println(m)
}
```
