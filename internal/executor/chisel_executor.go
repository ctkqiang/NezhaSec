package executor

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"nezha_sec/internal/model"
)

// ChiselExecutor Chisel工具执行器
// 用于执行Chisel命令进行端口转发和代理
type ChiselExecutor struct {}

// NewChiselExecutor 创建一个新的Chisel执行器
// 返回值列表：
//   - *ChiselExecutor Chisel执行器实例
//   - error 初始化过程中可能发生的错误
func NewChiselExecutor() (*ChiselExecutor, error) {
	return &ChiselExecutor{}, nil
}

// Execute 执行Chisel命令
// 参数列表：
//   - ctx 上下文，用于控制执行超时和取消
//   - arguments 执行参数，包含mode、server等
// 返回值列表：
//   - *model.ExecutionResult 执行结果
//   - error 执行过程中可能发生的错误
func (e *ChiselExecutor) Execute(ctx context.Context, arguments map[string]interface{}) (*model.ExecutionResult, error) {
	startTime := time.Now()

	// 获取模式参数
	mode, ok := arguments["mode"].(string)
	if !ok || mode == "" {
		return &model.ExecutionResult{
			Success: false,
			ToolName: e.GetToolName(),
			Error: "缺少mode参数",
			ExecutionTime: time.Since(startTime).Milliseconds(),
		}, nil
	}

	// 构建Chisel命令
	cmdArgs := []string{mode}

	// 添加服务器参数
	if server, ok := arguments["server"].(string); ok && server != "" {
		cmdArgs = append(cmdArgs, server)
	}

	// 添加端口转发参数
	if port, ok := arguments["port"].(string); ok && port != "" {
		cmdArgs = append(cmdArgs, port)
	}

	// 执行Chisel命令
	cmd := exec.CommandContext(ctx, "chisel", cmdArgs...)
	output, err := cmd.CombinedOutput()

	// 构建结果
	result := &model.ExecutionResult{
		Success: err == nil,
		ToolName: e.GetToolName(),
		Output: string(output),
		ExecutionTime: time.Since(startTime).Milliseconds(),
		StructuredOutput: map[string]interface{}{
			"mode": mode,
			"arguments": arguments,
			"output": string(output),
			"error": err != nil,
		},
	}

	if err != nil {
		result.Error = fmt.Sprintf("执行Chisel命令失败: %v", err)
	}

	return result, nil
}

// GetToolName 获取工具名称
// 返回值列表：
//   - string 工具名称
func (e *ChiselExecutor) GetToolName() string {
	return "Chisel"
}
