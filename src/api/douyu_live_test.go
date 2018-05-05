package api

import (
	"net/url"
	"testing"
)

const douyuTestUrl = "https://www.douyu.com/6655"

func TestDouyuLive_GetRoom(t *testing.T) {
	u, _ := url.Parse(douyuTestUrl)
	t.Log((&DouyuLive{Url: u}).GetRoom())
}

func TestDouyuLive_GetUrl(t *testing.T) {
	u, _ := url.Parse(douyuTestUrl)
	t.Log((&DouyuLive{Url: u}).GetUrls())
}
