package client

import (
	"net/http"

	tokensvc "github.com/haukened/tsky/internal/tokenSvc"
)

type Client struct {
	tokSvc     *tokensvc.Refresher
	httpClient *http.Client
	ua         string // user agent for requests
}

func New(tokSvc *tokensvc.Refresher, ua string) *Client {
	return &Client{
		tokSvc:     tokSvc,
		httpClient: &http.Client{},
		ua:         ua,
	}
}
