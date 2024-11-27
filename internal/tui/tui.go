package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/haukened/tsky/internal/config"
	"github.com/haukened/tsky/internal/messages"
)

var (
	skyBlue = lipgloss.Color("#2081FE")
)

type Model struct {
	conf         *config.Config
	models       map[string]tea.Model
	currentModel string
	h            int
	w            int
	statusMsg    string
	helpMsg      string
}

func NewModel(c *config.Config) Model {
	login := NewFormModel(c)
	return Model{
		conf: c,
		models: map[string]tea.Model{
			"splash": splash{countdown: 2},
			"login":  login,
		},
		currentModel: "splash",
		statusMsg:    "",
		helpMsg:      "",
	}
}

func (m Model) Init() tea.Cmd {
	var cmds []tea.Cmd
	for _, model := range m.models {
		cmds = append(cmds, model.Init())
	}
	return tea.Batch(cmds...)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.w = msg.Width - 2
		m.h = msg.Height - 2
		return m, nil
	case messages.SplashMsg:
		m.currentModel = "login"
	case messages.HelpMsg:
		m.helpMsg = string(msg)
	case messages.StatusMsg:
		m.statusMsg = string(msg)
	case messages.LoginFinishedMsg:
		// Switch to the auth model
		m.models["auth"] = NewAuthModel(m.conf)
		// Init the auth model
		cmd := m.models["auth"].Init()
		// Set the current model to auth
		m.currentModel = "auth"
		cmds = append(cmds, cmd)
	case messages.AuthStatusMsg:
		if !msg {
			// clear the password
			m.conf.AppPassword = ""
			// init a new login model
			m.models["login"] = NewFormModel(m.conf)
			// init the login model
			cmd := m.models["login"].Init()
			// set the current model to login
			m.currentModel = "login"
			cmds = append(cmds, cmd)
		} else {
			// init the app view
			m.models["app"] = NewAppView(m.conf)
			// init the app view
			cmd := m.models["app"].Init()
			// set the current model to app
			m.currentModel = "app"
			cmds = append(cmds, cmd)
		}
	}
	// Update the current model
	current := m.models[m.currentModel]
	current, cmd := current.Update(msg)
	cmds = append(cmds, cmd)
	m.models[m.currentModel] = current

	// Return the updated model and any commands
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	current, ok := m.models[m.currentModel]
	if !ok {
		return fmt.Sprintf("Error: Model %s not found", m.currentModel)
	}
	return m.Render(current.View())
}

func (m Model) Render(s string) string {
	doc := strings.Builder{}
	mainContent := lipgloss.NewStyle().
		Height(m.h-1).
		Width(m.w).
		Border(lipgloss.RoundedBorder()).
		BorderBottom(false).
		BorderForeground(skyBlue).
		Padding(1, 2).
		Align(lipgloss.Center, lipgloss.Center).
		Render(s)
	doc.WriteString(mainContent + "\n")
	doc.WriteString(m.MkFooter())
	return doc.String()
}

func (m Model) MkFooter() string {
	borderStyle := lipgloss.NewStyle().Foreground(skyBlue)
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#5e5e5e")).Bold(true)
	statusStyle := lipgloss.NewStyle().Bold(true)
	wS, _ := lipgloss.Size(m.statusMsg)
	wH, _ := lipgloss.Size(m.helpMsg)
	w := m.w - wS - wH - 7
	sB := strings.Builder{}
	sB.WriteString(borderStyle.Render("╰-"))
	sB.WriteString(borderStyle.Bold(true).Render("tSky-"))
	sB.WriteString(helpStyle.Render(m.helpMsg))
	for i := 0; i < w; i++ {
		sB.WriteString(borderStyle.Render("─"))
	}
	sB.WriteString(statusStyle.Render(m.statusMsg))
	sB.WriteString(borderStyle.Render("-╯"))
	return sB.String()
}
