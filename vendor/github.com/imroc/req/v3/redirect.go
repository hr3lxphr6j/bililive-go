package req

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
)

// RedirectPolicy represents the redirect policy for Client.
type RedirectPolicy func(req *http.Request, via []*http.Request) error

// MaxRedirectPolicy specifies the max number of redirect
func MaxRedirectPolicy(noOfRedirect int) RedirectPolicy {
	return func(req *http.Request, via []*http.Request) error {
		if len(via) >= noOfRedirect {
			return fmt.Errorf("stopped after %d redirects", noOfRedirect)
		}
		return nil
	}
}

// NoRedirectPolicy disable redirect behaviour
func NoRedirectPolicy() RedirectPolicy {
	return func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
}

// SameDomainRedirectPolicy allows redirect only if the redirected domain
// is the same as original domain, e.g. redirect to "www.imroc.cc" from
// "imroc.cc" is allowed, but redirect to "google.com" is not allowed.
func SameDomainRedirectPolicy() RedirectPolicy {
	return func(req *http.Request, via []*http.Request) error {
		if getDomain(req.URL.Host) != getDomain(via[0].URL.Host) {
			return errors.New("different domain name is not allowed")
		}
		return nil
	}
}

// SameHostRedirectPolicy allows redirect only if the redirected host
// is the same as original host, e.g. redirect to "www.imroc.cc" from
// "imroc.cc" is not the allowed.
func SameHostRedirectPolicy() RedirectPolicy {
	return func(req *http.Request, via []*http.Request) error {
		if getHostname(req.URL.Host) != getHostname(via[0].URL.Host) {
			return errors.New("different host name is not allowed")
		}
		return nil
	}
}

// AllowedHostRedirectPolicy allows redirect only if the redirected host
// match one of the host that specified.
func AllowedHostRedirectPolicy(hosts ...string) RedirectPolicy {
	m := make(map[string]bool)
	for _, h := range hosts {
		m[strings.ToLower(getHostname(h))] = true
	}

	return func(req *http.Request, via []*http.Request) error {
		h := getHostname(req.URL.Host)
		if _, ok := m[h]; !ok {
			return fmt.Errorf("redirect host [%s] is not allowed", h)
		}
		return nil
	}
}

// AllowedDomainRedirectPolicy allows redirect only if the redirected domain
// match one of the domain that specified.
func AllowedDomainRedirectPolicy(hosts ...string) RedirectPolicy {
	domains := make(map[string]bool)
	for _, h := range hosts {
		domains[strings.ToLower(getDomain(h))] = true
	}

	return func(req *http.Request, via []*http.Request) error {
		domain := getDomain(req.URL.Host)
		if _, ok := domains[domain]; !ok {
			return fmt.Errorf("redirect domain [%s] is not allowed", domain)
		}
		return nil
	}
}

func getHostname(host string) (hostname string) {
	if strings.Index(host, ":") > 0 {
		host, _, _ = net.SplitHostPort(host)
	}
	hostname = strings.ToLower(host)
	return
}

func getDomain(host string) string {
	host = getHostname(host)
	ss := strings.Split(host, ".")
	if len(ss) < 3 {
		return host
	}
	ss = ss[1:]
	return strings.Join(ss, ".")
}

// AlwaysCopyHeaderRedirectPolicy ensures that the given sensitive headers will
// always be copied on redirect.
// By default, golang will copy all of the original request's headers on redirect,
// unless they're sensitive, like "Authorization" or "Www-Authenticate". Only send
// sensitive ones to the same origin, or subdomains thereof (https://go-review.googlesource.com/c/go/+/28930/)
// Check discussion: https://github.com/golang/go/issues/4800
// For example:
//
//	client.SetRedirectPolicy(req.AlwaysCopyHeaderRedirectPolicy("Authorization"))
func AlwaysCopyHeaderRedirectPolicy(headers ...string) RedirectPolicy {
	return func(req *http.Request, via []*http.Request) error {
		for _, header := range headers {
			if len(req.Header.Values(header)) > 0 {
				continue
			}
			vals := via[0].Header.Values(header)
			for _, val := range vals {
				req.Header.Add(header, val)
			}
		}
		return nil
	}
}
