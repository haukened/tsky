package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/haukened/tsky/internal/config"
	"github.com/haukened/tsky/internal/tui"
	"github.com/haukened/tsky/internal/utils"
)

var Version string = "dev"

func dontPanic(err error) {
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}

func main() {
	utils.SetVersion(Version)
	c, err := config.New("~/.config/tsky/config.yaml")
	dontPanic(err)
	err = c.Load()
	dontPanic(err)
	p := tea.NewProgram(tui.NewModel(c), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(1)
	}
}
