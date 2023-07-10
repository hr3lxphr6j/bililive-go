package altsvcutil

import (
	"bytes"
	"fmt"
	"github.com/imroc/req/v3/internal/netutil"
	"github.com/imroc/req/v3/pkg/altsvc"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type altAvcParser struct {
	*bytes.Buffer
}

// validOptionalPort reports whether port is either an empty string
// or matches /^:\d*$/
func validOptionalPort(port string) bool {
	if port == "" {
		return true
	}
	if port[0] != ':' {
		return false
	}
	for _, b := range port[1:] {
		if b < '0' || b > '9' {
			return false
		}
	}
	return true
}

// splitHostPort separates host and port. If the port is not valid, it returns
// the entire input as host, and it doesn't check the validity of the host.
// Unlike net.SplitHostPort, but per RFC 3986, it requires ports to be numeric.
func splitHostPort(hostPort string) (host, port string) {
	host = hostPort

	colon := strings.LastIndexByte(host, ':')
	if colon != -1 && validOptionalPort(host[colon:]) {
		host, port = host[:colon], host[colon+1:]
	}

	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		host = host[1 : len(host)-1]
	}
	return
}

// ParseHeader parses the AltSvc from header value.
func ParseHeader(value string) ([]*altsvc.AltSvc, error) {
	p := newAltSvcParser(value)
	return p.Parse()
}

func newAltSvcParser(value string) *altAvcParser {
	buf := bytes.NewBufferString(value)
	return &altAvcParser{buf}
}

var endOfTime = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)

func (p *altAvcParser) Parse() (as []*altsvc.AltSvc, err error) {
	for {
		a, e := p.parseOne()
		if a != nil {
			as = append(as, a)
		}
		if e != nil {
			if e == io.EOF {
				return
			} else {
				err = e
				return
			}
		}
	}
	return
}

func (p *altAvcParser) parseKv() (key, value string, haveNextField bool, err error) {
	line, err := p.ReadBytes('=')
	if len(line) == 0 {
		return
	}
	key = strings.TrimSpace(string(line[:len(line)-1]))
	bs := p.Bytes()
	if len(bs) == 0 {
		err = io.EOF
		return
	}
	if bs[0] == '"' {
		quoteIndex := 0
		for i := 1; i < len(bs); i++ {
			if bs[i] == '"' {
				quoteIndex = i
				break
			}
		}
		if quoteIndex == 0 {
			err = fmt.Errorf("quote in alt-svc is not complete: %s", bs)
			return
		}
		value = string(bs[1:quoteIndex])
		p.Next(quoteIndex + 1)
		if len(bs) == quoteIndex+1 {
			err = io.EOF
			return
		}
		var b byte
		b, err = p.ReadByte()
		if err != nil {
			return
		}
		if b == ';' {
			haveNextField = true
		}
	} else {
		delimIndex := 0
	LOOP:
		for i, v := range bs {
			switch v {
			case ',':
				delimIndex = i
				break LOOP
			case ';':
				delimIndex = i
				haveNextField = true
				break LOOP
			}
		}
		if delimIndex == 0 {
			err = io.EOF
			value = strings.TrimSpace(string(bs))
			return
		}
		p.Next(delimIndex + 1)
		value = string(bs[:delimIndex])
	}
	return
}

func (p *altAvcParser) parseOne() (as *altsvc.AltSvc, err error) {
	proto, addr, haveNextField, err := p.parseKv()
	if proto == "" || addr == "" {
		return
	}
	host, port := splitHostPort(addr)

	as = &altsvc.AltSvc{
		Protocol: proto,
		Host:     host,
		Port:     port,
		Expire:   endOfTime,
	}

	if !haveNextField {
		return
	}

	key, ma, haveNextField, err := p.parseKv()
	if key == "" || ma == "" {
		return
	}
	if key != "ma" {
		err = fmt.Errorf("expect ma field, got %s", key)
		return
	}

	maInt, err := strconv.ParseInt(ma, 10, 64)
	if err != nil {
		return
	}
	as.Expire = time.Now().Add(time.Duration(maInt) * time.Second)

	if !haveNextField {
		return
	}

	// drain useless fields
	for {
		_, _, haveNextField, err = p.parseKv()
		if haveNextField {
			continue
		} else {
			break
		}
	}
	return
}

// ConvertURL converts the raw request url to expected alt-svc's url.
func ConvertURL(a *altsvc.AltSvc, u *url.URL) *url.URL {
	host, port := netutil.AuthorityHostPort(u.Scheme, u.Host)
	uu := *u
	modify := false
	if a.Host != "" && a.Host != host {
		host = a.Host
		modify = true
	}
	if a.Port != "" && a.Port != port {
		port = a.Port
		modify = true
	}
	if modify {
		uu.Host = net.JoinHostPort(host, port)
	}
	return &uu
}
