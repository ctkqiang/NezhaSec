package executor

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"nezha_sec/internal/model"
)

// HttpxExecutor Httpx工具执行器
// 用于执行Httpx命令进行HTTP探测
type HttpxExecutor struct {}

// NewHttpxExecutor 创建一个新的Httpx执行器
// 返回值列表：
//   - *HttpxExecutor Httpx执行器实例
//   - error 初始化过程中可能发生的错误
func NewHttpxExecutor() (*HttpxExecutor, error) {
	return &HttpxExecutor{}, nil
}

// Execute 执行Httpx命令
// 参数列表：
//   - ctx 上下文，用于控制执行超时和取消
//   - arguments 执行参数，包含target等
// 返回值列表：
//   - *model.ExecutionResult 执行结果
//   - error 执行过程中可能发生的错误
func (e *HttpxExecutor) Execute(ctx context.Context, arguments map[string]interface{}) (*model.ExecutionResult, error) {
	startTime := time.Now()

	// 获取目标参数
	target, ok := arguments["target"].(string)
	if !ok || target == "" {
		return &model.ExecutionResult{
			Success: false,
			ToolName: e.GetToolName(),
			Error: "缺少target参数",
			ExecutionTime: time.Since(startTime).Milliseconds(),
		}, nil
	}

	// 构建Httpx命令
	cmdArgs := []string{"-u", target}

	// 添加其他参数
	if silent, ok := arguments["silent"].(bool); ok && silent {
		cmdArgs = append(cmdArgs, "-silent")
	}

	if status, ok := arguments["status"].(bool); ok && status {
		cmdArgs = append(cmdArgs, "-status-code")
	}

	if title, ok := arguments["title"].(bool); ok && title {
		cmdArgs = append(cmdArgs, "-title")
	}

	// 执行Httpx命令
	cmd := exec.CommandContext(ctx, "httpx", cmdArgs...)
	output, err := cmd.CombinedOutput()

	// 构建结果
	result := &model.ExecutionResult{
		Success: err == nil,
		ToolName: e.GetToolName(),
		Output: string(output),
		ExecutionTime: time.Since(startTime).Milliseconds(),
		StructuredOutput: map[string]interface{}{
			"target": target,
			"arguments": arguments,
			"output": string(output),
			"error": err != nil,
		},
	}

	if err != nil {
		result.Error = fmt.Sprintf("执行Httpx命令失败: %v", err)
	}

	return result, nil
}

// GetToolName 获取工具名称
// 返回值列表：
//   - string 工具名称
func (e *HttpxExecutor) GetToolName() string {
	return "Httpx"
}
