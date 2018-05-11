package api

import (
	"net/url"
	"testing"
)

const pandaTestUrl = "https://www.panda.tv/1865618"

func TestPandaLive_GetInfo(t *testing.T) {
	u, _ := url.Parse(pandaTestUrl)
	t.Log(NewLive(u).GetInfo())
}

func TestPandaLive_GetStreamUrls(t *testing.T) {
	u, _ := url.Parse(pandaTestUrl)
	t.Log(NewLive(u).GetStreamUrls())
}
