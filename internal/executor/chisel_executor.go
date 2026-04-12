package executor

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"nezha_sec/internal/model"
)

type ChiselExecutor struct{}

func NewChiselExecutor() (*ChiselExecutor, error) {
	return &ChiselExecutor{}, nil
}

func (e *ChiselExecutor) Execute(ctx context.Context, arguments map[string]interface{}) (*model.ExecutionResult, error) {
	startTime := time.Now()

	mode, ok := arguments["mode"].(string)
	if !ok || mode == "" {
		return &model.ExecutionResult{
			Success:       false,
			ToolName:      e.GetToolName(),
			Error:         "缺少mode参数",
			ExecutionTime: time.Since(startTime).Milliseconds(),
		}, nil
	}

	cmdArgs := []string{mode}

	if server, ok := arguments["server"].(string); ok && server != "" {
		cmdArgs = append(cmdArgs, server)
	}

	if port, ok := arguments["port"].(string); ok && port != "" {
		cmdArgs = append(cmdArgs, port)
	}

	cmd := exec.CommandContext(ctx, "chisel", cmdArgs...)
	output, err := cmd.CombinedOutput()

	result := &model.ExecutionResult{
		Success:       err == nil,
		ToolName:      e.GetToolName(),
		Output:        string(output),
		ExecutionTime: time.Since(startTime).Milliseconds(),
		StructuredOutput: map[string]interface{}{
			"mode":      mode,
			"arguments": arguments,
			"output":    string(output),
			"error":     err != nil,
		},
	}

	if err != nil {
		result.Error = fmt.Sprintf("执行Chisel命令失败: %v", err)
	}

	return result, nil
}

func (e *ChiselExecutor) GetToolName() string {
	return "Chisel"
}
