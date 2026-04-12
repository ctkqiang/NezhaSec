package executor

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"nezha_sec/internal/model"
)

type SubfinderExecutor struct{}

func NewSubfinderExecutor() (*SubfinderExecutor, error) {
	return &SubfinderExecutor{}, nil
}

func (e *SubfinderExecutor) Execute(ctx context.Context, arguments map[string]interface{}) (*model.ExecutionResult, error) {
	startTime := time.Now()

	domain, ok := arguments["domain"].(string)
	if !ok || domain == "" {
		return &model.ExecutionResult{
			Success:       false,
			ToolName:      e.GetToolName(),
			Error:         "缺少domain参数",
			ExecutionTime: time.Since(startTime).Milliseconds(),
		}, nil
	}

	cmdArgs := []string{"-d", domain}

	if silent, ok := arguments["silent"].(bool); ok && silent {
		cmdArgs = append(cmdArgs, "-silent")
	}

	if all, ok := arguments["all"].(bool); ok && all {
		cmdArgs = append(cmdArgs, "-all")
	}

	cmd := exec.CommandContext(ctx, "subfinder", cmdArgs...)
	output, err := cmd.CombinedOutput()

	result := &model.ExecutionResult{
		Success:       err == nil,
		ToolName:      e.GetToolName(),
		Output:        string(output),
		ExecutionTime: time.Since(startTime).Milliseconds(),
		StructuredOutput: map[string]interface{}{
			"domain":    domain,
			"arguments": arguments,
			"output":    string(output),
			"error":     err != nil,
		},
	}

	if err != nil {
		result.Error = fmt.Sprintf("执行Subfinder命令失败: %v", err)
	}

	return result, nil
}

func (e *SubfinderExecutor) GetToolName() string {
	return "Subfinder"
}
