package api

import (
	"net/url"
	"testing"
)

const longzhuTestUrl = "http://star.longzhu.com/777777"

func TestLongzhuLive_GetRoom(t *testing.T) {
	u, _ := url.Parse(longzhuTestUrl)
	t.Log((&LongzhuLive{Url: u}).GetRoom())
}

func TestLongzhuLive_GetUrls(t *testing.T) {
	u, _ := url.Parse(longzhuTestUrl)
	t.Log((&LongzhuLive{Url: u}).GetUrls())
}
