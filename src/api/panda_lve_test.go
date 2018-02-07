package api

import (
	"net/url"
	"testing"
)

const pandaTestUrl = "https://www.panda.tv/10027"

func TestPandaLive_GetRoom(t *testing.T) {
	u, _ := url.Parse(pandaTestUrl)
	t.Log((&PandaLive{u}).GetRoom())
}

func TestPandaLive_GetUrls(t *testing.T) {
	u, _ := url.Parse(pandaTestUrl)
	t.Log((&PandaLive{u}).GetUrls())
}
