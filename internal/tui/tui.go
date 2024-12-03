package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/haukened/tsky/internal/config"
	"github.com/haukened/tsky/internal/debug"
	"github.com/haukened/tsky/internal/messages"
)

var (
	skyBlue = lipgloss.Color("#2081FE")
)

type Model struct {
	conf         *config.Config
	models       []NamedModel
	currentModel int
	h            int
	w            int
	statusMsg    string
	helpMsg      string
}

func NewModel(c *config.Config) Model {
	return Model{
		conf: c,
		models: []NamedModel{
			NewSplashModel(1),
			NewLoginModel(c),
			NewAuthModel(c),
			NewAppView(c),
		},
		currentModel: 0,
		statusMsg:    "",
		helpMsg:      "",
	}
}

func (m Model) Init() tea.Cmd {
	// initialize all the models
	var cmds []tea.Cmd
	for _, model := range m.models {
		cmd := model.Init()
		cmds = append(cmds, cmd)
	}
	return tea.Batch(cmds...)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.w = msg.Width - 2
		m.h = msg.Height - 1
		return m, nil
	case messages.StatusMsg:
		m.statusMsg = string(msg)
		return m, nil
	case messages.HelpMsg:
		m.helpMsg = string(msg)
		return m, nil
	case messages.NextMsg:
		if m.currentModel < len(m.models)-1 {
			debug.Debugf("Advancing from %s to %s", m.models[m.currentModel].Name(), m.models[m.currentModel+1].Name())
			// reset the help message
			m.helpMsg = ""
			// get the next model
			nextModel := m.models[m.currentModel+1]
			// initialize the model
			cmd = nextModel.Init()
			cmds = append(cmds, cmd)
			// update the current model
			m.currentModel++
		}
		return m, tea.Batch(cmds...)
	case messages.PrevMsg:
		if m.currentModel > 0 {
			debug.Debugf("Regressing from %s to %s", m.models[m.currentModel].Name(), m.models[m.currentModel-1].Name())
			// reset the help message
			m.helpMsg = ""
			// get the previous model
			prevModel := m.models[m.currentModel-1]
			switch prevModel.(type) {
			case LoginModel:
				prevModel = NewLoginModel(m.conf)
			case AuthModel:
				prevModel = NewAuthModel(m.conf)
			}
			// initialize the model
			cmd = prevModel.Init()
			cmds = append(cmds, cmd)
			// replace the current model
			m.models[m.currentModel] = prevModel
			// update the current model
			m.currentModel--
		}
		return m, tea.Batch(cmds...)
	}

	// update the current model
	model, cmd := m.models[m.currentModel].Update(msg)
	cmds = append(cmds, cmd)

	// sync the model back
	m.models[m.currentModel] = model

	// Return the updated model and any commands
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
		Padding(0, 1).
		Align(lipgloss.Center, lipgloss.Center).
		Render(s)
	doc.WriteString(mainContent + "\n")
	doc.WriteString(m.MkFooter())
	return doc.String()
}

func (m Model) MkFooter() string {
	dimensions := fmt.Sprintf("%dx%d", m.w, m.h)
	borderStyle := lipgloss.NewStyle().Foreground(skyBlue)
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#5e5e5e")).Bold(true)
	statusStyle := lipgloss.NewStyle().Bold(true)
	wS, _ := lipgloss.Size(m.statusMsg)
	wH, _ := lipgloss.Size(m.helpMsg)
	wD, _ := lipgloss.Size(dimensions)
	w := m.w - wS - wH - wD - 8
	sB := strings.Builder{}
	sB.WriteString(borderStyle.Render("╰-"))
	sB.WriteString(borderStyle.Bold(true).Render("tSky-"))
	sB.WriteString(helpStyle.Render(m.helpMsg))
	for i := 0; i < w; i++ {
		sB.WriteString(borderStyle.Render("─"))
	}
	sB.WriteString(statusStyle.Render(m.statusMsg))
	sB.WriteString(borderStyle.Render("─"))
	sB.WriteString(borderStyle.Render(dimensions))
	sB.WriteString(borderStyle.Render("-╯"))
	return sB.String()
}
