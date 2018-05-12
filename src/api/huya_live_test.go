package api

import (
	"net/url"
	"testing"
)

const huyaTestUrl = "https://www.huya.com/dongxiaosa"

func TestHuYaLive_GetInfo(t *testing.T) {
	u, _ := url.Parse(huyaTestUrl)
	live, _ := NewLive(u)
	t.Log(live.GetInfo())
}

func TestHuYaLive_GetStreamUrls(t *testing.T) {
	u, _ := url.Parse(huyaTestUrl)
	live, _ := NewLive(u)
	t.Log(live.GetStreamUrls())
}
