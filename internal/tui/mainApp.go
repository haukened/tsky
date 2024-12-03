package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/haukened/tsky/internal/config"
	"github.com/haukened/tsky/internal/tokensvc"
	"github.com/haukened/tsky/internal/tui/styles"
)

type AppView struct {
	jwt        *tokensvc.Refresher
	tabs       map[int]NamedModel
	currentTab int
	w          int
	h          int
}

func NewAppView(c *config.Config) AppView {
	// create a new token svc
	jwt, err := tokensvc.NewRefresher(c)
	if err != nil {
		jwt = nil
	}
	return AppView{
		jwt: jwt,
		tabs: map[int]NamedModel{
			0: NewProfileTab(c.Did, c.Server, jwt),
		},
		currentTab: 0,
		w:          0,
		h:          0,
	}
}

func (a AppView) Name() string {
	return "app"
}

func (a AppView) Init() tea.Cmd {
	var cmds []tea.Cmd
	for _, model := range a.tabs {
		cmds = append(cmds, model.Init())
	}
	return tea.Batch(cmds...)
}

func (a AppView) Update(msg tea.Msg) (NamedModel, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// nothing yet
	case tea.WindowSizeMsg:
		a.w = msg.Width
		a.h = msg.Height
	}
	// update all tabs
	for i, model := range a.tabs {
		model, cmd := model.Update(msg)
		cmds = append(cmds, cmd)
		a.tabs[i] = model
	}
	// return the updated model and a batch of commands
	return a, tea.Batch(cmds...)
}

func (a AppView) View() string {
	tabRow := a.RenderTabs()
	tabContent := a.tabs[a.currentTab].View()
	return lipgloss.JoinVertical(0, tabRow, tabContent)
}

func (a AppView) RenderTabs() string {
	var tabs []string
	for i, model := range a.tabs {
		if i == a.currentTab {
			tabs = append(tabs, styles.ActiveTab.Render(model.Name()))
		} else {
			tabs = append(tabs, styles.Tab.Render(model.Name()))
		}
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
	gap := styles.TabGap.Render(strings.Repeat("+", max(0, a.w-2)))
	row = lipgloss.JoinHorizontal(lipgloss.Bottom, row, gap)
	return row
}
