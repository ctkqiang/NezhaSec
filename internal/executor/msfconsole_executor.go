package executor

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"nezha_sec/internal/model"
)

// MsfconsoleExecutor Msfconsole工具执行器
// 用于执行Msfconsole命令进行漏洞利用
type MsfconsoleExecutor struct {}

// NewMsfconsoleExecutor 创建一个新的Msfconsole执行器
// 返回值列表：
//   - *MsfconsoleExecutor Msfconsole执行器实例
//   - error 初始化过程中可能发生的错误
func NewMsfconsoleExecutor() (*MsfconsoleExecutor, error) {
	return &MsfconsoleExecutor{}, nil
}

// Execute 执行Msfconsole命令
// 参数列表：
//   - ctx 上下文，用于控制执行超时和取消
//   - arguments 执行参数，包含module等
// 返回值列表：
//   - *model.ExecutionResult 执行结果
//   - error 执行过程中可能发生的错误
func (e *MsfconsoleExecutor) Execute(ctx context.Context, arguments map[string]interface{}) (*model.ExecutionResult, error) {
	startTime := time.Now()

	// 获取模块参数
	module, ok := arguments["module"].(string)
	if !ok || module == "" {
		return &model.ExecutionResult{
			Success: false,
			ToolName: e.GetToolName(),
			Error: "缺少module参数",
			ExecutionTime: time.Since(startTime).Milliseconds(),
		}, nil
	}

	// 构建Msfconsole命令
	cmdArgs := []string{"-x", fmt.Sprintf("use %s; show options; exit", module)}

	// 执行Msfconsole命令
	cmd := exec.CommandContext(ctx, "msfconsole", cmdArgs...)
	output, err := cmd.CombinedOutput()

	// 构建结果
	result := &model.ExecutionResult{
		Success: err == nil,
		ToolName: e.GetToolName(),
		Output: string(output),
		ExecutionTime: time.Since(startTime).Milliseconds(),
		StructuredOutput: map[string]interface{}{
			"module": module,
			"arguments": arguments,
			"output": string(output),
			"error": err != nil,
		},
	}

	if err != nil {
		result.Error = fmt.Sprintf("执行Msfconsole命令失败: %v", err)
	}

	return result, nil
}

// GetToolName 获取工具名称
// 返回值列表：
//   - string 工具名称
func (e *MsfconsoleExecutor) GetToolName() string {
	return "msfconsole"
}
