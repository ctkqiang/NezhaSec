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

	cmdArgs := []string{"--batch", "-u", url}

	if riskVal, ok := arguments["risk"]; ok {
		var riskStr string
		switch v := riskVal.(type) {
		case string:
			riskStr = v
		case int:
			riskStr = fmt.Sprintf("%d", v)
		case float64:
			riskStr = fmt.Sprintf("%.0f", v)
		}
		if riskStr != "" {
			cmdArgs = append(cmdArgs, "--risk", riskStr)
		}
	} else {
		cmdArgs = append(cmdArgs, "--risk", "1")
	}

	if levelVal, ok := arguments["level"]; ok {
		var levelStr string
		switch v := levelVal.(type) {
		case string:
			levelStr = v
		case int:
			levelStr = fmt.Sprintf("%d", v)
		case float64:
			levelStr = fmt.Sprintf("%.0f", v)
		}
		if levelStr != "" {
			cmdArgs = append(cmdArgs, "--level", levelStr)
		}
	} else {
		cmdArgs = append(cmdArgs, "--level", "1")
	}

	if threadsVal, ok := arguments["threads"]; ok {
		var threadsStr string
		switch v := threadsVal.(type) {
		case string:
			threadsStr = v
		case int:
			threadsStr = fmt.Sprintf("%d", v)
		case float64:
			threadsStr = fmt.Sprintf("%.0f", v)
		}
		if threadsStr != "" {
			cmdArgs = append(cmdArgs, "--threads", threadsStr)
		}
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
