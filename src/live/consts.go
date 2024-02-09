package live

import (
	"github.com/hr3lxphr6j/requests"
)

const (
	userAgent        = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Safari/537.36"
	androidBiliAgent = "Bilibili Freedoooooom/MarkII BiliDroid/5.49.0 os/android model/MuMu mobi_app/android build/5490400 channel/dw090 innerVer/5490400 osVer/6.0.1 network/2"
)

var CommonUserAgent = requests.UserAgent(userAgent)
var AndroidBiliAgent = requests.UserAgent(androidBiliAgent)
