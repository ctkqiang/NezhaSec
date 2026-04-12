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

// Orchestrator 核心调度器
// 负责管理状态机、与DeepSeek API交互、调度工具执行
type Orchestrator struct {
	// 工具注册表
	toolRegistry *registry.ToolRegistry
	
	// 当前状态
	state string
	
	// 目标URL
	targetURL string
	
	// 执行步骤历史
	executionSteps []model.ExecutionResult
	
	// 上下文存储
	contextStore map[string]interface{}
}

// NewOrchestrator 创建一个新的调度器
// 参数列表：
//   - toolRegistry 工具注册表
// 返回值列表：
//   - *Orchestrator 调度器实例
//   - error 初始化过程中可能发生的错误
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

// SetTarget 设置目标URL
// 参数列表：
//   - url 目标URL
//   - allowLocal 是否允许扫描本地地址
// 返回值列表：
//   - bool 是否设置成功
//   - string 失败原因
func (o *Orchestrator) SetTarget(url string, allowLocal bool) (bool, string) {
	// 检查目标是否允许扫描
	allowed, reason := security.IsTargetAllowed(url, allowLocal)
	if !allowed {
		return false, reason
	}

	o.targetURL = url
	o.contextStore["target_url"] = url
	return true, ""
}

// GetState 获取当前状态
// 返回值列表：
//   - string 当前状态
func (o *Orchestrator) GetState() string {
	return o.state
}

// GetExecutionSteps 获取执行步骤历史
// 返回值列表：
//   - []model.ExecutionResult 执行步骤历史
func (o *Orchestrator) GetExecutionSteps() []model.ExecutionResult {
	return o.executionSteps
}

// GetContextStore 获取上下文存储
// 返回值列表：
//   - map[string]interface{} 上下文存储
func (o *Orchestrator) GetContextStore() map[string]interface{} {
	return o.contextStore
}

// StartAnalysis 开始分析
// 返回值列表：
//   - api.ProgressMsg 进度消息
func (o *Orchestrator) StartAnalysis() api.ProgressMsg {
	o.state = "analyzing"
	return api.ProgressMsg("开始分析目标URL...")
}

// AnalyzeWithAI 使用AI分析目标
// 返回值列表：
//   - string AI分析结果
//   - error 可能的错误
func (o *Orchestrator) AnalyzeWithAI() (string, error) {
	// 构建上下文信息
	contextInfo := fmt.Sprintf("目标URL: %s\n", o.targetURL)
	contextInfo += "已执行步骤:\n"
	for _, step := range o.executionSteps {
		contextInfo += fmt.Sprintf("- %s: %s\n", step.ToolName, step.Output)
	}

	// 调用DeepSeek API
	apiKey := api.GetDeepSeekAPIKey()
	if apiKey == "" {
		return "", fmt.Errorf("未配置DeepSeek API密钥")
	}

	// 构建分析请求
	result, err := api.AnalyzeURLWithDeepSeek(
		o.targetURL,
		apiKey,
	)

	if err != nil {
		return "", fmt.Errorf("AI分析失败: %w", err)
	}

	return result, nil
}

// ExecuteTool 执行工具
// 参数列表：
//   - toolName 工具名称
//   - arguments 工具参数
// 返回值列表：
//   - *model.ExecutionResult 执行结果
//   - error 可能的错误
func (o *Orchestrator) ExecuteTool(toolName string, arguments map[string]interface{}) (*model.ExecutionResult, error) {
	// 获取工具
	tool, exists := o.toolRegistry.GetTool(toolName)
	if !exists {
		// 工具不存在，使用默认工具执行器模拟执行
		defaultExecutor := executor.NewDefaultExecutor(toolName)
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		result, err := defaultExecutor.Execute(ctx, arguments)
		if err != nil {
			return nil, fmt.Errorf("工具执行失败: %w", err)
		}
		// 保存执行结果
		o.executionSteps = append(o.executionSteps, *result)
		// 更新上下文
		o.contextStore[toolName] = result
		return result, nil
	}

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 执行工具
	result, err := tool.Execute(ctx, arguments)
	if err != nil {
		return nil, fmt.Errorf("工具执行失败: %w", err)
	}

	// 保存执行结果
	o.executionSteps = append(o.executionSteps, *result)

	// 更新上下文
	o.contextStore[toolName] = result

	return result, nil
}

// CompleteAnalysis 完成分析
// 返回值列表：
//   - api.ProgressMsg 进度消息
func (o *Orchestrator) CompleteAnalysis() api.ProgressMsg {
	o.state = "completed"
	return api.ProgressMsg("分析完成")
}
