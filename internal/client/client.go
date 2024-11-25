package client

import (
	"fmt"
	"net/http"

	"github.com/haukened/tsky/internal/tokensvc"
	"github.com/haukened/tsky/internal/utils"
)

type Client struct {
	tokSvc     *tokensvc.Refresher
	httpClient *http.Client
}

func New(tokSvc *tokensvc.Refresher) *Client {
	return &Client{
		tokSvc:     tokSvc,
		httpClient: &http.Client{},
	}
}

func (c *Client) NewRequest(method, url string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.tokSvc.AuthToken()))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", utils.UserAgent())
	return req, nil
}
