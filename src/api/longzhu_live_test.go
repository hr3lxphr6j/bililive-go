package api

import (
	"net/url"
	"testing"
)

const longzhuTestUrl = "http://star.longzhu.com/777777"

func TestLongzhuLive_GetInfo(t *testing.T) {
	u, _ := url.Parse(longzhuTestUrl)
	t.Log(NewLive(u).GetInfo())
}

func TestLongzhuLive_GetStreamUrls(t *testing.T) {
	u, _ := url.Parse(longzhuTestUrl)
	t.Log(NewLive(u).GetStreamUrls())
}
