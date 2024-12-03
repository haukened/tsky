package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/haukened/tsky/internal/messages"
	"github.com/haukened/tsky/internal/tokensvc"
	"github.com/haukened/tsky/internal/utils"
)

const PROFILE_URI_BASE = "https://%s/xrpc/app.bsky.actor.getProfile?actor=%s"

var (
	ErrNoJWT = fmt.Errorf("no JWT provided")
)

type ProfileTab struct {
	name     string
	loaded   bool
	Profile  messages.ProfileMessage
	Did      string
	server   string
	jwt      *tokensvc.Refresher
	TabIndex int
}

func (p ProfileTab) Name() string {
	return p.name
}

func (p ProfileTab) Server() string {
	return fmt.Sprintf(PROFILE_URI_BASE, p.server, p.Did)
}

func NewProfileTab(did, server string, jwt *tokensvc.Refresher) ProfileTab {
	return ProfileTab{
		name:     "Profile",
		Did:      did,
		server:   server,
		jwt:      jwt,
		TabIndex: 0,
		loaded:   false,
	}
}

func (p ProfileTab) Init() tea.Cmd {
	var msg messages.ProfileMessage
	if p.jwt == nil {
		msg.LoadingError = true
		msg.Error = ErrNoJWT
		return messages.SendProfileMsg(msg)
	}
	err := utils.HTTPGetAndParse(p.Server(), p.jwt.AuthToken(), &msg)
	if err != nil {
		msg.LoadingError = true
		msg.Error = err
	}
	return messages.SendProfileMsg(msg)
}

func (p ProfileTab) Update(msg tea.Msg) (NamedModel, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.ProfileMessage:
		p.loaded = true
		p.Profile = msg
	}
	return p, nil
}

func (p ProfileTab) View() string {
	if !p.loaded {
		return "Loading..."
	}
	if p.Profile.LoadingError {
		return fmt.Sprintf("Error: %s contacting %s", p.Profile.Error, p.Server())
	}
	return fmt.Sprintf("%+v", p)
}
