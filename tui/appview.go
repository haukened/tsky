package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type AppView struct {
	spinner spinner.Model
}

func NewAppView() *AppView {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(skyBlue)
	return &AppView{
		spinner: s,
	}
}

func (a *AppView) Init() tea.Cmd {
	return a.spinner.Tick
}

func (a *AppView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case spinner.TickMsg:
		a.spinner, cmd = a.spinner.Update(msg)
		return a, cmd
	}
	return a, nil
}

func (a *AppView) View() string {
	return fmt.Sprintf("Login Successful\n%s %s", a.spinner.View(), "Loading...")
}
