package executor

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"nezha_sec/internal/model"
)

// CrackMapExecExecutor CrackMapExec工具执行器
// 用于执行CrackMapExec命令进行横向移动
type CrackMapExecExecutor struct {}

// NewCrackMapExecExecutor 创建一个新的CrackMapExec执行器
// 返回值列表：
//   - *CrackMapExecExecutor CrackMapExec执行器实例
//   - error 初始化过程中可能发生的错误
func NewCrackMapExecExecutor() (*CrackMapExecExecutor, error) {
	return &CrackMapExecExecutor{}, nil
}

// Execute 执行CrackMapExec命令
// 参数列表：
//   - ctx 上下文，用于控制执行超时和取消
//   - arguments 执行参数，包含target等
// 返回值列表：
//   - *model.ExecutionResult 执行结果
//   - error 执行过程中可能发生的错误
func (e *CrackMapExecExecutor) Execute(ctx context.Context, arguments map[string]interface{}) (*model.ExecutionResult, error) {
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

	// 构建CrackMapExec命令
	cmdArgs := []string{"smb", target}

	// 添加其他参数
	if username, ok := arguments["username"].(string); ok && username != "" {
		cmdArgs = append(cmdArgs, "-u", username)
	}

	if password, ok := arguments["password"].(string); ok && password != "" {
		cmdArgs = append(cmdArgs, "-p", password)
	}

	// 执行CrackMapExec命令
	cmd := exec.CommandContext(ctx, "crackmapexec", cmdArgs...)
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
		result.Error = fmt.Sprintf("执行CrackMapExec命令失败: %v", err)
	}

	return result, nil
}

// GetToolName 获取工具名称
// 返回值列表：
//   - string 工具名称
func (e *CrackMapExecExecutor) GetToolName() string {
	return "CrackMapExec"
}
