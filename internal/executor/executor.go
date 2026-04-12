package executor

import (
	"context"
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
