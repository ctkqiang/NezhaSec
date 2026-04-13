package executor

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"nezha_sec/internal/model"
)

type CommixExecutor struct{}

func NewCommixExecutor() (*CommixExecutor, error) {
	return &CommixExecutor{}, nil
}

func (e *CommixExecutor) Execute(ctx context.Context, arguments map[string]interface{}) (*model.ExecutionResult, error) {
	startTime := time.Now()

	url, ok := arguments["url"].(string)
	if !ok || url == "" {
		return &model.ExecutionResult{
			Success:       false,
			ToolName:      e.GetToolName(),
			Error:         "缺少url参数",
			ExecutionTime: time.Since(startTime).Milliseconds(),
		}, nil
	}

	cmdArgs := []string{"--url", url}

	if batch, ok := arguments["batch"].(bool); ok && batch {
		cmdArgs = append(cmdArgs, "--batch")
	} else {

		cmdArgs = append(cmdArgs, "--batch")
	}

	cmd := exec.CommandContext(ctx, "commix", cmdArgs...)
	output, err := cmd.CombinedOutput()

	result := &model.ExecutionResult{
		Success:       err == nil,
		ToolName:      e.GetToolName(),
		Output:        string(output),
		ExecutionTime: time.Since(startTime).Milliseconds(),
		StructuredOutput: map[string]interface{}{
			"url":       url,
			"arguments": arguments,
			"output":    string(output),
			"error":     err != nil,
		},
	}

	if err != nil {
		result.Error = fmt.Sprintf("执行Commix命令失败: %v", err)
	}

	return result, nil
}

func (e *CommixExecutor) GetToolName() string {
	return "commix"
}
