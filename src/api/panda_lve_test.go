package api

import (
	"testing"
	"net/url"
)

const pandaTestUrl = "https://www.panda.tv/10300"

func TestPandaLive_GetRoom(t *testing.T) {
	u, _ := url.Parse(pandaTestUrl)
	t.Log((&PandaLive{u}).GetRoom())
}

func TestBiliBiliLive_GetUrls(t *testing.T) {
	u, _ := url.Parse(pandaTestUrl)
	t.Log((&PandaLive{u}).GetUrls())
}
