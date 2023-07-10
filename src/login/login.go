package login

import "net/http"

type Loginer interface {
	Login()
}

type LoginClient struct {
	client *http.Client
}
