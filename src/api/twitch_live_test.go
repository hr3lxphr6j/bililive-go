package api

import (
	"net/url"
	"testing"
)

const twitchTestUrl = "https://www.twitch.tv/wuyikoei"

func TestTwitchLive_GetInfo(t *testing.T) {
	u, _ := url.Parse(twitchTestUrl)
	t.Log(NewLive(u).GetInfo())
}

func TestTwitchLive_GetStreamUrls(t *testing.T) {
	u, _ := url.Parse(twitchTestUrl)
	t.Log(NewLive(u).GetStreamUrls())
}
