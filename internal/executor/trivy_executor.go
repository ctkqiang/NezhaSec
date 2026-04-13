package executor

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"nezha_sec/internal/model"
)

type TrivyExecutor struct{}

func NewTrivyExecutor() (*TrivyExecutor, error) {
	return &TrivyExecutor{}, nil
}

func (e *TrivyExecutor) Execute(ctx context.Context, arguments map[string]interface{}) (*model.ExecutionResult, error) {
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

	cmdArgs := []string{"image", target}

	if severity, ok := arguments["severity"].(string); ok && severity != "" {
		cmdArgs = append(cmdArgs, "--severity", severity)
	}

	cmd := exec.CommandContext(ctx, "trivy", cmdArgs...)
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
		result.Error = fmt.Sprintf("执行Trivy命令失败: %v", err)
	}

	return result, nil
}

func (e *TrivyExecutor) GetToolName() string {
	return "trivy"
}
