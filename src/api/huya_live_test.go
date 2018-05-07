package api

import (
	"net/url"
	"testing"
)

const huyaTestUrl = "https://www.huya.com/dongxiaosa"

func TestHuYaLive_GetRoom(t *testing.T) {
	u, _ := url.Parse(huyaTestUrl)
	t.Log((&HuYaLive{Url: u}).GetRoom())
}

func TestHuYaLive_GetUrls(t *testing.T) {
	u, _ := url.Parse(huyaTestUrl)
	t.Log((&HuYaLive{Url: u}).GetUrls())
}
