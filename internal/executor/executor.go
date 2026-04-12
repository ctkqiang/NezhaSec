package executor

import (
	"context"
	"fmt"
	"time"

	"nezha_sec/internal/model"
)

// ToolExecutorInterface 工具执行器接口
// 所有工具执行器必须实现此接口

type ToolExecutorInterface interface {
	// Execute 执行工具并返回结果
	// 参数列表：
	//   - ctx 上下文，用于控制执行超时和取消
	//   - arguments 工具执行参数
	// 返回值列表：
	//   - *model.ExecutionResult 执行结果
	//   - error 执行过程中可能发生的错误
	Execute(ctx context.Context, arguments map[string]interface{}) (*model.ExecutionResult, error)

	// GetToolName 获取工具名称
	// 返回值列表：
	//   - string 工具名称
	GetToolName() string
}

// DefaultExecutor 默认工具执行器
// 用于处理未注册的工具

type DefaultExecutor struct {
	toolName string
}

// NewDefaultExecutor 创建一个新的默认工具执行器
// 参数列表：
//   - toolName 工具名称
// 返回值列表：
//   - *DefaultExecutor 默认工具执行器实例
func NewDefaultExecutor(toolName string) *DefaultExecutor {
	return &DefaultExecutor{
		toolName: toolName,
	}
}

// Execute 执行工具并返回结果
// 参数列表：
//   - ctx 上下文，用于控制执行超时和取消
//   - arguments 工具执行参数
// 返回值列表：
//   - *model.ExecutionResult 执行结果
//   - error 执行过程中可能发生的错误
func (e *DefaultExecutor) Execute(ctx context.Context, arguments map[string]interface{}) (*model.ExecutionResult, error) {
	startTime := time.Now()

	// 构建结果
	result := &model.ExecutionResult{
		Success: true,
		ToolName: e.GetToolName(),
		Output: fmt.Sprintf("工具 %s 未实现，模拟执行成功", e.toolName),
		ExecutionTime: time.Since(startTime).Milliseconds(),
		StructuredOutput: map[string]interface{}{
			"tool_name": e.toolName,
			"arguments": arguments,
			"status": "simulated",
		},
	}

	return result, nil
}

// GetToolName 获取工具名称
// 返回值列表：
//   - string 工具名称
func (e *DefaultExecutor) GetToolName() string {
	return e.toolName
}
