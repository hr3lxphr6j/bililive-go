package api

import (
	"net/url"
	"testing"
)

const douyuTestUrl = "https://www.douyu.com/6655"

func TestDouyuLive_GetInfo(t *testing.T) {
	u, _ := url.Parse(douyuTestUrl)
	live, _ := NewLive(u)
	t.Log(live.GetInfo())
}

func TestDouyuLive_GetStreamUrls(t *testing.T) {
	u, _ := url.Parse(douyuTestUrl)
	live, _ := NewLive(u)
	t.Log(live.GetStreamUrls())
}
