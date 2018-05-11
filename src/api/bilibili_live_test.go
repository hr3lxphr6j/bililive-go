package api

import (
	"net/url"
	"testing"
)

const biliBiliTestUrl = "https://live.bilibili.com/161"

func TestBiliBiliLive_GetInfo(t *testing.T) {
	u, _ := url.Parse(biliBiliTestUrl)
	t.Log(NewLive(u).GetInfo())
}

func TestBiliBiliLive_GetStreamUrls(t *testing.T) {
	u, _ := url.Parse(biliBiliTestUrl)
	t.Log(NewLive(u).GetStreamUrls())
}
