// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !js

package req

import (
	"net/http"
)

// RoundTrip implements the RoundTripper interface.
//
// For higher-level HTTP client support (such as handling of cookies
// and redirects), see Get, Post, and the Client type.
//
// Like the RoundTripper interface, the error types returned
// by RoundTrip are unspecified.
func (t *Transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	if t.wrappedRoundTrip != nil {
		resp, err = t.wrappedRoundTrip.RoundTrip(req)
	} else {
		resp, err = t.roundTrip(req)
	}
	if err != nil {
		return
	}
	if resp.ProtoMajor != 3 && t.altSvcJar != nil {
		if v := resp.Header.Get("alt-svc"); v != "" {
			t.handleAltSvc(req, v)
		}
	}
	t.handleResponseBody(resp, req)
	return
}
