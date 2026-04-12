package executor

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"nezha_sec/internal/model"
)

type SqlmapExecutor struct{}

func NewSqlmapExecutor() (*SqlmapExecutor, error) {
	return &SqlmapExecutor{}, nil
}

func (e *SqlmapExecutor) Execute(ctx context.Context, arguments map[string]interface{}) (*model.ExecutionResult, error) {
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

	cmdArgs := []string{"-u", url}

	if risk, ok := arguments["risk"].(string); ok && risk != "" {
		cmdArgs = append(cmdArgs, "--risk", risk)
	} else {

		cmdArgs = append(cmdArgs, "--risk", "1")
	}

	if level, ok := arguments["level"].(string); ok && level != "" {
		cmdArgs = append(cmdArgs, "--level", level)
	} else {

		cmdArgs = append(cmdArgs, "--level", "1")
	}

	if batch, ok := arguments["batch"].(bool); ok && batch {
		cmdArgs = append(cmdArgs, "--batch")
	} else {

		cmdArgs = append(cmdArgs, "--batch")
	}

	if threads, ok := arguments["threads"].(string); ok && threads != "" {
		cmdArgs = append(cmdArgs, "--threads", threads)
	}

	cmd := exec.CommandContext(ctx, "sqlmap", cmdArgs...)
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
		result.Error = fmt.Sprintf("执行sqlmap命令失败: %v", err)
	}

	return result, nil
}

func (e *SqlmapExecutor) GetToolName() string {
	return "sqlmap"
}
