package tokensvc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/haukened/tsky/internal/utils"
)

const REFRESH_URI_BASE = "https://%s/xrpc/com.atproto.server.refreshSession"

var (
	ErrHttpError            = errors.New("http error in underlying refresher client")
	ErrUnableToRefreshToken = errors.New("unable to refresh token")
)

type Refresher struct {
	authToken    string
	refreshToken string
	server       string
}

type RefreshOutput struct {
	AccessJwt  string `json:"accessJwt"`
	RefreshJwt string `json:"refreshJwt"`
}

func NewRefresher(server, refreshToken string) (*Refresher, error) {
	r := &Refresher{
		refreshToken: refreshToken,
		server:       server,
	}
	// refresh now
	err := r.Refresh()
	if err != nil {
		return nil, err
	}
	// get the expiration time of the token
	exp := utils.GetTokenExpiration(r.authToken)
	// refresh 5 minutes before expiration
	early := exp.Add(-5 * time.Minute)
	// set a timer to refresh at that time
	time.AfterFunc(time.Until(early), func() {
		r.Refresh()
	})
	// return the refresher
	return r, nil
}

func (r *Refresher) AuthToken() string {
	if utils.IsJwtExpired(r.authToken) {
		r.Refresh()
	}
	return r.authToken
}

func (r *Refresher) Refresh() error {
	URL := fmt.Sprintf(REFRESH_URI_BASE, r.server)
	req, err := http.NewRequest("POST", URL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", r.refreshToken))
	req.Header.Set("User-Agent", utils.UserAgent())
	req.Header.Set("Content-Type", "application/json")
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return ErrHttpError
	}
	if resp.StatusCode != http.StatusOK {
		return ErrUnableToRefreshToken
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var output RefreshOutput
	err = json.Unmarshal(body, &output)
	if err != nil {
		return err
	}
	r.authToken = output.AccessJwt
	r.refreshToken = output.RefreshJwt
	return nil
}

func (r *Refresher) RefreshToken() string {
	return r.refreshToken
}
