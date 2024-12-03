package styles

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/gamut"
)

// Style definitions.
// initially sourced from https://github.com/charmbracelet/lipgloss/blob/master/examples/layout/main.go
var (

	// General.

	Normal    = lipgloss.Color("#EEEEEE")
	Error     = lipgloss.Color("#FF0000")
	Subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	Highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	Special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}
	Blends    = gamut.Blends(lipgloss.Color("#F25D94"), lipgloss.Color("#EDFF82"), 50)

	Base = lipgloss.NewStyle().Foreground(Normal)

	Divider = lipgloss.NewStyle().
		SetString("•").
		Padding(0, 1).
		Foreground(Subtle).
		String()

	URL = lipgloss.NewStyle().Foreground(Special).Render

	// Tabs.

	ActiveTabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      " ",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┘",
		BottomRight: "└",
	}

	TabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┴",
		BottomRight: "┴",
	}

	Tab = lipgloss.NewStyle().
		Border(TabBorder, true).
		BorderForeground(Highlight).
		Padding(0, 1)

	ActiveTab = Tab.Border(ActiveTabBorder, true)

	TabGap = Tab.
		BorderTop(false).
		BorderLeft(false).
		BorderRight(false)

	// Title.

	TitleStyle = lipgloss.NewStyle().
			MarginLeft(1).
			MarginRight(5).
			Padding(0, 1).
			Italic(true).
			Foreground(lipgloss.Color("#FFF7DB")).
			SetString("Lip Gloss")

	// Page.

	DocStyle = lipgloss.NewStyle().Padding(1, 2, 1, 2)
)
