package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/haukened/tsky/internal/config"
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
	case splashMsg:
		m.currentModel = "login"
		return m, nil
	case helpMsg:
		m.helpMsg = string(msg)
		return m, nil
	case statusMsg:
		m.statusMsg = string(msg)
		return m, nil
	case authFinishedMsg:
		// Switch to the auth model
		m.models["auth"] = NewAuthModel(m.conf)
		// Init the auth model
		cmd := m.models["auth"].Init()
		// Set the current model to auth
		m.currentModel = "auth"
		// remove the login model
		delete(m.models, "login")
		return m, cmd
	}

	// Update all models
	var cmds []tea.Cmd
	for i := range m.models {
		newModel, cmd := m.models[i].Update(msg)
		m.models[i] = newModel
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	return m.Render(m.models[m.currentModel].View())
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
