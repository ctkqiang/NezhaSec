package model

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

type TUIState string

const (
	StateInput    TUIState = "input"
	StateThinking TUIState = "thinking"
	StatePaused   TUIState = "paused"
	StateResult   TUIState = "result"
)

type TUI struct {
	State    TUIState
	UrlInput textinput.Model
	Spinner  spinner.Model
	Steps    []string
	Result   string
}

func InitialModel() TUI {
	ti := textinput.New()
	ti.Placeholder = "请输入 URL (例如: https://github.com)..."
	ti.Focus()

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return TUI{
		State:    StateInput,
		UrlInput: ti,
		Spinner:  s,
		Steps:    []string{},
	}
}
