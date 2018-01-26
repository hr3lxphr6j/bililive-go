package api

import (
	"testing"
	"net/url"
)

const bilibiliTestUrl = "https://api.bilibili.com/1030"

func TestBiliBiliLive_GetRoom(t *testing.T) {
	u, _ := url.Parse(bilibiliTestUrl)
	t.Log((&BiliBiliLive{u}).GetRoom())
}

func TestBiliBiliLive_GetUrl(t *testing.T) {
	u, _ := url.Parse(bilibiliTestUrl)
	t.Log((&BiliBiliLive{u}).GetUrls())
}
