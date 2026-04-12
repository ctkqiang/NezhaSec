package views

import (
	"context"
	"encoding/json"
	"fmt"
	"nezha_sec/internal/api"
	"nezha_sec/internal/model"
	"nezha_sec/internal/orchestrator"
	"nezha_sec/internal/registry"
	"regexp"
	"strings"

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
	// 确认提示消息
	confirmationMessage string
	// 待执行的工具调用
	pendingToolCalls []api.ToolCallMsg
	// 自动确认模式
	autoConfirm bool
}

// 定义消息类型
type (
	ResultChunkMsg   string
	DoneMsg          struct{}
	ToolExecutionMsg struct {
		ToolName string
		Result   *model.ExecutionResult
		Error    error
	}
)

// 解析AI响应中的工具调用
func parseToolCalls(response string) []api.ToolCallMsg {
	// 查找tool_calls部分
	toolCallsRegex := regexp.MustCompile(`tool_calls\s*=\s*\[(.*?)\]`)
	matches := toolCallsRegex.FindStringSubmatch(response)
	if len(matches) < 2 {
		// 尝试另一种格式
		altRegex := regexp.MustCompile(`工具调用列表\s*:\s*\[(.*?)\]`)
		matches = altRegex.FindStringSubmatch(response)
		if len(matches) < 2 {
			return []api.ToolCallMsg{}
		}
	}

	// 提取工具调用JSON
	toolCallsJSON := "[" + matches[1] + "]"

	// 解析JSON
	var toolCalls []api.ToolCallMsg
	err := json.Unmarshal([]byte(toolCallsJSON), &toolCalls)
	if err != nil {
		// 尝试解析单个工具调用
		singleToolRegex := regexp.MustCompile(`-\s*` + "`" + `(\w+)` + "`" + `.*?\{([^}]+)\}`)
		singleMatches := singleToolRegex.FindAllStringSubmatch(response, -1)
		for _, match := range singleMatches {
			if len(match) >= 3 {
				toolName := match[1]
				argsStr := "{" + match[2] + "}"

				// 清理argsStr
				argsStr = strings.ReplaceAll(argsStr, "url", "\"url\"")
				argsStr = strings.ReplaceAll(argsStr, "=", ":")
				argsStr = strings.ReplaceAll(argsStr, "https://", "\"https://")
				argsStr = strings.ReplaceAll(argsStr, "http://", "\"http://")
				argsStr = strings.ReplaceAll(argsStr, "", "\"")
				argsStr = strings.ReplaceAll(argsStr, ", ", ", ")

				var args map[string]interface{}
				err := json.Unmarshal([]byte(argsStr), &args)
				if err == nil {
					toolCalls = append(toolCalls, api.ToolCallMsg{
						ToolName:  toolName,
						Arguments: args,
					})
				}
			}
		}
		return toolCalls
	}

	return toolCalls
}

// 执行工具调用
func executeToolCall(m *ChatModel, toolCall api.ToolCallMsg) tea.Cmd {
	return func() tea.Msg {
		if m == nil || m.orchestrator == nil {
			return ToolExecutionMsg{
				ToolName: toolCall.ToolName,
				Result:   nil,
				Error:    fmt.Errorf("模型或调度器为nil"),
			}
		}
		result, err := m.orchestrator.ExecuteTool(toolCall.ToolName, toolCall.Arguments)
		return ToolExecutionMsg{
			ToolName: toolCall.ToolName,
			Result:   result,
			Error:    err,
		}
	}
}

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
			} else if m.State == model.StateConfirmation {
				// 用户确认执行工具调用
				m.Steps = append(m.Steps, "用户确认执行工具调用")
				m.State = model.StateThinking
				// 执行第一个工具调用
				if len(m.pendingToolCalls) > 0 {
					return m, executeToolCall(&m, m.pendingToolCalls[0])
				}
				return m, nil
			}
		case "n":
			if m.State == model.StateConfirmation {
				// 用户取消执行工具调用
				m.Steps = append(m.Steps, "用户取消执行工具调用")
				m.State = model.StateResult
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
		m.Steps = append(m.Steps, "AI分析完成，准备执行工具调用...")

		// 解析工具调用
		toolCalls := parseToolCalls(msg.Result)
		if len(toolCalls) > 0 {
			// 存储待执行的工具调用
			m.pendingToolCalls = toolCalls

			if m.autoConfirm {
				// 自动确认模式：直接执行第一个工具调用
				m.Steps = append(m.Steps, "自动确认执行工具调用")
				return m, executeToolCall(&m, m.pendingToolCalls[0])
			} else {
				// 构建确认提示消息
				m.confirmationMessage = "AI 建议执行以下工具调用：\n"
				for i, toolCall := range toolCalls {
					m.confirmationMessage += fmt.Sprintf("%d. %s\n", i+1, toolCall.ToolName)
				}
				m.confirmationMessage += "\n是否开始执行？"

				// 切换到确认状态
				m.State = model.StateConfirmation
				return m, nil
			}
		} else {
			m.State = model.StateResult
			return m, nil
		}

	case ToolExecutionMsg:
		if msg.Error != nil {
			m.Steps = append(m.Steps, fmt.Sprintf("工具执行失败 %s: %v", msg.ToolName, msg.Error))
		} else {
			m.Steps = append(m.Steps, fmt.Sprintf("工具执行成功 %s", msg.ToolName))
		}

		// 检查是否还有更多工具需要执行
		if len(m.pendingToolCalls) > 1 {
			// 移除已执行的工具调用
			m.pendingToolCalls = m.pendingToolCalls[1:]
			// 自动执行下一个工具调用
			m.Steps = append(m.Steps, "自动执行下一个工具调用")
			return m, executeToolCall(&m, m.pendingToolCalls[0])
		} else {
			// 所有工具调用已执行完成
			m.State = model.StateResult
			return m, nil
		}

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
	case model.StateConfirmation:
		title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("13")).Render("⚠️  确认执行")
		return style.Render(fmt.Sprintf("%s\n\n%s\n\n%s", title, m.confirmationMessage, "ENTER 确认 • N 取消 • CTRL+C 退出"))
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
		autoConfirm:  true, // 启用自动确认模式
	}, nil
}
