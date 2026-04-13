package views

import (
	"context"
	"fmt"
	"strings"
	"time"

	"nezha_sec/internal/orchestrator"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type CLIWorkflowModel struct {
	orchestrator    *orchestrator.Orchestrator
	workflowManager *orchestrator.WorkflowManager

	cancelFunc context.CancelFunc

	targetInput textinput.Model

	spinner spinner.Model
	progress progress.Model

	viewport viewport.Model

	state        string
	currentPhase string
	progressPercent float64

	logs []string

	errors []string

	config map[string]interface{}
}

func NewCLIWorkflowModel(orchestrator *orchestrator.Orchestrator) (CLIWorkflowModel, error) {
	workflowManager := orchestrator.NewWorkflowManager()

	spinner := spinner.New()
	spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))

	progress := progress.New()

	viewport := viewport.New(20, 80)
	viewport.Style = lipgloss.NewStyle().Border(lipgloss.NormalBorder())

	targetInput := textinput.New()
	targetInput.Placeholder = "输入目标 URL 或 IP 地址"
	targetInput.Focus()

	return CLIWorkflowModel{
		orchestrator:    orchestrator,
		workflowManager: workflowManager,
		targetInput:     targetInput,
		spinner:         spinner,
		progress:        progress,
		viewport:        viewport,
		state:           "input",
		logs:            []string{},
		errors:          []string{},
		config: map[string]interface{}{
			"aggressive": false,
			"timeout":    300,
			"threads":    10,
		},
	}, nil
}

func (m CLIWorkflowModel) Init() tea.Cmd {
	return textinput.Blink
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
				m.logs = append(m.logs, fmt.Sprintf("开始渗透测试工作流，目标: %s", m.targetInput.Value()))

				ctx, cancel := context.WithCancel(context.Background())
				m.cancelFunc = cancel

				return m, tea.Batch(
					m.spinner.Tick,
					func() tea.Msg {
						err := m.workflowManager.ExecuteWorkflow(ctx)
						if err != nil {
							m.errors = append(m.errors, fmt.Sprintf("工作流执行失败: %v", err))
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
				m.logs = []string{}
				m.errors = []string{}
				m.progressPercent = 0
				m.currentPhase = ""
				return m, nil
			}
		}

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case WorkflowPhaseMsg:
		m.currentPhase = msg.Phase
		m.logs = append(m.logs, fmt.Sprintf("[%s] 开始阶段: %s", time.Now().Format("15:04:05"), msg.Phase))
		m.viewport.SetContent(strings.Join(m.logs, "\n"))
		m.viewport.GotoBottom()
		return m, nil

	case WorkflowLogMsg:
		m.logs = append(m.logs, fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), msg.Log))
		m.viewport.SetContent(strings.Join(m.logs, "\n"))
		m.viewport.GotoBottom()
		return m, nil

	case WorkflowErrorMsg:
		m.errors = append(m.errors, fmt.Sprintf("[%s] 错误: %s", time.Now().Format("15:04:05"), msg.Error))
		m.viewport.SetContent(strings.Join(append(m.logs, m.errors...), "\n"))
		m.viewport.GotoBottom()
		return m, nil

	case WorkflowProgressMsg:
		m.progressPercent = msg.Progress
		return m, nil

	case WorkflowDoneMsg:
		m.state = "completed"
		m.logs = append(m.logs, fmt.Sprintf("[%s] 工作流执行完成", time.Now().Format("15:04:05")))
		m.viewport.SetContent(strings.Join(m.logs, "\n"))
		m.viewport.GotoBottom()
		return m, nil
	}

	if m.state == "input" {
		m.targetInput, cmd = m.targetInput.Update(msg)
	} else if m.state == "running" {
		m.spinner, cmd = m.spinner.Update(msg)
	}

	return m, cmd
}

func (m CLIWorkflowModel) View() string {
	baseStyle := lipgloss.NewStyle().Padding(1, 2)
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")).Background(lipgloss.Color("235"))
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("13"))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("160"))
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	phaseStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)

	header := headerStyle.Render("🔐 哪吒网络安全分析器 - 红队工作流")

	content := ""
	switch m.state {
	case "input":
		title := titleStyle.Render("🎯 目标设置")
		subtitle := infoStyle.Render("请输入目标 URL 或 IP 地址")
		input := m.targetInput.View()
		footer := infoStyle.Render("ENTER 开始 • ESC 退出")

		separator := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("─" + strings.Repeat("─", 60) + "─")

		content = fmt.Sprintf(
			"%s\n%s\n\n%s\n\n%s\n\n%s",
			title,
			separator,
			subtitle,
			input,
			footer,
		)

	case "running":
		s := fmt.Sprintf("%s %s\n\n", m.spinner.View(), phaseStyle.Render("当前阶段: " + m.currentPhase))

		s += titleStyle.Render("📊 执行进度") + "\n"
		s += m.progress.ViewAs(m.progressPercent) + "\n\n"

		s += titleStyle.Render("📋 执行日志") + "\n"
		s += m.viewport.View() + "\n"

		if len(m.errors) > 0 {
			s += "\n" + titleStyle.Render("❌ 错误信息") + "\n"
			for _, err := range m.errors {
				s += errorStyle.Render(err) + "\n"
			}
		}

		footer := infoStyle.Render("P 暂停 • CTRL+C 退出")
		s += "\n" + footer

		content = s

	case "paused":
		title := titleStyle.Render("⏸️  工作流已暂停")
		s := title + "\n\n"

		s += titleStyle.Render("当前阶段: " + m.currentPhase) + "\n\n"

		s += titleStyle.Render("📋 执行日志") + "\n"
		s += m.viewport.View() + "\n"

		footer := infoStyle.Render("P 恢复 • CTRL+C 退出")
		s += "\n" + footer

		content = s

	case "completed":
		title := titleStyle.Render("✅ 工作流执行完成")
		s := title + "\n\n"

		s += titleStyle.Render("📋 执行日志") + "\n"
		s += m.viewport.View() + "\n"

		if len(m.errors) > 0 {
			s += "\n" + titleStyle.Render("❌ 错误信息") + "\n"
			for _, err := range m.errors {
				s += errorStyle.Render(err) + "\n"
			}
		}

		footer := infoStyle.Render("C 重新开始 • CTRL+C 退出")
		s += "\n" + footer

		content = s
	}

	return baseStyle.Render(fmt.Sprintf("%s\n\n%s", header, content))
}

type WorkflowPhaseMsg struct {
	Phase string
}

type WorkflowLogMsg struct {
	Log string
}

type WorkflowErrorMsg struct {
	Error string
}

type WorkflowProgressMsg struct {
	Progress float64
}

type WorkflowDoneMsg struct {}
