package orchestrator

import (
	"context"
	"fmt"
	"time"

	"nezha_sec/internal/api"
	"nezha_sec/internal/executor"
	"nezha_sec/internal/model"
	"nezha_sec/internal/registry"
	"nezha_sec/internal/security"
)

type Orchestrator struct {
	toolRegistry *registry.ToolRegistry

	state string

	targetURL string

	executionSteps []model.ExecutionResult

	contextStore map[string]interface{}
}

func NewOrchestrator(toolRegistry *registry.ToolRegistry) (*Orchestrator, error) {
	if toolRegistry == nil {
		return nil, fmt.Errorf("工具注册表不能为空")
	}

	return &Orchestrator{
		toolRegistry:   toolRegistry,
		state:          "initial",
		executionSteps: []model.ExecutionResult{},
		contextStore:   make(map[string]interface{}),
	}, nil
}

func (o *Orchestrator) SetTarget(url string, allowLocal bool) (bool, string) {

	allowed, reason := security.IsTargetAllowed(url, allowLocal)
	if !allowed {
		return false, reason
	}

	o.targetURL = url
	o.contextStore["target_url"] = url
	return true, ""
}

func (o *Orchestrator) GetState() string {
	return o.state
}

func (o *Orchestrator) GetExecutionSteps() []model.ExecutionResult {
	return o.executionSteps
}

func (o *Orchestrator) GetContextStore() map[string]interface{} {
	return o.contextStore
}

func (o *Orchestrator) StartAnalysis() api.ProgressMsg {
	o.state = "analyzing"
	return api.ProgressMsg("开始分析目标URL...")
}

func (o *Orchestrator) AnalyzeWithAI() (string, error) {

	contextInfo := fmt.Sprintf("目标URL: %s\n", o.targetURL)
	contextInfo += "已执行步骤:\n"
	for _, step := range o.executionSteps {
		contextInfo += fmt.Sprintf("- %s: %s\n", step.ToolName, step.Output)
	}

	apiKey := api.GetDeepSeekAPIKey()
	if apiKey == "" {
		return "", fmt.Errorf("未配置DeepSeek API密钥")
	}

	result, err := api.AnalyzeURLWithDeepSeek(
		o.targetURL,
		apiKey,
	)

	if err != nil {
		return "", fmt.Errorf("AI分析失败: %w", err)
	}

	return result, nil
}

func (o *Orchestrator) ExecuteTool(toolName string, arguments map[string]interface{}) (*model.ExecutionResult, error) {

	tool, exists := o.toolRegistry.GetTool(toolName)
	if !exists {

		defaultExecutor := executor.NewDefaultExecutor(toolName)
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		result, err := defaultExecutor.Execute(ctx, arguments)
		if err != nil {
			return nil, fmt.Errorf("工具执行失败: %w", err)
		}

		o.executionSteps = append(o.executionSteps, *result)

		o.contextStore[toolName] = result
		return result, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	result, err := tool.Execute(ctx, arguments)
	if err != nil {
		return nil, fmt.Errorf("工具执行失败: %w", err)
	}

	o.executionSteps = append(o.executionSteps, *result)

	o.contextStore[toolName] = result

	return result, nil
}

func (o *Orchestrator) CompleteAnalysis() api.ProgressMsg {
	o.state = "completed"
	return api.ProgressMsg("分析完成")
}

func (o *Orchestrator) NewWorkflowManager() *WorkflowManager {
	return NewWorkflowManager(o)
}
