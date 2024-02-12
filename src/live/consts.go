package live

import (
	"github.com/hr3lxphr6j/requests"
)

const (
	userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Safari/537.36"
)

var CommonUserAgent = requests.UserAgent(userAgent)
