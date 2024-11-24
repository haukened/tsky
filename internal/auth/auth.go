package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/haukened/tsky/internal/config"
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
	postUrl := fmt.Sprintf(URI_BASE, c.Server)
	b, err := json.Marshal(Request{
		Identifier: c.Username,
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
	req.Header.Set("User-Agent", c.UserAgent)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to authenticate user: %s", resp.Status)
	}
	defer resp.Body.Close()
	var response Response
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return err
	}
	fmt.Printf("Response: %+v\n", response)
	return nil
}
