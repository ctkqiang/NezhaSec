package executor

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"nezha_sec/internal/model"
)

// ResponderExecutor Responder工具执行器
// 用于执行Responder命令进行横向移动
type ResponderExecutor struct {}

// NewResponderExecutor 创建一个新的Responder执行器
// 返回值列表：
//   - *ResponderExecutor Responder执行器实例
//   - error 初始化过程中可能发生的错误
func NewResponderExecutor() (*ResponderExecutor, error) {
	return &ResponderExecutor{}, nil
}

// Execute 执行Responder命令
// 参数列表：
//   - ctx 上下文，用于控制执行超时和取消
//   - arguments 执行参数，包含interface等
// 返回值列表：
//   - *model.ExecutionResult 执行结果
//   - error 执行过程中可能发生的错误
func (e *ResponderExecutor) Execute(ctx context.Context, arguments map[string]interface{}) (*model.ExecutionResult, error) {
	startTime := time.Now()

	// 获取接口参数
	interfaceName, ok := arguments["interface"].(string)
	if !ok || interfaceName == "" {
		return &model.ExecutionResult{
			Success: false,
			ToolName: e.GetToolName(),
			Error: "缺少interface参数",
			ExecutionTime: time.Since(startTime).Milliseconds(),
		}, nil
	}

	// 构建Responder命令
	cmdArgs := []string{"-I", interfaceName}

	// 添加其他参数
	if verbose, ok := arguments["verbose"].(bool); ok && verbose {
		cmdArgs = append(cmdArgs, "-v")
	}

	// 执行Responder命令
	cmd := exec.CommandContext(ctx, "responder", cmdArgs...)
	output, err := cmd.CombinedOutput()

	// 构建结果
	result := &model.ExecutionResult{
		Success: err == nil,
		ToolName: e.GetToolName(),
		Output: string(output),
		ExecutionTime: time.Since(startTime).Milliseconds(),
		StructuredOutput: map[string]interface{}{
			"interface": interfaceName,
			"arguments": arguments,
			"output": string(output),
			"error": err != nil,
		},
	}

	if err != nil {
		result.Error = fmt.Sprintf("执行Responder命令失败: %v", err)
	}

	return result, nil
}

// GetToolName 获取工具名称
// 返回值列表：
//   - string 工具名称
func (e *ResponderExecutor) GetToolName() string {
	return "Responder"
}
