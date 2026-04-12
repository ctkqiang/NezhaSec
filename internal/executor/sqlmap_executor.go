package executor

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"nezha_sec/internal/model"
)

// SqlmapExecutor sqlmap工具执行器
// 用于执行sqlmap命令进行SQL注入检测
type SqlmapExecutor struct {}

// NewSqlmapExecutor 创建一个新的sqlmap执行器
// 返回值列表：
//   - *SqlmapExecutor sqlmap执行器实例
//   - error 初始化过程中可能发生的错误
func NewSqlmapExecutor() (*SqlmapExecutor, error) {
	return &SqlmapExecutor{}, nil
}

// Execute 执行sqlmap命令
// 参数列表：
//   - ctx 上下文，用于控制执行超时和取消
//   - arguments 执行参数，包含url、risk、level等
// 返回值列表：
//   - *model.ExecutionResult 执行结果
//   - error 执行过程中可能发生的错误
func (e *SqlmapExecutor) Execute(ctx context.Context, arguments map[string]interface{}) (*model.ExecutionResult, error) {
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

	// 构建sqlmap命令
	cmdArgs := []string{"-u", url}

	// 添加风险参数
	if risk, ok := arguments["risk"].(string); ok && risk != "" {
		cmdArgs = append(cmdArgs, "--risk", risk)
	} else {
		// 默认风险等级为1
		cmdArgs = append(cmdArgs, "--risk", "1")
	}

	// 添加级别参数
	if level, ok := arguments["level"].(string); ok && level != "" {
		cmdArgs = append(cmdArgs, "--level", level)
	} else {
		// 默认级别为1
		cmdArgs = append(cmdArgs, "--level", "1")
	}

	// 添加其他参数
	if batch, ok := arguments["batch"].(bool); ok && batch {
		cmdArgs = append(cmdArgs, "--batch")
	} else {
		// 默认使用batch模式
		cmdArgs = append(cmdArgs, "--batch")
	}

	if threads, ok := arguments["threads"].(string); ok && threads != "" {
		cmdArgs = append(cmdArgs, "--threads", threads)
	}

	// 执行sqlmap命令
	cmd := exec.CommandContext(ctx, "sqlmap", cmdArgs...)
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
		result.Error = fmt.Sprintf("执行sqlmap命令失败: %v", err)
	}

	return result, nil
}

// GetToolName 获取工具名称
// 返回值列表：
//   - string 工具名称
func (e *SqlmapExecutor) GetToolName() string {
	return "sqlmap"
}
