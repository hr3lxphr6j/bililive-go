package api

import (
	"net/url"
	"testing"
)

const huomaoTestUrl = "https://www.huomao.com/762719"

func TestHuoMaoLive_GetRoom(t *testing.T) {
	u, _ := url.Parse(huomaoTestUrl)
	t.Log((&HuoMaoLive{Url: u}).GetRoom())
}

func TestHuoMaoLive_GetUrls(t *testing.T) {
	u, _ := url.Parse(huomaoTestUrl)
	t.Log((&HuoMaoLive{Url: u}).GetUrls())
}
