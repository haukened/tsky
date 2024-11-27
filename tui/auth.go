package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/haukened/tsky/internal/config"
)

type AuthModel struct {
	c *config.Config
	s spinner.Model
}

func NewAuthModel(c *config.Config) AuthModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(skyBlue)
	return AuthModel{
		s: s,
		c: c,
	}
}

func (a AuthModel) Init() tea.Cmd {
	return a.s.Tick
}

func (a AuthModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	a.s, cmd = a.s.Update(msg)
	return a, cmd
}

func (a AuthModel) View() string {
	return fmt.Sprintf("%s Authenicating as %s...", a.s.View(), a.c.Identifier)
}
