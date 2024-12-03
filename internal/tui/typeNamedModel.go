package tui

import tea "github.com/charmbracelet/bubbletea"

type NamedModel interface {
	Name() string
	Init() tea.Cmd
	Update(msg tea.Msg) (NamedModel, tea.Cmd)
	View() string
}
