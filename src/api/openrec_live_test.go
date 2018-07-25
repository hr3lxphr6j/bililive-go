package api

import (
	"net/url"
	"testing"
)

const openrecTestUrl = "https://www.openrec.tv/live/JEmT0qQP1BM"

func TestOpenRecLive_GetInfo(t *testing.T) {
	u, _ := url.Parse(openrecTestUrl)
	live, _ := NewLive(u)
	t.Log(live.GetInfo())
}

func TestOpenRecLive_GetStreamUrls(t *testing.T) {
	u, _ := url.Parse(openrecTestUrl)
	live, _ := NewLive(u)
	if live.GetCachedInfo().Status {
		t.Log(live.GetStreamUrls())
	}
}
