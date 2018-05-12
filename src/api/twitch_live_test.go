package api

import (
	"net/url"
	"testing"
)

const twitchTestUrl = "https://www.twitch.tv/wuyikoei"

func TestTwitchLive_GetInfo(t *testing.T) {
	u, _ := url.Parse(twitchTestUrl)
	live, _ := NewLive(u)
	t.Log(live.GetInfo())
}

func TestTwitchLive_GetStreamUrls(t *testing.T) {
	u, _ := url.Parse(twitchTestUrl)
	live, _ := NewLive(u)
	t.Log(live.GetStreamUrls())
}
