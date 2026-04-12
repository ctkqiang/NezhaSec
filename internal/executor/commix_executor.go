package executor

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"nezha_sec/internal/model"
)

// CommixExecutor Commix工具执行器
// 用于执行Commix命令进行命令注入漏洞检测
type CommixExecutor struct {}

// NewCommixExecutor 创建一个新的Commix执行器
// 返回值列表：
//   - *CommixExecutor Commix执行器实例
//   - error 初始化过程中可能发生的错误
func NewCommixExecutor() (*CommixExecutor, error) {
	return &CommixExecutor{}, nil
}

// Execute 执行Commix命令
// 参数列表：
//   - ctx 上下文，用于控制执行超时和取消
//   - arguments 执行参数，包含url等
// 返回值列表：
//   - *model.ExecutionResult 执行结果
//   - error 执行过程中可能发生的错误
func (e *CommixExecutor) Execute(ctx context.Context, arguments map[string]interface{}) (*model.ExecutionResult, error) {
	startTime := time.Now()

	// 获取URL参数
	url, ok := arguments["url"].(string)
	if !ok || url == "" {
		return &model.ExecutionResult{
			Success: false,
			ToolName: e.GetToolName(),
			Error: "缺少url参数",
			ExecutionTime: time.Since(startTime).Milliseconds(),
		}, nil
	}

	// 构建Commix命令
	cmdArgs := []string{"--url", url}

	// 添加其他参数
	if batch, ok := arguments["batch"].(bool); ok && batch {
		cmdArgs = append(cmdArgs, "--batch")
	} else {
		// 默认使用batch模式
		cmdArgs = append(cmdArgs, "--batch")
	}

	// 执行Commix命令
	cmd := exec.CommandContext(ctx, "commix", cmdArgs...)
	output, err := cmd.CombinedOutput()

	// 构建结果
	result := &model.ExecutionResult{
		Success: err == nil,
		ToolName: e.GetToolName(),
		Output: string(output),
		ExecutionTime: time.Since(startTime).Milliseconds(),
		StructuredOutput: map[string]interface{}{
			"url": url,
			"arguments": arguments,
			"output": string(output),
			"error": err != nil,
		},
	}

	if err != nil {
		result.Error = fmt.Sprintf("执行Commix命令失败: %v", err)
	}

	return result, nil
}

// GetToolName 获取工具名称
// 返回值列表：
//   - string 工具名称
func (e *CommixExecutor) GetToolName() string {
	return "Commix"
}
