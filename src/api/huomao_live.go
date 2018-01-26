package api

import "net/url"

type HuoMaoLive struct {
	Url *url.URL
}

func (h *HuoMaoLive) GetRoom() (*Info, error) {
	return nil, nil
}

func (h *HuoMaoLive) GetUrls() ([]*url.URL, error) {
	return nil, nil
}
