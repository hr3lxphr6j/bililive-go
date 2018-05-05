package api

import (
	"net/url"
	"testing"
)

const twitchTestUrl = "https://www.twitch.tv/wuyikoei"

func TestTwitchLive_GetRoom(t *testing.T) {
	u, _ := url.Parse(twitchTestUrl)
	t.Log((&TwitchLive{Url: u}).GetRoom())
}

func TestTwitchLive_GetUrls(t *testing.T) {
	u, _ := url.Parse(twitchTestUrl)
	t.Log((&TwitchLive{Url: u}).GetUrls())
}
