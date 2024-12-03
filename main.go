package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/haukened/tsky/internal/config"
	"github.com/haukened/tsky/internal/debug"
	"github.com/haukened/tsky/internal/tui"
	"github.com/haukened/tsky/internal/utils"
)

var Version string = "dev"

const LOG_FILE = "./tsky_debug.log"

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
	if c.Debug {
		debug.SetDebug(true)
		logFile, err := os.OpenFile(LOG_FILE, os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0644)
		dontPanic(err)
		defer logFile.Close()
		log.SetOutput(logFile)
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Println("Starting tsky")
	}
	p := tea.NewProgram(tui.NewModel(c), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(1)
	}
}
