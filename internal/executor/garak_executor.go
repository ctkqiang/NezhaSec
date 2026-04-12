package executor

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"nezha_sec/internal/model"
)

type GarakExecutor struct{}

func NewGarakExecutor() (*GarakExecutor, error) {
	return &GarakExecutor{}, nil
}

func (e *GarakExecutor) Execute(ctx context.Context, arguments map[string]interface{}) (*model.ExecutionResult, error) {
	startTime := time.Now()

	modelName, ok := arguments["model"].(string)
	if !ok || modelName == "" {
		return &model.ExecutionResult{
			Success:       false,
			ToolName:      e.GetToolName(),
			Error:         "缺少model参数",
			ExecutionTime: time.Since(startTime).Milliseconds(),
		}, nil
	}

	cmdArgs := []string{"--model", modelName}

	if probes, ok := arguments["probes"].(string); ok && probes != "" {
		cmdArgs = append(cmdArgs, "--probes", probes)
	}

	cmd := exec.CommandContext(ctx, "garak", cmdArgs...)
	output, err := cmd.CombinedOutput()

	result := &model.ExecutionResult{
		Success:       err == nil,
		ToolName:      e.GetToolName(),
		Output:        string(output),
		ExecutionTime: time.Since(startTime).Milliseconds(),
		StructuredOutput: map[string]interface{}{
			"model":     modelName,
			"arguments": arguments,
			"output":    string(output),
			"error":     err != nil,
		},
	}

	if err != nil {
		result.Error = fmt.Sprintf("执行Garak命令失败: %v", err)
	}

	return result, nil
}

func (e *GarakExecutor) GetToolName() string {
	return "Garak"
}
