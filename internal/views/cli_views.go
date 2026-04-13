package views

import (
	"context"
	"fmt"
	"nezha_sec/internal/orchestrator"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/joho/godotenv"
)

const (
	ColorPrimary    = lipgloss.Color("#6366f1")
	ColorSecondary  = lipgloss.Color("#8b5cf6")
	ColorSuccess    = lipgloss.Color("#10b981")
	ColorWarning    = lipgloss.Color("#f59e0b")
	ColorError      = lipgloss.Color("#ef4444")
	ColorInfo       = lipgloss.Color("#3b82f6")
	ColorBackground = lipgloss.Color("#1e1e2e")
	ColorSurface    = lipgloss.Color("#313244")
	ColorText       = lipgloss.Color("#cdd6f4")
	ColorMuted      = lipgloss.Color("#6c7086")
)

type LogEntry struct {
	Time    time.Time
	Level   string
	Message string
}

type CLIWorkflowModel struct {
	orchestrator    *orchestrator.Orchestrator
	workflowManager *orchestrator.WorkflowManager

	cancelFunc context.CancelFunc

	targetInput textinput.Model

	spinner       spinner.Model
	prog          progress.Model
	secondaryProg progress.Model

	viewport viewport.Model

	state                    string
	currentPhase             string
	progressPercent          float64
	secondaryProgressPercent float64

	logs      []LogEntry
	errors    []string
	successes []string

	config map[string]interface{}

	totalPhases       int
	currentPhaseIndex int

	messageChan chan interface{}
}

func getAppName() string {
	_ = godotenv.Load(".env")
	appName := os.Getenv("APP_NAME")
	if appName == "" {
		appName = "哪吒网络安全分析器"
	}
	return appName
}

func getAuthor() string {
	return "钟智强"
}

func NewCLIWorkflowModel(orchestrator *orchestrator.Orchestrator) (CLIWorkflowModel, error) {
	workflowManager := orchestrator.NewWorkflowManager()

	messageChan := make(chan interface{}, 100)
	workflowManager.SetMessageChannel(messageChan)

	spinner := spinner.New()
	spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))

	prog := progress.New()

	secondaryProg := progress.New()

	viewport := viewport.New(70, 15)
	viewport.Style = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(ColorSurface).
		Padding(1)

	targetInput := textinput.New()
	targetInput.Placeholder = "输入目标 URL 或 IP 地址"
	targetInput.Focus()

	return CLIWorkflowModel{
		orchestrator:      orchestrator,
		workflowManager:   workflowManager,
		messageChan:       messageChan,
		targetInput:       targetInput,
		spinner:           spinner,
		prog:              prog,
		secondaryProg:     secondaryProg,
		viewport:          viewport,
		state:             "input",
		logs:              []LogEntry{},
		errors:            []string{},
		successes:         []string{},
		totalPhases:       7,
		currentPhaseIndex: 0,
		config: map[string]interface{}{
			"aggressive": false,
			"timeout":    300,
			"threads":    10,
		},
	}, nil
}

func (m CLIWorkflowModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick, m.listenForMessages())
}

func (m CLIWorkflowModel) listenForMessages() tea.Cmd {
	return func() tea.Msg {
		select {
		case msg, ok := <-m.messageChan:
			if !ok {
				return nil
			}
			switch val := msg.(type) {
			case struct{ Phase string }:
				return WorkflowPhaseMsg{Phase: val.Phase}
			case struct{ Log string }:
				return WorkflowLogMsg{Log: val.Log}
			case struct{ Message string }:
				return WorkflowSuccessMsg{Message: val.Message}
			case struct{ Error string }:
				return WorkflowErrorMsg{Error: val.Error}
			}
		default:
		}
		return nil
	}
}

func (m CLIWorkflowModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			if m.cancelFunc != nil {
				m.cancelFunc()
			}
			return m, tea.Quit
		case "enter":
			if m.state == "input" && m.targetInput.Value() != "" {
				err := m.workflowManager.SetTarget(m.targetInput.Value())
				if err != nil {
					m.errors = append(m.errors, fmt.Sprintf("设置目标失败: %v", err))
					return m, nil
				}

				m.state = "running"
				m.currentPhase = "初始化"
				m.currentPhaseIndex = 0
				m.progressPercent = 0
				m.secondaryProgressPercent = 0
				m.logs = append(m.logs, LogEntry{
					Time:    time.Now(),
					Level:   "INFO",
					Message: fmt.Sprintf("开始渗透测试工作流，目标: %s", m.targetInput.Value()),
				})

				ctx, cancel := context.WithCancel(context.Background())
				m.cancelFunc = cancel

				return m, tea.Batch(
					m.spinner.Tick,
					m.listenForMessages(),
					func() tea.Msg {
						err := m.workflowManager.ExecuteWorkflow(ctx)
						if err != nil {
							return WorkflowErrorMsg{Error: err.Error()}
						}
						return WorkflowDoneMsg{}
					},
				)
			}
		case "p":
			if m.state == "running" {
				m.state = "paused"
				return m, nil
			} else if m.state == "paused" {
				m.state = "running"
				return m, m.spinner.Tick
			}
		case "c":
			if m.state == "completed" {
				m.state = "input"
				m.targetInput.Reset()
				m.targetInput.Focus()
				m.logs = []LogEntry{}
				m.errors = []string{}
				m.successes = []string{}
				m.progressPercent = 0
				m.secondaryProgressPercent = 0
				m.currentPhase = ""
				m.currentPhaseIndex = 0
				return m, nil
			}
		}

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		if m.state == "running" {
			return m, tea.Batch(cmd, m.listenForMessages())
		}
		return m, cmd

	case WorkflowPhaseMsg:
		m.currentPhase = msg.Phase
		m.currentPhaseIndex++
		m.progressPercent = float64(m.currentPhaseIndex) / float64(m.totalPhases)
		m.secondaryProgressPercent = 0
		m.logs = append(m.logs, LogEntry{
			Time:    time.Now(),
			Level:   "PHASE",
			Message: fmt.Sprintf("开始阶段: %s", msg.Phase),
		})
		m.updateViewportContent()
		return m, m.listenForMessages()

	case WorkflowLogMsg:
		m.logs = append(m.logs, LogEntry{
			Time:    time.Now(),
			Level:   "INFO",
			Message: msg.Log,
		})
		m.secondaryProgressPercent += 0.05
		if m.secondaryProgressPercent > 1 {
			m.secondaryProgressPercent = 1
		}
		m.updateViewportContent()
		return m, m.listenForMessages()

	case WorkflowSuccessMsg:
		m.successes = append(m.successes, msg.Message)
		m.logs = append(m.logs, LogEntry{
			Time:    time.Now(),
			Level:   "SUCCESS",
			Message: msg.Message,
		})
		m.updateViewportContent()
		return m, m.listenForMessages()

	case WorkflowErrorMsg:
		m.errors = append(m.errors, fmt.Sprintf("[%s] 错误: %s", time.Now().Format("15:04:05"), msg.Error))
		m.logs = append(m.logs, LogEntry{
			Time:    time.Now(),
			Level:   "ERROR",
			Message: msg.Error,
		})
		m.updateViewportContent()
		return m, m.listenForMessages()

	case WorkflowProgressMsg:
		m.progressPercent = msg.Progress
		if msg.Secondary > 0 {
			m.secondaryProgressPercent = msg.Secondary
		}
		return m, nil

	case WorkflowDoneMsg:
		m.state = "completed"
		m.progressPercent = 1
		m.logs = append(m.logs, LogEntry{
			Time:    time.Now(),
			Level:   "SUCCESS",
			Message: "工作流执行完成",
		})
		m.updateViewportContent()
		return m, nil
	}

	if m.state == "input" {
		m.targetInput, cmd = m.targetInput.Update(msg)
	} else if m.state == "running" {
		m.spinner, cmd = m.spinner.Update(msg)
	}

	return m, cmd
}

func (m *CLIWorkflowModel) updateViewportContent() {
	var content string
	for _, entry := range m.logs {
		content += m.formatLogEntry(entry) + "\n"
	}
	m.viewport.SetContent(content)
	m.viewport.GotoBottom()
}

func (m CLIWorkflowModel) formatLogEntry(entry LogEntry) string {
	timeStyle := lipgloss.NewStyle().Foreground(ColorMuted)
	levelStyle := m.getLevelStyle(entry.Level)
	messageStyle := lipgloss.NewStyle().Foreground(ColorText)

	timeStr := timeStyle.Render(fmt.Sprintf("[%s]", entry.Time.Format("15:04:05")))
	levelStr := levelStyle.Render(fmt.Sprintf("[%s]", entry.Level))
	messageStr := messageStyle.Render(entry.Message)

	return fmt.Sprintf("%s %s %s", timeStr, levelStr, messageStr)
}

func (m CLIWorkflowModel) getLevelStyle(level string) lipgloss.Style {
	switch level {
	case "SUCCESS":
		return lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true)
	case "ERROR":
		return lipgloss.NewStyle().Foreground(ColorError).Bold(true)
	case "PHASE":
		return lipgloss.NewStyle().Foreground(ColorSecondary).Bold(true)
	case "WARNING":
		return lipgloss.NewStyle().Foreground(ColorWarning).Bold(true)
	default:
		return lipgloss.NewStyle().Foreground(ColorInfo).Bold(true)
	}
}

func (m CLIWorkflowModel) View() string {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorText).
		Background(ColorBackground).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary)
	subtitleStyle := lipgloss.NewStyle().Foreground(ColorMuted)
	phaseStyle := lipgloss.NewStyle().Foreground(ColorSecondary).Bold(true)

	appName := getAppName()
	header := headerStyle.Render(fmt.Sprintf("%s - 红队工作流", appName))

	content := ""
	switch m.state {
	case "input":
		content = m.viewInputState(titleStyle, subtitleStyle)
	case "running":
		content = m.viewRunningState(titleStyle, subtitleStyle, phaseStyle)
	case "paused":
		content = m.viewPausedState(titleStyle, subtitleStyle)
	case "completed":
		content = m.viewCompletedState(titleStyle, subtitleStyle)
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(fmt.Sprintf("%s\n\n%s", header, content))
}

func (m CLIWorkflowModel) viewInputState(titleStyle, subtitleStyle lipgloss.Style) string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorSurface).
		Padding(2, 3).
		Background(ColorBackground)

	title := titleStyle.Render("[+] 目标设置")
	subtitle := subtitleStyle.Render("请输入目标 URL 或 IP 地址")
	input := m.targetInput.View()
	footer := subtitleStyle.Render("ENTER 开始 | ESC 退出")

	return boxStyle.Render(fmt.Sprintf(
		"%s\n%s\n\n%s\n\n%s\n\n%s",
		title,
		m.createSeparator(),
		subtitle,
		input,
		footer,
	))
}

func (m CLIWorkflowModel) viewRunningState(titleStyle, subtitleStyle, phaseStyle lipgloss.Style) string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorSurface).
		Padding(2, 3).
		Background(ColorBackground)

	var result strings.Builder

	result.WriteString(fmt.Sprintf("%s %s\n\n", m.spinner.View(), phaseStyle.Render("当前阶段: "+m.currentPhase)))

	result.WriteString(titleStyle.Render("[*] 整体进度") + "\n")
	result.WriteString(m.prog.ViewAs(m.progressPercent) + "\n\n")

	result.WriteString(titleStyle.Render("[*] 当前阶段进度") + "\n")
	result.WriteString(m.secondaryProg.ViewAs(m.secondaryProgressPercent) + "\n\n")

	result.WriteString(titleStyle.Render("[+] 执行日志") + "\n")
	result.WriteString(m.viewport.View() + "\n")

	if len(m.successes) > 0 {
		successStyle := lipgloss.NewStyle().Foreground(ColorSuccess)
		result.WriteString("\n" + titleStyle.Render("[OK] 成功信息") + "\n")
		for _, success := range m.successes {
			result.WriteString(successStyle.Render("  [OK] "+success) + "\n")
		}
	}

	if len(m.errors) > 0 {
		errorStyle := lipgloss.NewStyle().Foreground(ColorError)
		result.WriteString("\n" + titleStyle.Render("[ERR] 错误信息") + "\n")
		for _, err := range m.errors {
			result.WriteString(errorStyle.Render("  [ERR] "+err) + "\n")
		}
	}

	footer := subtitleStyle.Render("P 暂停 | CTRL+C 退出")
	result.WriteString("\n" + footer)

	return boxStyle.Render(result.String())
}

func (m CLIWorkflowModel) viewPausedState(titleStyle, subtitleStyle lipgloss.Style) string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorWarning).
		Padding(2, 3).
		Background(ColorBackground)

	title := titleStyle.Render("[PAUSE] 工作流已暂停")
	phase := subtitleStyle.Render("当前阶段: " + m.currentPhase)

	var result strings.Builder
	result.WriteString(fmt.Sprintf("%s\n%s\n\n", title, phase))
	result.WriteString(m.prog.ViewAs(m.progressPercent) + "\n\n")
	result.WriteString(titleStyle.Render("[+] 执行日志") + "\n")
	result.WriteString(m.viewport.View() + "\n")

	footer := subtitleStyle.Render("P 恢复 | CTRL+C 退出")
	result.WriteString("\n" + footer)

	return boxStyle.Render(result.String())
}

func (m CLIWorkflowModel) viewCompletedState(titleStyle, subtitleStyle lipgloss.Style) string {
	var boxStyle lipgloss.Style
	if len(m.errors) > 0 {
		boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorError).
			Padding(2, 3).
			Background(ColorBackground)
	} else {
		boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSuccess).
			Padding(2, 3).
			Background(ColorBackground)
	}

	title := titleStyle.Render("[OK] 工作流执行完成")

	var result strings.Builder
	result.WriteString(title + "\n\n")
	result.WriteString(titleStyle.Render("[*] 最终进度") + "\n")
	result.WriteString(m.prog.ViewAs(m.progressPercent) + "\n\n")
	result.WriteString(titleStyle.Render("[+] 执行日志") + "\n")
	result.WriteString(m.viewport.View() + "\n")

	if len(m.successes) > 0 {
		successStyle := lipgloss.NewStyle().Foreground(ColorSuccess)
		result.WriteString("\n" + titleStyle.Render("[OK] 成功信息") + "\n")
		for _, success := range m.successes {
			result.WriteString(successStyle.Render("  [OK] "+success) + "\n")
		}
	}

	if len(m.errors) > 0 {
		errorStyle := lipgloss.NewStyle().Foreground(ColorError)
		result.WriteString("\n" + titleStyle.Render("[ERR] 错误信息") + "\n")
		for _, err := range m.errors {
			result.WriteString(errorStyle.Render("  [ERR] "+err) + "\n")
		}
	}

	footer := subtitleStyle.Render("C 重新开始 | CTRL+C 退出")
	result.WriteString("\n" + footer)

	return boxStyle.Render(result.String())
}

func (m CLIWorkflowModel) createSeparator() string {
	return lipgloss.NewStyle().Foreground(ColorSurface).Render("-" + strings.Repeat("-", 50) + "-")
}

type WorkflowPhaseMsg struct {
	Phase string
}

type WorkflowLogMsg struct {
	Log string
}

type WorkflowSuccessMsg struct {
	Message string
}

type WorkflowErrorMsg struct {
	Error string
}

type WorkflowProgressMsg struct {
	Progress  float64
	Secondary float64
}

type WorkflowDoneMsg struct{}
