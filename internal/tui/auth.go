package tui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/haukened/tsky/internal/config"
	"github.com/haukened/tsky/internal/utils"
)

type AuthModel struct {
	c  *config.Config
	s  spinner.Model
	ch chan authResult
}

func NewAuthModel(c *config.Config) AuthModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(skyBlue)
	return AuthModel{
		s:  s,
		c:  c,
		ch: make(chan authResult),
	}
}

func (a AuthModel) Init() tea.Cmd {
	go func(ch chan authResult) {
		doAuth(a.c, ch)
	}(a.ch)
	return a.s.Tick
}

func (a AuthModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case spinner.TickMsg:
		a.s, cmd = a.s.Update(msg)
		cmds = append(cmds, cmd)
	}
	select {
	case result := <-a.ch:
		cmds = append(cmds, authStatus(result.success))
		if !result.success {
			cmds = append(cmds, SendStatusErr(result.message))
		}
	default:
	}
	return a, tea.Batch(cmds...)
}

func (a AuthModel) View() string {
	return fmt.Sprintf("%s Authenicating as %s...", a.s.View(), a.c.Identifier)
}

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

func doAuth(c *config.Config, ch chan authResult) {
	if c.Identifier != "" && c.RefreshJwt != "" {
		if !utils.IsJwtExpired(c.RefreshJwt) {
			ch <- authResult{true, "Already authenticated"}
			return
		}
	} else if c.Identifier == "" {
		ch <- authResult{false, "No username provided"}
		return
	} else if c.AppPassword == "" {
		ch <- authResult{false, "No password provided"}
		return
	}
	err := loginWithPassword(c)
	if err != nil {
		ch <- authResult{false, err.Error()}
		return
	}
	ch <- authResult{true, "Authenticated"}
}

func loginWithPassword(c *config.Config) (err error) {
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
		//lint:ignore ST1005 the capital is for Formatting
		err = fmt.Errorf("Login Failed: %s", resp.Status)
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
