package api

import (
	"net/url"
	"testing"
)

const yizhiboTestUrl = "https://www.yizhibo.com/l/ytbdVP1SSmWXzUx_.html"

func TestYiZhiBoLive_GetRoom(t *testing.T) {
	u, _ := url.Parse(yizhiboTestUrl)
	t.Log((&YiZhiBoLive{Url: u}).GetRoom())
}

func TestYiZhiBoLive_GetUrls(t *testing.T) {
	u, _ := url.Parse(yizhiboTestUrl)
	t.Log((&YiZhiBoLive{Url: u}).GetUrls())
}
