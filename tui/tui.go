package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/haukened/tsky/internal/config"
)

var (
	skyBlue = lipgloss.Color("#2081FE")
)

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type Model struct {
	conf         *config.Config
	models       []tea.Model
	currentModel int
	h            int
	w            int
}

func NewModel(c *config.Config) Model {
	return Model{
		conf: c,
		models: []tea.Model{
			splash{countdown: 30},
		},
		currentModel: 0,
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
		m.currentModel++
		return m, nil
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
	if m.currentModel == 0 {
		return m.Render(m.models[m.currentModel].View())
	}
	return m.Render(fmt.Sprintf("Hey!\nDimensions: %dx%d", m.w, m.h))
}

func (m Model) Render(s string) string {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(skyBlue).
		Height(m.h).
		Width(m.w).
		Padding(1, 2).
		Align(lipgloss.Center, lipgloss.Center).
		Render(s)
}
