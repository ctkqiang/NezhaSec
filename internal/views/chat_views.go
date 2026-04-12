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

	orchestrator *orchestrator.Orchestrator

	cancelFunc context.CancelFunc

	confirmationMessage string

	pendingToolCalls []api.ToolCallMsg

	autoConfirm bool
}

type (
	ResultChunkMsg   string
	DoneMsg          struct{}
	ToolExecutionMsg struct {
		ToolName string
		Result   *model.ExecutionResult
		Error    error
	}
)

func parseToolCalls(response string) []api.ToolCallMsg {

	var toolCalls []api.ToolCallMsg

	toolCallsRegex := regexp.MustCompile(`tool_calls\s*=\s*\[(.*?)\]`)
	matches := toolCallsRegex.FindStringSubmatch(response)
	if len(matches) >= 2 {
		toolCallsJSON := "[" + matches[1] + "]"
		var parsedCalls []api.ToolCallMsg
		if err := json.Unmarshal([]byte(toolCallsJSON), &parsedCalls); err == nil {
			return parsedCalls
		}
	}

	altRegex := regexp.MustCompile(`工具调用列表\s*:\s*\[(.*?)\]`)
	matches = altRegex.FindStringSubmatch(response)
	if len(matches) >= 2 {
		toolCallsJSON := "[" + matches[1] + "]"
		var parsedCalls []api.ToolCallMsg
		if err := json.Unmarshal([]byte(toolCallsJSON), &parsedCalls); err == nil {
			return parsedCalls
		}
	}

	singleToolRegex := regexp.MustCompile(`工具调用:\s*(\w+)\s*\{([^}]+)\}`)
	singleMatches := singleToolRegex.FindAllStringSubmatch(response, -1)
	for _, match := range singleMatches {
		if len(match) >= 3 {
			toolName := match[1]
			argsStr := "{" + match[2] + "}"

			argsStr = strings.ReplaceAll(argsStr, "url", "\"url\"")
			argsStr = strings.ReplaceAll(argsStr, "=", ":")
			argsStr = strings.ReplaceAll(argsStr, "https://", "https://")
			argsStr = strings.ReplaceAll(argsStr, "http://", "http://")
			argsStr = strings.ReplaceAll(argsStr, "", "\"")
			argsStr = strings.ReplaceAll(argsStr, ", ", ", ")

			var args map[string]interface{}
			if err := json.Unmarshal([]byte(argsStr), &args); err == nil {
				toolCalls = append(toolCalls, api.ToolCallMsg{
					ToolName:  toolName,
					Arguments: args,
				})
			}
		}
	}

	listToolRegex := regexp.MustCompile(`-\s*(\w+)\s*\{([^}]+)\}`)
	listMatches := listToolRegex.FindAllStringSubmatch(response, -1)
	for _, match := range listMatches {
		if len(match) >= 3 {
			toolName := match[1]
			argsStr := "{" + match[2] + "}"

			argsStr = strings.ReplaceAll(argsStr, "url", "\"url\"")
			argsStr = strings.ReplaceAll(argsStr, "=", ":")
			argsStr = strings.ReplaceAll(argsStr, "https://", "https://")
			argsStr = strings.ReplaceAll(argsStr, "http://", "http://")
			argsStr = strings.ReplaceAll(argsStr, "", "\"")
			argsStr = strings.ReplaceAll(argsStr, ", ", ", ")

			var args map[string]interface{}
			if err := json.Unmarshal([]byte(argsStr), &args); err == nil {

				exists := false
				for _, existingCall := range toolCalls {
					if existingCall.ToolName == toolName {
						exists = true
						break
					}
				}
				if !exists {
					toolCalls = append(toolCalls, api.ToolCallMsg{
						ToolName:  toolName,
						Arguments: args,
					})
				}
			}
		}
	}

	runToolRegex := regexp.MustCompile(`工具名称:\s*run_(\w+)`)
	runMatches := runToolRegex.FindAllStringSubmatch(response, -1)
	for _, match := range runMatches {
		if len(match) >= 2 {
			toolName := "run_" + match[1]

			args := make(map[string]interface{})

			paramRegex := regexp.MustCompile(`参数:\s*\{([^}]+)\}`)
			paramMatches := paramRegex.FindStringSubmatch(response)
			if len(paramMatches) >= 2 {
				paramStr := "{" + paramMatches[1] + "}"

				paramStr = strings.ReplaceAll(paramStr, "target", "\"target\"")
				paramStr = strings.ReplaceAll(paramStr, "ports", "\"ports\"")
				paramStr = strings.ReplaceAll(paramStr, "scan_type", "\"scan_type\"")
				paramStr = strings.ReplaceAll(paramStr, "=", ":")
				paramStr = strings.ReplaceAll(paramStr, "https://", "https://")
				paramStr = strings.ReplaceAll(paramStr, "http://", "http://")
				paramStr = strings.ReplaceAll(paramStr, "", "\"")
				paramStr = strings.ReplaceAll(paramStr, ", ", ", ")

				var paramArgs map[string]interface{}
				if err := json.Unmarshal([]byte(paramStr), &paramArgs); err == nil {
					args = paramArgs
				}
			}

			exists := false
			for _, existingCall := range toolCalls {
				if existingCall.ToolName == toolName {
					exists = true
					break
				}
			}
			if !exists {
				toolCalls = append(toolCalls, api.ToolCallMsg{
					ToolName:  toolName,
					Arguments: args,
				})
			}
		}
	}

	if len(toolCalls) == 0 {

		commonTools := []string{"curl", "nmap", "sqlmap", "gobuster", "whatweb", "wpscan"}
		for _, tool := range commonTools {
			if strings.Contains(response, tool) {

				args := make(map[string]interface{})

				toolCalls = append(toolCalls, api.ToolCallMsg{
					ToolName:  tool,
					Arguments: args,
				})
			}
		}
	}

	if len(toolCalls) == 0 {

		args := make(map[string]interface{})
		args["url"] = "localhost"
		toolCalls = append(toolCalls, api.ToolCallMsg{
			ToolName:  "curl",
			Arguments: args,
		})
	}

	return toolCalls
}

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

func (m ChatModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":

			if m.cancelFunc != nil {
				m.cancelFunc()
			}
			return m, tea.Quit
		case "p":

			if m.State == model.StateThinking {
				m.State = model.StatePaused
				return m, nil
			} else if m.State == model.StatePaused {
				m.State = model.StateThinking
				return m, m.Spinner.Tick
			}
		case "enter":
			if m.State == model.StateInput && m.UrlInput.Value() != "" {

				allowed, reason := m.orchestrator.SetTarget(m.UrlInput.Value(), true)
				if !allowed {

					m.Steps = append(m.Steps, "错误: "+reason)
					return m, nil
				}

				m.State = model.StateThinking

				progressMsg := m.orchestrator.StartAnalysis()
				m.Steps = append(m.Steps, string(progressMsg))

				ctx, cancel := context.WithCancel(context.Background())
				m.cancelFunc = cancel

				return m, tea.Batch(m.Spinner.Tick, api.CallDeepSeekAPIWithContext(ctx, m.UrlInput.Value()))
			} else if m.State == model.StateResult {

				m.State = model.StateInput
				m.Steps = []string{}
				m.Result = ""
				m.UrlInput.Reset()
				m.UrlInput.Focus()
				return m, nil
			} else if m.State == model.StateConfirmation {

				m.Steps = append(m.Steps, "用户确认执行工具调用")
				m.State = model.StateThinking

				if len(m.pendingToolCalls) > 0 {
					return m, executeToolCall(&m, m.pendingToolCalls[0])
				}
				return m, nil
			}
		case "n":
			if m.State == model.StateConfirmation {

				m.Steps = append(m.Steps, "用户取消执行工具调用")
				m.State = model.StateResult
				return m, nil
			}
		case "y", "yes":
			if m.State == model.StateConfirmation {

				m.Steps = append(m.Steps, "用户确认执行工具调用")
				m.State = model.StateThinking

				if len(m.pendingToolCalls) > 0 {
					return m, executeToolCall(&m, m.pendingToolCalls[0])
				}
				return m, nil
			}
		}

	case spinner.TickMsg:
		m.Spinner, cmd = m.Spinner.Update(msg)
		return m, cmd

	case api.ProgressMsg:
		msgStr := string(msg)
		m.Steps = append(m.Steps, msgStr)

		if len(msgStr) > 4 && msgStr[:4] == "错误: " {
			m.Spinner, _ = m.Spinner.Update(spinner.TickMsg{})
		}
		return m, nil

	case api.DeepSeekResponseMsg:
		m.Result = msg.Result
		m.Steps = append(m.Steps, "AI分析完成，准备执行工具调用...")

		toolCalls := parseToolCalls(msg.Result)
		if len(toolCalls) > 0 {

			m.pendingToolCalls = toolCalls

			if m.autoConfirm {

				m.Steps = append(m.Steps, "自动确认执行工具调用")
				return m, executeToolCall(&m, m.pendingToolCalls[0])
			} else {

				m.confirmationMessage = "AI 建议执行以下工具调用：\n"
				for i, toolCall := range toolCalls {
					m.confirmationMessage += fmt.Sprintf("%d. %s\n", i+1, toolCall.ToolName)
				}
				m.confirmationMessage += "\n是否开始执行？"

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

		if len(m.pendingToolCalls) > 1 {

			m.pendingToolCalls = m.pendingToolCalls[1:]

			m.Steps = append(m.Steps, "自动执行下一个工具调用")
			return m, executeToolCall(&m, m.pendingToolCalls[0])
		} else {

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

func (m ChatModel) View() string {

	baseStyle := lipgloss.NewStyle().Padding(1, 2)
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	subheaderStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("13"))
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("160"))
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	resultStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("5")).
		Padding(1, 2).
		MarginTop(1)
	stepStyle := lipgloss.NewStyle().PaddingLeft(2)

	switch m.State {
	case model.StateInput:
		header := headerStyle.Render("🌐 哪吒网络安全分析器")
		subheader := subheaderStyle.Render("请输入目标 URL 或 IP 地址")
		input := m.UrlInput.View()
		footer := infoStyle.Render("ENTER 确认 • ESC 退出")

		separator := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("─" + strings.Repeat("─", 60) + "─")

		return baseStyle.Render(fmt.Sprintf(
			"%s\n%s\n\n%s\n\n%s\n%s",
			header,
			separator,
			subheader,
			input,
			footer,
		))

	case model.StateThinking:
		s := fmt.Sprintf("%s %s\n\n", m.Spinner.View(), successStyle.Render("DEEPSEEK 思考中..."))

		if len(m.Steps) > 0 {
			s += subheaderStyle.Render("执行步骤") + "\n"
			for i, step := range m.Steps {
				check := successStyle.Render("✓")
				stepText := step

				if strings.HasPrefix(step, "错误:") {
					stepText = errorStyle.Render(step)
				}
				s += stepStyle.Render(fmt.Sprintf("%s %d. %s", check, i+1, stepText)) + "\n"
			}
		}

		if m.Result != "" {
			s += "\n" + subheaderStyle.Render("AI 分析结果") + "\n"
			s += resultStyle.Render(m.Result)
		}

		footer := infoStyle.Render("P 暂停 • CTRL+C 退出")
		s += "\n" + footer

		return baseStyle.Render(s)

	case model.StatePaused:
		title := subheaderStyle.Render("⏸️  分析已暂停")
		s := title + "\n\n"

		if len(m.Steps) > 0 {
			s += subheaderStyle.Render("已执行步骤") + "\n"
			for i, step := range m.Steps {
				check := successStyle.Render("✓")
				stepText := step

				if strings.HasPrefix(step, "错误:") {
					stepText = errorStyle.Render(step)
				}
				s += stepStyle.Render(fmt.Sprintf("%s %d. %s", check, i+1, stepText)) + "\n"
			}
		}

		if m.Result != "" {
			s += "\n" + subheaderStyle.Render("AI 分析结果") + "\n"
			s += resultStyle.Render(m.Result)
		}

		footer := infoStyle.Render("P 恢复 • CTRL+C 退出")
		s += "\n" + footer

		return baseStyle.Render(s)

	case model.StateResult:
		title := headerStyle.Render("💡 分析完成")
		subheader := subheaderStyle.Render("扫描结果摘要")
		result := resultStyle.Render(m.Result)
		footer := infoStyle.Render("ENTER 返回 • CTRL+C 退出")

		separator := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("─" + strings.Repeat("─", 60) + "─")

		return baseStyle.Render(fmt.Sprintf(
			"%s\n%s\n\n%s\n\n%s\n\n%s",
			title,
			separator,
			subheader,
			result,
			footer,
		))
	case model.StateConfirmation:
		title := subheaderStyle.Render("⚠️  确认执行")
		message := m.confirmationMessage
		footer := infoStyle.Render("ENTER 确认 • N 取消 • CTRL+C 退出")

		separator := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("─" + strings.Repeat("─", 60) + "─")

		return baseStyle.Render(fmt.Sprintf(
			"%s\n%s\n\n%s\n\n%s",
			title,
			separator,
			message,
			footer,
		))
	}

	return ""
}

func NewChatModel() (ChatModel, error) {

	toolRegistry, err := registry.NewToolRegistry()
	if err != nil {
		return ChatModel{}, err
	}

	orchestratorInstance, err := orchestrator.NewOrchestrator(toolRegistry)
	if err != nil {
		return ChatModel{}, err
	}

	return ChatModel{
		TUI:          model.InitialModel(),
		orchestrator: orchestratorInstance,
		autoConfirm:  true,
	}, nil
}
