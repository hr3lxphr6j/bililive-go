package api

import (
	"net/url"
	"testing"
)

const huyaTestUrl = "https://www.huya.com/wexiamo"

func TestHuYaLive_GetInfo(t *testing.T) {
	u, _ := url.Parse(huyaTestUrl)
	live, _ := NewLive(u)
	t.Log(live.GetInfo())
}

func TestHuYaLive_GetStreamUrls(t *testing.T) {
	u, _ := url.Parse(huyaTestUrl)
	live, _ := NewLive(u)
	info, _ := live.GetInfo()
	if info.Status {
		t.Log(live.GetStreamUrls())
	}
}
