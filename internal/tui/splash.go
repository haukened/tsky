package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/haukened/tsky/internal/messages"
)

type splash struct {
	countdown int
	done      bool
}

func (s splash) Init() tea.Cmd {
	return messages.Tick()
}

func hideSplash() tea.Msg {
	return messages.SplashMsg{}
}

func (s splash) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	if s.done {
		return s, nil
	}
	switch message.(type) {
	case messages.TickMsg:
		if s.countdown > 0 {
			s.countdown--
			if s.countdown <= 0 {
				s.done = true
				return s, hideSplash
			}
			return s, messages.Tick()
		}
	}
	return s, messages.Tick()
}

func (s splash) View() string {
	return lipgloss.NewStyle().Foreground(skyBlue).Render(logo)
}

const logo = `  ***                               ***  
 ********                       ******** 
 **********                   ********** 
 ************               ************ 
 **************           ************** 
 ***************         *************** 
 *****************     *****************    Welcome to
 ******************   ****************** 
  ****************** ******************  
    ********************************* 888    .d8888b.  888   
        *************************     888   d88P  Y88b 888
       ***************************    888   Y88b.      888
      *****************************   888888 "Y888b.   888  888 888  888
     *************** ***************  888       "Y88b. 888 .88P 888  888
     **************   **************  888         "888 888888K  888  888
      *************   *************   Y88b. Y88b  d88P 888 "88b Y88b 888
        *********       *********      "Y888 "Y8888P"  888  888  "Y88888
            ***           ***       								 888
                                                                Y8b d88P 
                                                                 "Y88P"`
