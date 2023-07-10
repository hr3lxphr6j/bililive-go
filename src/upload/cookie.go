package upload

type CookieInfo struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	Ttl     int64  `json:"ttl"`
	Data    struct {
		IsNew        bool   `json:"is_new"`
		Mid          int64  `json:"mid"`
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
		TokenInfo    struct {
			Mid          int64  `json:"mid"`
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
			ExpiresIn    int64  `json:"expires_in"`
		} `json:"token_info"`
		CookieInfo struct {
			Cookies []struct {
				Name     string `json:"name"`
				Value    string `json:"value"`
				HttpOnly int64  `json:"http_only"`
				Expires  int64  `json:"expires"`
				Secure   int64  `json:"secure"`
			} `json:"cookies"`
			Domains []string `json:"domains"`
		} `json:"cookie_info"`
		Sso []string `json:"sso"`
	} `json:"data"`
}
