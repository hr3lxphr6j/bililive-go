package api

import (
	"net/url"
	"testing"
)

const biliBiliTestUrl = "https://live.bilibili.com/161"

func TestBiliBiliLive_GetInfo(t *testing.T) {
	u, _ := url.Parse(biliBiliTestUrl)
	live, _ := NewLive(u)
	t.Log(live.GetInfo())
}

func TestBiliBiliLive_GetStreamUrls(t *testing.T) {
	u, _ := url.Parse(biliBiliTestUrl)
	live, _ := NewLive(u)
	t.Log(live.GetStreamUrls())
}
