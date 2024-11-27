package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/haukened/tsky/internal/config"
	"github.com/haukened/tsky/internal/utils"
)

const BASE_AUTH_URI = "https://%s/xrpc/com.atproto.server.createSession"

type RequestBody struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

type AuthResponse struct {
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

func LoginWithPassword(c *config.Config) (err error) {
	// Create the URL
	postURL := fmt.Sprintf(BASE_AUTH_URI, c.Server)

	// Create the body
	body := RequestBody{
		Identifier: c.Identifier,
		Password:   c.AppPassword,
	}

	// Marshal the body
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return
	}

	// Create the request
	req, err := http.NewRequest("POST", postURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return
	}

	// Set the headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", utils.UserAgent())

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("failed to login, status code: %d", resp.StatusCode)
		return
	}

	// read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	// Unmarshal the response
	var authResponse AuthResponse
	err = json.Unmarshal(bodyBytes, &authResponse)
	if err != nil {
		return
	}

	// set the access and refresh tokens
	c.AccessJwt = authResponse.AccessJwt
	c.RefreshJwt = authResponse.RefreshJwt

	return
}
