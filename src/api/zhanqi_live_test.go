package api

import (
	"net/url"
	"testing"
)

const zhanQiTestUrl = "https://www.zhanqi.tv/12qaq"

func TestZhanQiLive_GetInfo(t *testing.T) {
	u, _ := url.Parse(zhanQiTestUrl)
	t.Log(NewLive(u).GetInfo())
}

func TestZhanQiLive_GetStreamUrls(t *testing.T) {
	u, _ := url.Parse(zhanQiTestUrl)
	t.Log(NewLive(u).GetStreamUrls())
}
