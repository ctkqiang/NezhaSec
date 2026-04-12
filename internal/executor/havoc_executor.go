package executor

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"nezha_sec/internal/model"
)

// HavocExecutor Havoc工具执行器
// 用于执行Havoc命令进行C2操作
type HavocExecutor struct {}

// NewHavocExecutor 创建一个新的Havoc执行器
// 返回值列表：
//   - *HavocExecutor Havoc执行器实例
//   - error 初始化过程中可能发生的错误
func NewHavocExecutor() (*HavocExecutor, error) {
	return &HavocExecutor{}, nil
}

// Execute 执行Havoc命令
// 参数列表：
//   - ctx 上下文，用于控制执行超时和取消
//   - arguments 执行参数，包含command等
// 返回值列表：
//   - *model.ExecutionResult 执行结果
//   - error 执行过程中可能发生的错误
func (e *HavocExecutor) Execute(ctx context.Context, arguments map[string]interface{}) (*model.ExecutionResult, error) {
	startTime := time.Now()

	// 获取命令参数
	command, ok := arguments["command"].(string)
	if !ok || command == "" {
		return &model.ExecutionResult{
			Success: false,
			ToolName: e.GetToolName(),
			Error: "缺少command参数",
			ExecutionTime: time.Since(startTime).Milliseconds(),
		}, nil
	}

	// 构建Havoc命令
	cmdArgs := []string{"-c", command}

	// 执行Havoc命令
	cmd := exec.CommandContext(ctx, "havoc", cmdArgs...)
	output, err := cmd.CombinedOutput()

	// 构建结果
	result := &model.ExecutionResult{
		Success: err == nil,
		ToolName: e.GetToolName(),
		Output: string(output),
		ExecutionTime: time.Since(startTime).Milliseconds(),
		StructuredOutput: map[string]interface{}{
			"command": command,
			"arguments": arguments,
			"output": string(output),
			"error": err != nil,
		},
	}

	if err != nil {
		result.Error = fmt.Sprintf("执行Havoc命令失败: %v", err)
	}

	return result, nil
}

// GetToolName 获取工具名称
// 返回值列表：
//   - string 工具名称
func (e *HavocExecutor) GetToolName() string {
	return "Havoc"
}
