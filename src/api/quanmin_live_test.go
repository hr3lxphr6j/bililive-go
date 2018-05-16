package api

import (
	"net/url"
	"testing"
)

const quanminTestUrl = "https://www.quanmin.tv/8741269"

func TestQuanMinLive_GetInfo(t *testing.T) {
	u, _ := url.Parse(quanminTestUrl)
	live, _ := NewLive(u)
	t.Log(live.GetInfo())
}

func TestQuanMinLive_GetStreamUrls(t *testing.T) {
	u, _ := url.Parse(quanminTestUrl)
	live, _ := NewLive(u)
	t.Log(live.GetStreamUrls())
}
