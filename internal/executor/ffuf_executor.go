package executor

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"nezha_sec/internal/model"
)

type FfufExecutor struct{}

func NewFfufExecutor() (*FfufExecutor, error) {
	return &FfufExecutor{}, nil
}

func (e *FfufExecutor) Execute(ctx context.Context, arguments map[string]interface{}) (*model.ExecutionResult, error) {
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

	if wordlist, ok := arguments["wordlist"].(string); ok && wordlist != "" {
		cmdArgs = append(cmdArgs, "-w", wordlist)
	} else {

		cmdArgs = append(cmdArgs, "-w", "/usr/share/wordlists/dirb/common.txt")
	}

	if silent, ok := arguments["silent"].(bool); ok && silent {
		cmdArgs = append(cmdArgs, "-s")
	}

	cmd := exec.CommandContext(ctx, "ffuf", cmdArgs...)
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
		result.Error = fmt.Sprintf("执行Ffuf命令失败: %v", err)
	}

	return result, nil
}

func (e *FfufExecutor) GetToolName() string {
	return "Ffuf"
}
