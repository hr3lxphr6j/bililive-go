package netutil

import (
	"golang.org/x/net/idna"
	"net"
	"net/url"
	"strings"
)

func AuthorityKey(u *url.URL) string {
	return u.Scheme + "://" + AuthorityAddr(u.Scheme, u.Host)
}

// AuthorityAddr returns a given authority (a host/IP, or host:port / ip:port)
// and returns a host:port. The port 443 is added if needed.
func AuthorityAddr(scheme, authority string) (addr string) {
	host, port := AuthorityHostPort(scheme, authority)
	// IPv6 address literal, without a port:
	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		return host + ":" + port
	}
	addr = net.JoinHostPort(host, port)
	return
}

func AuthorityHostPort(scheme, authority string) (host, port string) {
	host, port, err := net.SplitHostPort(authority)
	if err != nil { // authority didn't have a port
		port = "443"
		if scheme == "http" {
			port = "80"
		}
		host = authority
	}
	if a, err := idna.ToASCII(host); err == nil {
		host = a
	}
	return
}
