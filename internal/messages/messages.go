package messages

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type AuthResult struct {
	Success bool
	Message string
}

type AuthStatusMsg bool

func AuthStatus(success bool) tea.Cmd {
	return func() tea.Msg {
		return AuthStatusMsg(success)
	}
}

type SplashMsg struct{}

type TickMsg time.Time

func Tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

type HelpMsg string

func SendHelpText(msg string) tea.Cmd {
	return func() tea.Msg {
		return HelpMsg(msg)
	}
}

type LoginFinishedMsg bool

func LoginFinished() tea.Msg {
	return LoginFinishedMsg(true)
}

type ProfileMessage struct {
	LoadingError bool   `json:"-"` // Used to display error message
	Error        error  `json:"-"` // Used to store error message
	Did          string `json:"did"`
	Handle       string `json:"handle"`
	DisplayName  string `json:"displayName"`
	Avatar       string `json:"avatar"`
	Associated   struct {
		Lists        int  `json:"lists"`
		Feedgens     int  `json:"feedgens"`
		StarterPacks int  `json:"starterPacks"`
		Labeler      bool `json:"labeler"`
	} `json:"associated"`
	Viewer struct {
		Muted          bool `json:"muted"`
		BlockedBy      bool `json:"blockedBy"`
		KnownFollowers struct {
			Count     int `json:"count"`
			Followers []struct {
				Did         string `json:"did"`
				Handle      string `json:"handle"`
				DisplayName string `json:"displayName"`
				Avatar      string `json:"avatar"`
				Associated  struct {
					Chat struct {
						AllowIncoming string `json:"allowIncoming"`
					} `json:"chat"`
				} `json:"associated"`
				Viewer struct {
					Muted      bool   `json:"muted"`
					BlockedBy  bool   `json:"blockedBy"`
					Following  string `json:"following"`
					FollowedBy string `json:"followedBy"`
				} `json:"viewer"`
				Labels    []any     `json:"labels"`
				CreatedAt time.Time `json:"createdAt"`
			} `json:"followers"`
		} `json:"knownFollowers"`
	} `json:"viewer"`
	Labels         []any     `json:"labels"`
	CreatedAt      time.Time `json:"createdAt"`
	Description    string    `json:"description"`
	IndexedAt      time.Time `json:"indexedAt"`
	Banner         string    `json:"banner"`
	FollowersCount int       `json:"followersCount"`
	FollowsCount   int       `json:"followsCount"`
	PostsCount     int       `json:"postsCount"`
}

func SendProfileMsg(msg ProfileMessage) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}

type StatusMsg string

func SendStatusMsg(msg string) tea.Cmd {
	return func() tea.Msg {
		return StatusMsg(msg)
	}
}

func SendStatusErr(msg string) tea.Cmd {
	return func() tea.Msg {
		return StatusMsg(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF007C")).Render(msg))
	}
}

func ClearStatusMsg() tea.Cmd {
	return func() tea.Msg {
		return StatusMsg("")
	}
}