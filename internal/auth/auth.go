package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/haukened/tsky/internal/config"
	loginform "github.com/haukened/tsky/internal/loginForm"
	"github.com/haukened/tsky/internal/utils"
)

const URI_BASE = "https://%s/xrpc/com.atproto.server.createSession"

type Request struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

type Response struct {
	Did    string `json:"did"`
	DidDoc struct {
		Context            []string `json:"@context"`
		ID                 string   `json:"id"`
		AlsoKnownAs        []string `json:"alsoKnownAs"`
		VerificationMethod []struct {
			ID                 string `json:"id"`
			Type               string `json:"type"`
			Controller         string `json:"controller"`
			PublicKeyMultibase string `json:"publicKeyMultibase"`
		} `json:"verificationMethod"`
		Service []struct {
			ID              string `json:"id"`
			Type            string `json:"type"`
			ServiceEndpoint string `json:"serviceEndpoint"`
		} `json:"service"`
	} `json:"didDoc"`
	Handle          string `json:"handle"`
	Email           string `json:"email"`
	EmailConfirmed  bool   `json:"emailConfirmed"`
	EmailAuthFactor bool   `json:"emailAuthFactor"`
	AccessJwt       string `json:"accessJwt"`
	RefreshJwt      string `json:"refreshJwt"`
	Active          bool   `json:"active"`
}

func AuthUser(c *config.Config) error {
	err := getCredentials(c)
	if err != nil {
		return err
	}
	return nil
}

func passwordAuth(c *config.Config) error {
	postUrl := fmt.Sprintf(URI_BASE, c.Server)
	b, err := json.Marshal(Request{
		Identifier: c.Identifier,
		Password:   c.AppPassword,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, postUrl, bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("User-Agent", utils.UserAgent())
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to authenticate user: %s", resp.Status)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}
	// set the access and refresh tokens
	c.AccessJwt = response.AccessJwt
	c.RefreshJwt = response.RefreshJwt
	// clear the password
	c.AppPassword = ""
	return nil
}

func getCredentials(c *config.Config) error {
	// first check if we have a valid identifier
	if c.Identifier == "" {
		// if we don't we know we need to prompt for credentials
		err := doPasswordLogin(c)
		if err != nil {
			return err
		}
	}
	// next we need to check if we have a valid refresh token
	if c.RefreshJwt == "" {
		// if we don't we need to prompt for credentials
		err := doPasswordLogin(c)
		if err != nil {
			return err
		}
	} else {
		// we need to check if the refresh token is still valid
		if utils.IsJwtExpired(c.RefreshJwt) {
			err := doPasswordLogin(c)
			if err != nil {
				return err
			}
		}
	}
	// otherwise we can use the refreshtoken later
	return nil
}

func doPasswordLogin(c *config.Config) error {
	err := loginform.Show(c)
	if err != nil {
		return err
	}
	return passwordAuth(c)
}
