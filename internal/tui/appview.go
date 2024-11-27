package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/haukened/tsky/internal/config"
	"github.com/haukened/tsky/internal/tokensvc"
	"github.com/haukened/tsky/internal/tui/tabs"
)

type AppView struct {
	jwt        *tokensvc.Refresher
	tabs       map[int]tea.Model
	currentTab int
}

func NewAppView(c *config.Config) AppView {
	// create a new token svc
	jwt, err := tokensvc.NewRefresher(c)
	if err != nil {
		jwt = nil
	}
	return AppView{
		jwt: jwt,
		tabs: map[int]tea.Model{
			0: tabs.NewProfileTab(c.Did, c.Server, jwt),
		},
		currentTab: 0,
	}
}

func (a AppView) Init() tea.Cmd {
	var cmds []tea.Cmd
	for _, model := range a.tabs {
		cmds = append(cmds, model.Init())
	}
	return tea.Batch(cmds...)
}

func (a AppView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	for i, model := range a.tabs {
		model, cmd := model.Update(msg)
		cmds = append(cmds, cmd)
		a.tabs[i] = model
	}
	return a, tea.Batch(cmds...)
}

func (a AppView) View() string {
	t, ok := a.tabs[a.currentTab]
	if ok {
		return t.View()
	}
	return "Error"
}
