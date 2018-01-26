package api

import (
	"testing"
	"net/url"
)

const zhanQiTestUrl = "https://www.zhanqi.tv/12qaq"

func TestZhanQiLive_GetRoom(t *testing.T) {
	u, _ := url.Parse(zhanQiTestUrl)
	t.Log((&ZhanQiLive{u}).GetRoom())
}

func TestZhanQiLive_GetUrls(t *testing.T) {
	u, _ := url.Parse(zhanQiTestUrl)
	t.Log((&ZhanQiLive{u}).GetUrls())
}
