package executor

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"nezha_sec/internal/model"
)

// GarakExecutor Garak工具执行器
// 用于执行Garak命令进行AI/LLM扫描
type GarakExecutor struct {}

// NewGarakExecutor 创建一个新的Garak执行器
// 返回值列表：
//   - *GarakExecutor Garak执行器实例
//   - error 初始化过程中可能发生的错误
func NewGarakExecutor() (*GarakExecutor, error) {
	return &GarakExecutor{}, nil
}

// Execute 执行Garak命令
// 参数列表：
//   - ctx 上下文，用于控制执行超时和取消
//   - arguments 执行参数，包含model等
// 返回值列表：
//   - *model.ExecutionResult 执行结果
//   - error 执行过程中可能发生的错误
func (e *GarakExecutor) Execute(ctx context.Context, arguments map[string]interface{}) (*model.ExecutionResult, error) {
	startTime := time.Now()

	// 获取模型参数
	modelName, ok := arguments["model"].(string)
	if !ok || modelName == "" {
		return &model.ExecutionResult{
			Success: false,
			ToolName: e.GetToolName(),
			Error: "缺少model参数",
			ExecutionTime: time.Since(startTime).Milliseconds(),
		}, nil
	}

	// 构建Garak命令
	cmdArgs := []string{"--model", modelName}

	// 添加其他参数
	if probes, ok := arguments["probes"].(string); ok && probes != "" {
		cmdArgs = append(cmdArgs, "--probes", probes)
	}

	// 执行Garak命令
	cmd := exec.CommandContext(ctx, "garak", cmdArgs...)
	output, err := cmd.CombinedOutput()

	// 构建结果
	result := &model.ExecutionResult{
		Success: err == nil,
		ToolName: e.GetToolName(),
		Output: string(output),
		ExecutionTime: time.Since(startTime).Milliseconds(),
		StructuredOutput: map[string]interface{}{
			"model": modelName,
			"arguments": arguments,
			"output": string(output),
			"error": err != nil,
		},
	}

	if err != nil {
		result.Error = fmt.Sprintf("执行Garak命令失败: %v", err)
	}

	return result, nil
}

// GetToolName 获取工具名称
// 返回值列表：
//   - string 工具名称
func (e *GarakExecutor) GetToolName() string {
	return "Garak"
}
