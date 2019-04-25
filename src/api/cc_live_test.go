package api

import (
	"net/url"
	"testing"
)

const ccTestUrl = "https://cc.163.com/361433"

func TestCCLive_GetInfo(t *testing.T) {
	u, _ := url.Parse(ccTestUrl)
	live, _ := NewLive(u)
	t.Log(live.GetInfo())
}

func TestCCLive_GetStreamUrls(t *testing.T) {
	u, _ := url.Parse(ccTestUrl)
	live, _ := NewLive(u)
	t.Log(live.GetStreamUrls())
}
