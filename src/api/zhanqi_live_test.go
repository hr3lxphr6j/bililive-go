package api

import (
	"net/url"
	"testing"
)

const zhanQiTestUrl = "https://www.zhanqi.tv/12qaq"

func TestZhanQiLive_GetInfo(t *testing.T) {
	u, _ := url.Parse(zhanQiTestUrl)
	live, _ := NewLive(u)
	t.Log(live.GetInfo())
}

func TestZhanQiLive_GetStreamUrls(t *testing.T) {
	u, _ := url.Parse(zhanQiTestUrl)
	live, _ := NewLive(u)
	t.Log(live.GetStreamUrls())
}
