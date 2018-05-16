package api

import (
	"net/url"
	"testing"
)

const openrecTestUrl = "https://www.openrec.tv/live/J84JHlBu1vT"

func TestOpenRecLive_GetInfo(t *testing.T) {
	u, _ := url.Parse(openrecTestUrl)
	live, _ := NewLive(u)
	t.Log(live.GetInfo())
}

func TestOpenRecLive_GetStreamUrls(t *testing.T) {
	u, _ := url.Parse(openrecTestUrl)
	live, _ := NewLive(u)
	t.Log(live.GetStreamUrls())
}
