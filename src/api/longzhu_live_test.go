package api

import (
	"net/url"
	"testing"
)

const longzhuTestUrl = "http://star.longzhu.com/777777"

func TestLongzhuLive_GetInfo(t *testing.T) {
	u, _ := url.Parse(longzhuTestUrl)
	live, _ := NewLive(u)
	t.Log(live.GetInfo())
}

func TestLongzhuLive_GetStreamUrls(t *testing.T) {
	u, _ := url.Parse(longzhuTestUrl)
	live, _ := NewLive(u)
	t.Log(live.GetStreamUrls())
}
