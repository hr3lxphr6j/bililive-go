package live

import (
	"testing"
	"net/url"
)

func TestPandaLive_GetRoom(t *testing.T) {
	u, _ := url.Parse("https://www.panda.tv/10300")
	(&PandaLive{u}).GetRoom()
}
