package views

import (
	"context"
	"fmt"
	"strings"
	"time"

	"nezha_sec/internal/api"
	"nezha_sec/internal/model"
	"nezha_sec/internal/orchestrator"
	"nezha_sec/internal/registry"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MCPModel struct {
	model.TUI

	orchestrator *orchestrator.Orchestrator

	cancelFunc context.CancelFunc

	pendingToolCalls []api.ToolCallMsg

	autoConfirm bool

	activeTab string

	toolsList     list.Model
	tasksList     list.Model
	monitoringView viewport.Model

	targetInput textinput.Model

	currentTask string
	taskStatus  string

	confirmationMessage string

	userRole string
}

type ToolItem struct {
	name        string
	description string
	status      string
}

func (i ToolItem) Title() string       { return i.name }
func (i ToolItem) Description() string { return i.description }
func (i ToolItem) FilterValue() string { return i.name }

type TaskItem struct {
	id          string
	target      string
	status      string
	progress    string
	lastUpdated string
}

func (i TaskItem) Title() string       { return i.id }
func (i TaskItem) Description() string { return fmt.Sprintf("%s | %s | %s | %s", i.target, i.status, i.progress, i.lastUpdated) }
func (i TaskItem) FilterValue() string { return i.id }

func executeToolCallMCP(m *MCPModel, toolCall api.ToolCallMsg) tea.Cmd {
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

func NewMCPModel() (MCPModel, error) {
	toolRegistry, err := registry.NewToolRegistry()
	if err != nil {
		return MCPModel{}, err
	}

	orchestratorInstance, err := orchestrator.NewOrchestrator(toolRegistry)
	if err != nil {
		return MCPModel{}, err
	}

	tools := []list.Item{
		ToolItem{name: "nmap", description: "端口扫描工具", status: "可用"},
		ToolItem{name: "sqlmap", description: "SQL注入检测工具", status: "可用"},
		ToolItem{name: "curl", description: "HTTP请求工具", status: "可用"},
		ToolItem{name: "gobuster", description: "目录爆破工具", status: "可用"},
		ToolItem{name: "whatweb", description: "Web技术检测工具", status: "可用"},
		ToolItem{name: "wpscan", description: "WordPress扫描工具", status: "可用"},
		ToolItem{name: "httpx", description: "HTTP探测工具", status: "可用"},
		ToolItem{name: "nuclei", description: "漏洞扫描工具", status: "可用"},
		ToolItem{name: "ffuf", description: "模糊测试工具", status: "可用"},
		ToolItem{name: "trivy", description: "容器安全扫描工具", status: "可用"},
		ToolItem{name: "garak", description: "AI安全扫描工具", status: "可用"},
		ToolItem{name: "subfinder", description: "子域名发现工具", status: "可用"},
		ToolItem{name: "amass", description: "资产发现工具", status: "可用"},
		ToolItem{name: "naabu", description: "端口扫描工具", status: "可用"},
		ToolItem{name: "commix", description: "命令注入检测工具", status: "可用"},
		ToolItem{name: "sliver-cli", description: "C2工具", status: "可用"},
		ToolItem{name: "havoc", description: "C2工具", status: "可用"},
		ToolItem{name: "impacket", description: "横向移动工具", status: "可用"},
		ToolItem{name: "responder", description: "横向移动工具", status: "可用"},
		ToolItem{name: "crackmapexec", description: "横向移动工具", status: "可用"},
		ToolItem{name: "chisel", description: "端口转发工具", status: "可用"},
		ToolItem{name: "questexploit", description: "QuestDB SQL注入工具", status: "可用"},
	}

	tasks := []list.Item{
		TaskItem{id: "task-1", target: "example.com", status: "完成", progress: "100%", lastUpdated: time.Now().Format("2006-01-02 15:04:05")},
		TaskItem{id: "task-2", target: "localhost", status: "进行中", progress: "50%", lastUpdated: time.Now().Format("2006-01-02 15:04:05")},
		TaskItem{id: "task-3", target: "192.168.1.1", status: "等待中", progress: "0%", lastUpdated: time.Now().Format("2006-01-02 15:04:05")},
	}

	toolsList := list.New(tools, list.NewDefaultDelegate(), 0, 0)
	toolsList.Title = "可用工具"

	tasksList := list.New(tasks, list.NewDefaultDelegate(), 0, 0)
	tasksList.Title = "任务列表"

	monitoringView := viewport.New(0, 0)
	monitoringView.SetContent(`🔍 监控面板

系统状态: 正常
活跃任务: 1
完成任务: 1
等待任务: 1

最近活动:
- task-1: 完成扫描 example.com
- task-2: 正在扫描 localhost
- task-3: 等待扫描 192.168.1.1
`)

	targetInput := textinput.New()
	targetInput.Placeholder = "输入目标 URL 或 IP 地址"
	targetInput.Focus()

	return MCPModel{
		TUI:             model.InitialModel(),
		orchestrator:    orchestratorInstance,
		autoConfirm:     true,
		activeTab:       "dashboard",
		toolsList:       toolsList,
		tasksList:       tasksList,
		monitoringView:  monitoringView,
		targetInput:     targetInput,
		userRole:        "admin", // 默认角色
	}, nil
}

func (m MCPModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m MCPModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			if m.cancelFunc != nil {
				m.cancelFunc()
			}
			return m, tea.Quit
		case "tab":
			switch m.activeTab {
			case "dashboard":
				m.activeTab = "tools"
			case "tools":
				m.activeTab = "tasks"
			case "tasks":
				m.activeTab = "monitoring"
			case "monitoring":
				m.activeTab = "dashboard"
			}
			return m, nil
		case "enter":
			if m.activeTab == "dashboard" && m.targetInput.Value() != "" {
				allowed, reason := m.orchestrator.SetTarget(m.targetInput.Value(), true)
				if !allowed {
					m.Steps = append(m.Steps, "错误: "+reason)
					return m, nil
				}

				m.State = model.StateThinking

				progressMsg := m.orchestrator.StartAnalysis()
				m.Steps = append(m.Steps, string(progressMsg))

				ctx, cancel := context.WithCancel(context.Background())
				m.cancelFunc = cancel

				return m, tea.Batch(m.Spinner.Tick, api.CallDeepSeekAPIWithContext(ctx, m.targetInput.Value()))
			}
		case "p":
			if m.State == model.StateThinking {
				m.State = model.StatePaused
				return m, nil
			} else if m.State == model.StatePaused {
				m.State = model.StateThinking
				return m, m.Spinner.Tick
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
				return m, executeToolCallMCP(&m, m.pendingToolCalls[0])
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
			return m, executeToolCallMCP(&m, m.pendingToolCalls[0])
		} else {
			m.State = model.StateResult
			return m, nil
		}
	}

	switch m.activeTab {
	case "tools":
		m.toolsList, cmd = m.toolsList.Update(msg)
	case "tasks":
		m.tasksList, cmd = m.tasksList.Update(msg)
	case "monitoring":
		m.monitoringView, cmd = m.monitoringView.Update(msg)
	case "dashboard":
		if m.State == model.StateInput {
			m.targetInput, cmd = m.targetInput.Update(msg)
		}
	}

	return m, cmd
}

func (m MCPModel) View() string {
	baseStyle := lipgloss.NewStyle().Padding(1, 2)
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")).Background(lipgloss.Color("235"))
	tabStyle := lipgloss.NewStyle().Padding(1, 2).Border(lipgloss.NormalBorder(), false, false, true, false)
	activeTabStyle := tabStyle.Foreground(lipgloss.Color("12")).Background(lipgloss.Color("235")).Bold(true)
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("13"))
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("160"))
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	resultStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("5")).
		Padding(1, 2).
		MarginTop(1)
	stepStyle := lipgloss.NewStyle().PaddingLeft(2)

	tabs := lipgloss.JoinHorizontal(
		lipgloss.Top,
		ifThenElse(m.activeTab == "dashboard", activeTabStyle, tabStyle).Render("📊 仪表盘"),
		ifThenElse(m.activeTab == "tools", activeTabStyle, tabStyle).Render("🔧 工具"),
		ifThenElse(m.activeTab == "tasks", activeTabStyle, tabStyle).Render("📋 任务"),
		ifThenElse(m.activeTab == "monitoring", activeTabStyle, tabStyle).Render("👁️  监控"),
	)

	header := headerStyle.Render("🔐 哪吒网络安全分析器 (MCP)")
	userInfo := infoStyle.Render(fmt.Sprintf("用户: admin | 角色: %s | 时间: %s", m.userRole, time.Now().Format("2006-01-02 15:04:05")))

	content := ""
	switch m.activeTab {
	case "dashboard":
		switch m.State {
		case model.StateInput:
			title := titleStyle.Render("🌐 目标扫描")
			subtitle := infoStyle.Render("请输入目标 URL 或 IP 地址")
			input := m.targetInput.View()
			footer := infoStyle.Render("ENTER 确认 • TAB 切换 • ESC 退出")

			separator := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("─" + strings.Repeat("─", 60) + "─")

			content = fmt.Sprintf(
				"%s\n%s\n\n%s\n\n%s\n\n%s",
				title,
				separator,
				subtitle,
				input,
				footer,
			)
		case model.StateThinking:
			s := fmt.Sprintf("%s %s\n\n", m.Spinner.View(), successStyle.Render("DEEPSEEK 思考中..."))

			if len(m.Steps) > 0 {
				s += titleStyle.Render("执行步骤") + "\n"
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
				s += "\n" + titleStyle.Render("AI 分析结果") + "\n"
				s += resultStyle.Render(m.Result)
			}

			footer := infoStyle.Render("P 暂停 • TAB 切换 • CTRL+C 退出")
			s += "\n" + footer

			content = s
		case model.StatePaused:
			title := titleStyle.Render("⏸️  分析已暂停")
			s := title + "\n\n"

			if len(m.Steps) > 0 {
				s += titleStyle.Render("已执行步骤") + "\n"
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
				s += "\n" + titleStyle.Render("AI 分析结果") + "\n"
				s += resultStyle.Render(m.Result)
			}

			footer := infoStyle.Render("P 恢复 • TAB 切换 • CTRL+C 退出")
			s += "\n" + footer

			content = s
		case model.StateResult:
			title := titleStyle.Render("💡 分析完成")
			subtitle := infoStyle.Render("扫描结果摘要")
			result := resultStyle.Render(m.Result)
			footer := infoStyle.Render("ENTER 返回 • TAB 切换 • CTRL+C 退出")

			separator := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("─" + strings.Repeat("─", 60) + "─")

			content = fmt.Sprintf(
				"%s\n%s\n\n%s\n\n%s\n\n%s",
				title,
				separator,
				subtitle,
				result,
				footer,
			)
		case model.StateConfirmation:
			title := titleStyle.Render("⚠️  确认执行")
			message := m.confirmationMessage
			footer := infoStyle.Render("ENTER 确认 • N 取消 • TAB 切换 • CTRL+C 退出")

			separator := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("─" + strings.Repeat("─", 60) + "─")

			content = fmt.Sprintf(
				"%s\n%s\n\n%s\n\n%s",
				title,
				separator,
				message,
				footer,
			)
		}
	case "tools":
		content = m.toolsList.View()
	case "tasks":
		content = m.tasksList.View()
	case "monitoring":
		content = m.monitoringView.View()
	}

	return baseStyle.Render(fmt.Sprintf("%s\n%s\n\n%s\n\n%s", header, userInfo, tabs, content))
}

func ifThenElse(condition bool, trueStyle, falseStyle lipgloss.Style) lipgloss.Style {
	if condition {
		return trueStyle
	}
	return falseStyle
}
