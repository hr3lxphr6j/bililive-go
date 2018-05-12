package api

import (
	"net/url"
	"testing"
)

const huomaoTestUrl = "https://www.huomao.com/762719"

func TestHuoMaoLive_GetInfo(t *testing.T) {
	u, _ := url.Parse(huomaoTestUrl)
	live, _ := NewLive(u)
	t.Log(live.GetInfo())
}

func TestHuoMaoLive_GetStreamUrls(t *testing.T) {
	u, _ := url.Parse(huomaoTestUrl)
	live, _ := NewLive(u)
	t.Log(live.GetStreamUrls())
}
