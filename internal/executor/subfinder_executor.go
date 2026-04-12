package executor

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"nezha_sec/internal/model"
)

// SubfinderExecutor Subfinder工具执行器
// 用于执行Subfinder命令进行域名枚举
type SubfinderExecutor struct {}

// NewSubfinderExecutor 创建一个新的Subfinder执行器
// 返回值列表：
//   - *SubfinderExecutor Subfinder执行器实例
//   - error 初始化过程中可能发生的错误
func NewSubfinderExecutor() (*SubfinderExecutor, error) {
	return &SubfinderExecutor{}, nil
}

// Execute 执行Subfinder命令
// 参数列表：
//   - ctx 上下文，用于控制执行超时和取消
//   - arguments 执行参数，包含domain等
// 返回值列表：
//   - *model.ExecutionResult 执行结果
//   - error 执行过程中可能发生的错误
func (e *SubfinderExecutor) Execute(ctx context.Context, arguments map[string]interface{}) (*model.ExecutionResult, error) {
	startTime := time.Now()

	// 获取域名参数
	domain, ok := arguments["domain"].(string)
	if !ok || domain == "" {
		return &model.ExecutionResult{
			Success: false,
			ToolName: e.GetToolName(),
			Error: "缺少domain参数",
			ExecutionTime: time.Since(startTime).Milliseconds(),
		}, nil
	}

	// 构建Subfinder命令
	cmdArgs := []string{"-d", domain}

	// 添加其他参数
	if silent, ok := arguments["silent"].(bool); ok && silent {
		cmdArgs = append(cmdArgs, "-silent")
	}

	if all, ok := arguments["all"].(bool); ok && all {
		cmdArgs = append(cmdArgs, "-all")
	}

	// 执行Subfinder命令
	cmd := exec.CommandContext(ctx, "subfinder", cmdArgs...)
	output, err := cmd.CombinedOutput()

	// 构建结果
	result := &model.ExecutionResult{
		Success: err == nil,
		ToolName: e.GetToolName(),
		Output: string(output),
		ExecutionTime: time.Since(startTime).Milliseconds(),
		StructuredOutput: map[string]interface{}{
			"domain": domain,
			"arguments": arguments,
			"output": string(output),
			"error": err != nil,
		},
	}

	if err != nil {
		result.Error = fmt.Sprintf("执行Subfinder命令失败: %v", err)
	}

	return result, nil
}

// GetToolName 获取工具名称
// 返回值列表：
//   - string 工具名称
func (e *SubfinderExecutor) GetToolName() string {
	return "Subfinder"
}
