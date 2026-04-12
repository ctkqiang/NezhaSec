package executor

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"nezha_sec/internal/model"
)

// AmassExecutor Amass工具执行器
// 用于执行Amass命令进行域名枚举和资产发现
type AmassExecutor struct {}

// NewAmassExecutor 创建一个新的Amass执行器
// 返回值列表：
//   - *AmassExecutor Amass执行器实例
//   - error 初始化过程中可能发生的错误
func NewAmassExecutor() (*AmassExecutor, error) {
	return &AmassExecutor{}, nil
}

// Execute 执行Amass命令
// 参数列表：
//   - ctx 上下文，用于控制执行超时和取消
//   - arguments 执行参数，包含domain等
// 返回值列表：
//   - *model.ExecutionResult 执行结果
//   - error 执行过程中可能发生的错误
func (e *AmassExecutor) Execute(ctx context.Context, arguments map[string]interface{}) (*model.ExecutionResult, error) {
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

	// 构建Amass命令
	cmdArgs := []string{"enum", "-d", domain}

	// 添加其他参数
	if passive, ok := arguments["passive"].(bool); ok && passive {
		cmdArgs = append(cmdArgs, "-passive")
	}

	if active, ok := arguments["active"].(bool); ok && active {
		cmdArgs = append(cmdArgs, "-active")
	}

	// 执行Amass命令
	cmd := exec.CommandContext(ctx, "amass", cmdArgs...)
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
		result.Error = fmt.Sprintf("执行Amass命令失败: %v", err)
	}

	return result, nil
}

// GetToolName 获取工具名称
// 返回值列表：
//   - string 工具名称
func (e *AmassExecutor) GetToolName() string {
	return "Amass"
}
