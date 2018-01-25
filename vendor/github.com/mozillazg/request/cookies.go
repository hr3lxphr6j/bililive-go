package request

import "net/http"

func applyCookies(a *Args, req *http.Request) {
	if a.Cookies == nil {
		return
	}
	cookies := a.Client.Jar.Cookies(req.URL)
	for k, v := range a.Cookies {
		cookies = append(cookies, &http.Cookie{Name: k, Value: v})
	}
	a.Client.Jar.SetCookies(req.URL, cookies)
}
