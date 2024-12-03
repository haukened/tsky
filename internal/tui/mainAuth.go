package tui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/haukened/tsky/internal/config"
	"github.com/haukened/tsky/internal/debug"
	"github.com/haukened/tsky/internal/messages"
	"github.com/haukened/tsky/internal/utils"
)

type AuthModel struct {
	c *config.Config
	s spinner.Model
	r chan authResult
	m string
}

type authResult struct {
	Success bool
	Message string
}

func NewAuthModel(c *config.Config) AuthModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(skyBlue)
	return AuthModel{
		s: s,
		c: c,
		r: make(chan authResult),
		m: "Initializing",
	}
}

func (a AuthModel) Name() string {
	return "auth"
}

func (a AuthModel) Init() tea.Cmd {
	debug.Debugf("Initializing AuthModel")
	return a.s.Tick
}

func (a AuthModel) Update(msg tea.Msg) (NamedModel, tea.Cmd) {
	debug.Debugf("Updating AuthModel")
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case spinner.TickMsg:
		debug.Debugf("Auth Spinner Tick")
		a.s, cmd = a.s.Update(msg)
		cmds = append(cmds, cmd, a.s.Tick)
	case messages.StartAuthMsg:
		debug.Debugf("got start auth message")
		a.m = "Authenticating"
		go func(ch chan authResult) {
			doAuth(a.c, ch)
		}(a.r)
	}
	select {
	case result := <-a.r:
		if result.Success {
			a.m = "Success"
			cmds = append(cmds, messages.SendStatusMsg(result.Message))
			cmds = append(cmds, messages.Next)
		} else {
			a.m = "Failed"
			cmds = append(cmds, messages.Prev)
			cmds = append(cmds, messages.SendErrorMsg(result.Message))
		}
	default:
		// No Action
	}
	return a, tea.Batch(cmds...)
}

func (a AuthModel) View() string {
	return fmt.Sprintf("%s Authenicating as %s...\nStatus: %s", a.s.View(), a.c.Identifier, a.m)
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
	debug.Debugf("starting auth")
	if c.Identifier != "" && c.RefreshJwt != "" {
		debug.Debugf("Checking refresh token")
		if !utils.IsJwtExpired(c.RefreshJwt) {
			debug.Debugf("Refresh token is still valid")
			ch <- authResult{Success: true, Message: "Authenticated"}
			return
		} else {
			debug.Debugf("Refresh token is expired")
		}
	} else if c.Identifier == "" {
		debug.Debugf("No username provided")
		ch <- authResult{Success: false, Message: "No username provided"}
		return
	} else if c.AppPassword == "" {
		debug.Debugf("No password provided")
		ch <- authResult{Success: false, Message: "No password provided"}
		return
	}
	err := loginWithPassword(c)
	if err != nil {
		ch <- authResult{Success: false, Message: err.Error()}
		return
	}
	c.Save()
	ch <- authResult{Success: true, Message: "Authenticated"}
}

func loginWithPassword(c *config.Config) (err error) {
	debug.Debugf("Logging in with password")
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

	// log the request
	log.Printf("Request: %+v\n", req)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Response: %+v\n", resp)
		//lint:ignore ST1005 the capital is for Formatting
		err = fmt.Errorf("Login Failed: %s", resp.Status)
		// blank out the password
		c.AppPassword = ""
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
	c.Did = authResponse.Did

	// blank out the password
	c.AppPassword = ""

	return
}
