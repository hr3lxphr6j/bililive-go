package api

import (
	"net/url"
	"testing"
)

const huyaTestUrl = "https://www.huya.com/dongxiaosa"

func TestHuYaLive_GetInfo(t *testing.T) {
	u, _ := url.Parse(huyaTestUrl)
	t.Log(NewLive(u).GetInfo())
}

func TestHuYaLive_GetStreamUrls(t *testing.T) {
	u, _ := url.Parse(huyaTestUrl)
	t.Log(NewLive(u).GetStreamUrls())
}
