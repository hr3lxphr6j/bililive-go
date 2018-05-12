package api

import (
	"net/url"
	"testing"
)

const pandaTestUrl = "https://www.panda.tv/1909193"

func TestPandaLive_GetInfo(t *testing.T) {
	u, _ := url.Parse(pandaTestUrl)
	live, _ := NewLive(u)
	t.Log(live.GetInfo())
}

func TestPandaLive_GetStreamUrls(t *testing.T) {
	u, _ := url.Parse(pandaTestUrl)
	live, _ := NewLive(u)
	t.Log(live.GetStreamUrls())
}
