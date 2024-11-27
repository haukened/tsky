package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type authResult struct {
	success bool
	message string
}

type authStatusMsg bool

func authStatus(success bool) tea.Cmd {
	return func() tea.Msg {
		return authStatusMsg(success)
	}
}

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type helpMsg string

func SendHelpText(msg string) tea.Cmd {
	return func() tea.Msg {
		return helpMsg(msg)
	}
}

type statusMsg string

func SendStatusMsg(msg string) tea.Cmd {
	return func() tea.Msg {
		return statusMsg(msg)
	}
}

func SendStatusErr(msg string) tea.Cmd {
	return func() tea.Msg {
		return statusMsg(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF007C")).Render(msg))
	}
}

func ClearStatusMsg() tea.Cmd {
	return func() tea.Msg {
		return statusMsg("")
	}
}
