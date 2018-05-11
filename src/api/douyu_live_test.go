package api

import (
	"net/url"
	"testing"
)

const douyuTestUrl = "https://www.douyu.com/6655"

func TestDouyuLive_GetInfo(t *testing.T) {
	u, _ := url.Parse(douyuTestUrl)
	t.Log(NewLive(u).GetInfo())
}

func TestDouyuLive_GetStreamUrls(t *testing.T) {
	u, _ := url.Parse(douyuTestUrl)
	t.Log(NewLive(u).GetStreamUrls())
}
