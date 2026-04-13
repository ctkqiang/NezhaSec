package executor

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"nezha_sec/internal/model"
)

type NucleiExecutor struct{}

func NewNucleiExecutor() (*NucleiExecutor, error) {
	return &NucleiExecutor{}, nil
}

func (e *NucleiExecutor) Execute(ctx context.Context, arguments map[string]interface{}) (*model.ExecutionResult, error) {
	startTime := time.Now()

	target, ok := arguments["target"].(string)
	if !ok || target == "" {
		return &model.ExecutionResult{
			Success:       false,
			ToolName:      e.GetToolName(),
			Error:         "缺少target参数",
			ExecutionTime: time.Since(startTime).Milliseconds(),
		}, nil
	}

	cmdArgs := []string{"-u", target}

	if silent, ok := arguments["silent"].(bool); ok && silent {
		cmdArgs = append(cmdArgs, "-silent")
	}

	if templates, ok := arguments["templates"].(string); ok && templates != "" {
		cmdArgs = append(cmdArgs, "-t", templates)
	}

	cmd := exec.CommandContext(ctx, "nuclei", cmdArgs...)
	output, err := cmd.CombinedOutput()

	result := &model.ExecutionResult{
		Success:       err == nil,
		ToolName:      e.GetToolName(),
		Output:        string(output),
		ExecutionTime: time.Since(startTime).Milliseconds(),
		StructuredOutput: map[string]interface{}{
			"target":    target,
			"arguments": arguments,
			"output":    string(output),
			"error":     err != nil,
		},
	}

	if err != nil {
		result.Error = fmt.Sprintf("执行Nuclei命令失败: %v", err)
	}

	return result, nil
}

func (e *NucleiExecutor) GetToolName() string {
	return "nuclei"
}
