package executor

import (
	"context"
	"fmt"
	"time"

	"nezha_sec/internal/model"
)

type ToolExecutorInterface interface {
	Execute(ctx context.Context, arguments map[string]interface{}) (*model.ExecutionResult, error)

	GetToolName() string
}

type DefaultExecutor struct {
	toolName string
}

func NewDefaultExecutor(toolName string) *DefaultExecutor {
	return &DefaultExecutor{
		toolName: toolName,
	}
}

func (e *DefaultExecutor) Execute(ctx context.Context, arguments map[string]interface{}) (*model.ExecutionResult, error) {
	startTime := time.Now()

	result := &model.ExecutionResult{
		Success:       true,
		ToolName:      e.GetToolName(),
		Output:        fmt.Sprintf("工具 %s 未实现，模拟执行成功", e.toolName),
		ExecutionTime: time.Since(startTime).Milliseconds(),
		StructuredOutput: map[string]interface{}{
			"tool_name": e.toolName,
			"arguments": arguments,
			"status":    "simulated",
		},
	}

	return result, nil
}

func (e *DefaultExecutor) GetToolName() string {
	return e.toolName
}
