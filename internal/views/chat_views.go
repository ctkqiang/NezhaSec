package views

import (
	"context"
	"fmt"
	"nezha_sec/internal/api"
	"nezha_sec/internal/model"
	"nezha_sec/internal/orchestrator"
	"nezha_sec/internal/registry"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ChatModel struct {
	model.TUI
	// 核心调度器
	orchestrator *orchestrator.Orchestrator
	// 取消函数，用于取消正在进行的操作
	cancelFunc context.CancelFunc
}

// 定义消息类型
type (
	ResultChunkMsg string
	DoneMsg        struct{}
)

// 实现Init方法
func (m ChatModel) Init() tea.Cmd {
	return textinput.Blink
}

// 实现Update方法
func (m ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			// 清理资源
			if m.cancelFunc != nil {
				m.cancelFunc()
			}
			return m, tea.Quit
		case "p":
			// 切换暂停/恢复状态
			if m.State == model.StateThinking {
				m.State = model.StatePaused
				return m, nil
			} else if m.State == model.StatePaused {
				m.State = model.StateThinking
				return m, m.Spinner.Tick
			}
		case "enter":
			if m.State == model.StateInput && m.UrlInput.Value() != "" {
				// 设置目标URL，暂时允许扫描本地地址
				allowed, reason := m.orchestrator.SetTarget(m.UrlInput.Value(), true)
				if !allowed {
					// 目标不允许扫描，显示错误信息
					m.Steps = append(m.Steps, "错误: "+reason)
					return m, nil
				}

				// 切换到思考状态
				m.State = model.StateThinking

				// 启动分析
				progressMsg := m.orchestrator.StartAnalysis()
				m.Steps = append(m.Steps, string(progressMsg))

				// 创建上下文和取消函数
				ctx, cancel := context.WithCancel(context.Background())
				m.cancelFunc = cancel

				// 调用DeepSeek API进行分析
				return m, tea.Batch(m.Spinner.Tick, api.CallDeepSeekAPIWithContext(ctx, m.UrlInput.Value()))
			} else if m.State == model.StateResult {
				// 分析完成后按Enter键返回输入状态
				m.State = model.StateInput
				m.Steps = []string{}
				m.Result = ""
				m.UrlInput.Reset()
				m.UrlInput.Focus()
				return m, nil
			}
		}

	case spinner.TickMsg:
		m.Spinner, cmd = m.Spinner.Update(msg)
		return m, cmd

	case api.ProgressMsg:
		msgStr := string(msg)
		m.Steps = append(m.Steps, msgStr)

		// 检查是否是错误消息
		if len(msgStr) > 4 && msgStr[:4] == "错误: " {
			// 错误发生，切换回输入状态
			m.State = model.StateInput
			m.Spinner, _ = m.Spinner.Update(spinner.TickMsg{})
			m.UrlInput.Focus()
		}
		return m, nil

	case api.DeepSeekResponseMsg:
		m.Result = msg.Result
		m.State = model.StateResult
		return m, nil

	case ResultChunkMsg:
		m.Result += string(msg)
		return m, nil

	case DoneMsg:
		m.State = model.StateResult
		return m, nil
	}

	if m.State == model.StateInput {
		m.UrlInput, cmd = m.UrlInput.Update(msg)
	}

	return m, cmd
}

// 实现View方法
func (m ChatModel) View() string {
	style := lipgloss.NewStyle().Padding(1, 2)

	switch m.State {
	case model.StateInput:
		header := lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true).Render("🌐 AI URL ANALYZER")
		return style.Render(fmt.Sprintf(
			"%s\n\n%s\n\n%s",
			header,
			m.UrlInput.View(),
			lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("ENTER 确认 • ESC 退出"),
		))

	case model.StateThinking:
		s := fmt.Sprintf("%s %s\n\n", m.Spinner.View(), "DEEPSEEK 思考中...")

		for _, step := range m.Steps {
			check := lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓")
			s += fmt.Sprintf("%s %s\n", check, step)
		}

		if m.Result != "" {
			resStyle := lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(lipgloss.Color("5")).
				PaddingLeft(2).
				MarginTop(1)
			s += "\n" + resStyle.Render(m.Result)
		}
		s += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("P 暂停 • CTRL+C 退出")
		return style.Render(s)

	case model.StatePaused:
		title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("13")).Render("⏸️  分析已暂停")
		s := title + "\n\n"

		for _, step := range m.Steps {
			check := lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓")
			s += fmt.Sprintf("%s %s\n", check, step)
		}

		if m.Result != "" {
			resStyle := lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(lipgloss.Color("5")).
				PaddingLeft(2).
				MarginTop(1)
			s += "\n" + resStyle.Render(m.Result)
		}
		s += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("P 恢复 • CTRL+C 退出")
		return style.Render(s)

	case model.StateResult:
		title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5")).Render("💡 分析完成")
		return style.Render(fmt.Sprintf("%s\n\n%s\n\n%s", title, m.Result, "ENTER 返回 • CTRL+C 退出"))
	}

	return ""
}

// NewChatModel 创建一个新的ChatModel实例
// 返回值列表：
//   - ChatModel 新创建的ChatModel实例
//   - error 初始化过程中可能发生的错误
func NewChatModel() (ChatModel, error) {
	// 创建工具注册表
	toolRegistry, err := registry.NewToolRegistry()
	if err != nil {
		return ChatModel{}, err
	}

	// 创建调度器
	orchestratorInstance, err := orchestrator.NewOrchestrator(toolRegistry)
	if err != nil {
		return ChatModel{}, err
	}

	return ChatModel{
		TUI:          model.InitialModel(),
		orchestrator: orchestratorInstance,
	}, nil
}
