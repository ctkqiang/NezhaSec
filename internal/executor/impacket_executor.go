package executor

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"nezha_sec/internal/model"
)

// ImpacketExecutor Impacket工具执行器
// 用于执行Impacket命令进行横向移动
type ImpacketExecutor struct {}

// NewImpacketExecutor 创建一个新的Impacket执行器
// 返回值列表：
//   - *ImpacketExecutor Impacket执行器实例
//   - error 初始化过程中可能发生的错误
func NewImpacketExecutor() (*ImpacketExecutor, error) {
	return &ImpacketExecutor{}, nil
}

// Execute 执行Impacket命令
// 参数列表：
//   - ctx 上下文，用于控制执行超时和取消
//   - arguments 执行参数，包含command、target等
// 返回值列表：
//   - *model.ExecutionResult 执行结果
//   - error 执行过程中可能发生的错误
func (e *ImpacketExecutor) Execute(ctx context.Context, arguments map[string]interface{}) (*model.ExecutionResult, error) {
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

	// 构建Impacket命令
	cmdArgs := []string{command, target}

	// 添加其他参数
	if username, ok := arguments["username"].(string); ok && username != "" {
		cmdArgs = append(cmdArgs, "-u", username)
	}

	if password, ok := arguments["password"].(string); ok && password != "" {
		cmdArgs = append(cmdArgs, "-p", password)
	}

	// 执行Impacket命令
	cmd := exec.CommandContext(ctx, command, cmdArgs...)
	output, err := cmd.CombinedOutput()

	// 构建结果
	result := &model.ExecutionResult{
		Success: err == nil,
		ToolName: e.GetToolName(),
		Output: string(output),
		ExecutionTime: time.Since(startTime).Milliseconds(),
		StructuredOutput: map[string]interface{}{
			"command": command,
			"target": target,
			"arguments": arguments,
			"output": string(output),
			"error": err != nil,
		},
	}

	if err != nil {
		result.Error = fmt.Sprintf("执行Impacket命令失败: %v", err)
	}

	return result, nil
}

// GetToolName 获取工具名称
// 返回值列表：
//   - string 工具名称
func (e *ImpacketExecutor) GetToolName() string {
	return "Impacket"
}
