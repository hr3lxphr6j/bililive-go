package api

import (
	"net/url"
	"testing"
)

const yizhiboTestUrl = "https://www.yizhibo.com/l/ytbdVP1SSmWXzUx_.html"

func TestYiZhiBoLive_GetInfo(t *testing.T) {
	u, _ := url.Parse(yizhiboTestUrl)
	live, _ := NewLive(u)
	t.Log(live.GetInfo())
}

func TestYiZhiBoLive_GetStreamUrls(t *testing.T) {
	u, _ := url.Parse(yizhiboTestUrl)
	live, _ := NewLive(u)
	t.Log(live.GetStreamUrls())
}
