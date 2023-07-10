package altsvc

import (
	"sync"
	"time"
)

// AltSvcJar is default implementation of Jar, which stores
// AltSvc in memory.
type AltSvcJar struct {
	entries map[string]*AltSvc
	mu      sync.Mutex
}

// NewAltSvcJar create a AltSvcJar which implements Jar.
func NewAltSvcJar() *AltSvcJar {
	return &AltSvcJar{
		entries: make(map[string]*AltSvc),
	}
}

func (j *AltSvcJar) GetAltSvc(addr string) *AltSvc {
	if addr == "" {
		return nil
	}
	as, ok := j.entries[addr]
	if !ok {
		return nil
	}
	now := time.Now()
	j.mu.Lock()
	defer j.mu.Unlock()
	if as.Expire.Before(now) { // expired
		delete(j.entries, addr)
		return nil
	}
	return as
}

func (j *AltSvcJar) SetAltSvc(addr string, as *AltSvc) {
	if addr == "" {
		return
	}
	j.mu.Lock()
	defer j.mu.Unlock()
	j.entries[addr] = as
}

// AltSvc is the parsed alt-svc.
type AltSvc struct {
	// Protocol is the alt-svc proto, e.g. h3.
	Protocol string
	// Host is the alt-svc's host, could be empty if
	// it's the same host as the raw request.
	Host string
	// Port is the alt-svc's port.
	Port string
	// Expire is the time that the alt-svc should expire.
	Expire time.Time
}
