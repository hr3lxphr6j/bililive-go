package api

import (
	"net/url"
	"testing"
)

const (
	qqTestUrl = "https://egame.qq.com/497383565"
)

func TestQQ_GetInfo(t *testing.T) {
	u, _ := url.Parse(qqTestUrl)
	live, _ := NewLive(u)
	t.Log(live.GetInfo())
}

func TestQQ_GetStreamUrls(t *testing.T) {
	u, _ := url.Parse(qqTestUrl)
	live, _ := NewLive(u)
	t.Log(live.GetStreamUrls())
}
