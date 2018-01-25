// +build !go1.1 !go1.2

package request

import (
	"net"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/proxy"
)

func applyProxy(a *Args) (err error) {
	if a.Proxy == "" {
		return nil
	}

	u, err := url.Parse(a.Proxy)
	if err != nil {
		return err
	}
	switch u.Scheme {
	case "http", "https":
		a.Client.Transport = &http.Transport{
			Proxy: http.ProxyURL(u),
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 10 * time.Second,
		}
	case "socks5":
		dialer, err := proxy.FromURL(u, proxy.Direct)
		if err != nil {
			return err
		}
		a.Client.Transport = &http.Transport{
			Proxy:               http.ProxyFromEnvironment,
			Dial:                dialer.Dial,
			TLSHandshakeTimeout: 10 * time.Second,
		}
	}
	return
}
