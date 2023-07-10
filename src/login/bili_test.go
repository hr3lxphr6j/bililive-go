package login

import (
	"net/http"
	"testing"
)

func TestBilibili_Login(t *testing.T) {
	var (
		authapi = "http://passport.bilibili.com/x/passport-tv-login/qrcode/auth_code"
		api     = "http://passport.bilibili.com/x/passport-tv-login/qrcode/poll"
	)

	b := NewBilibili(authapi, api, http.DefaultClient)
	b.Login()
}
