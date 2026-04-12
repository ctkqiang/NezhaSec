package executor

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"nezha_sec/internal/model"
)

// FfufExecutor Ffuf工具执行器
// 用于执行Ffuf命令进行模糊测试和目录扫描
type FfufExecutor struct {}

// NewFfufExecutor 创建一个新的Ffuf执行器
// 返回值列表：
//   - *FfufExecutor Ffuf执行器实例
//   - error 初始化过程中可能发生的错误
func NewFfufExecutor() (*FfufExecutor, error) {
	return &FfufExecutor{}, nil
}

// Execute 执行Ffuf命令
// 参数列表：
//   - ctx 上下文，用于控制执行超时和取消
//   - arguments 执行参数，包含url等
// 返回值列表：
//   - *model.ExecutionResult 执行结果
//   - error 执行过程中可能发生的错误
func (e *FfufExecutor) Execute(ctx context.Context, arguments map[string]interface{}) (*model.ExecutionResult, error) {
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

	// 构建Ffuf命令
	cmdArgs := []string{"-u", url}

	// 添加词表参数
	if wordlist, ok := arguments["wordlist"].(string); ok && wordlist != "" {
		cmdArgs = append(cmdArgs, "-w", wordlist)
	} else {
		// 默认使用common.txt词表
		cmdArgs = append(cmdArgs, "-w", "/usr/share/wordlists/dirb/common.txt")
	}

	// 添加其他参数
	if silent, ok := arguments["silent"].(bool); ok && silent {
		cmdArgs = append(cmdArgs, "-s")
	}

	// 执行Ffuf命令
	cmd := exec.CommandContext(ctx, "ffuf", cmdArgs...)
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
		result.Error = fmt.Sprintf("执行Ffuf命令失败: %v", err)
	}

	return result, nil
}

// GetToolName 获取工具名称
// 返回值列表：
//   - string 工具名称
func (e *FfufExecutor) GetToolName() string {
	return "Ffuf"
}
