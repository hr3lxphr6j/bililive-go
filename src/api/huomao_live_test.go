package api

import (
	"net/url"
	"testing"
)

const (
	huomaoTestUrl  = "https://www.huomao.com/6710"   // 正常直播间
	huomaoTestUrl2 = "https://www.huomao.com/954927" // 娱乐直播间
)

func TestHuoMaoLive_GetInfo(t *testing.T) {
	u, _ := url.Parse(huomaoTestUrl2)
	live, _ := NewLive(u)
	t.Log(live.GetInfo())
}

func TestHuoMaoLive_GetStreamUrls(t *testing.T) {
	u, _ := url.Parse(huomaoTestUrl)
	live, _ := NewLive(u)
	t.Log(live.GetStreamUrls())
}
