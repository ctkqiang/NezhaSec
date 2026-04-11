package main

import (
	"fmt"
	"nezha_sec/internal/model"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type (
	stepMsg   string
	resultMsg string
)

type modelTUI struct {
	model.TUI
}

func (m modelTUI) Init() tea.Cmd {
	return nil
}

func (m modelTUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.State {

	case model.StateInput:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "enter" && m.UrlInput.Value() != "" {
				m.State = model.StateThinking
				// 切换到思考状态，并启动模拟的“思考过程”
				return m, tea.Batch(m.Spinner.Tick, simulateProcess())
			}
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
		}
		var cmd tea.Cmd
		m.UrlInput, cmd = m.UrlInput.Update(msg)
		return m, cmd

	case model.StateThinking:
		switch msg := msg.(type) {
		case stepMsg:
			m.Steps = append(m.Steps, string(msg))
			return m, nil
		case resultMsg:
			m.State = model.StateResult
			m.Result = string(msg)
			return m, nil
		case spinner.TickMsg:
			var cmd tea.Cmd
			m.Spinner, cmd = m.Spinner.Update(msg)
			return m, cmd
		}

	case model.StateResult:
		if msg, ok := msg.(tea.KeyMsg); ok && msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m modelTUI) View() string {
	var s string

	style := lipgloss.NewStyle().Padding(1, 2)

	switch m.State {
	case model.StateInput:
		s = fmt.Sprintf(
			"%s\n\n%s\n\n%s",
			"请输入需要分析的 URL",
			m.UrlInput.View(),
			lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("按 Enter 确认 • Ctrl+C 退出"),
		)

	case model.StateThinking:
		s = fmt.Sprintf("%s %s\n\n", m.Spinner.View(), "AI 正在思考中...")
		for _, step := range m.Steps {
			checkmark := lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓")
			s += fmt.Sprintf("%s %s\n", checkmark, step)
		}

	case model.StateResult:
		title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99")).Render("分析报告")
		s = fmt.Sprintf("%s\n\n%s\n\n%s", title, m.Result, "按 Ctrl+C 退出程序")
	}

	return style.Render(s)
}

func simulateProcess() tea.Cmd {
	return func() tea.Msg {
		// 模拟步骤一
		time.Sleep(800 * time.Millisecond)
		return resultMsg("这是一篇关于 Golang TUI 开发的高质量文章，建议阅读。")
	}
}

func main() {
	// 创建本地modelTUI实例
	initialModel := modelTUI{model.InitialModel()}
	p := tea.NewProgram(initialModel)
	if _, err := p.Run(); err != nil {
		fmt.Printf("程序运行出错: %v", err)
	}
}
