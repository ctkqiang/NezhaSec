package executor

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"nezha_sec/internal/model"
)

type ImpacketExecutor struct{}

func NewImpacketExecutor() (*ImpacketExecutor, error) {
	return &ImpacketExecutor{}, nil
}

func (e *ImpacketExecutor) Execute(ctx context.Context, arguments map[string]interface{}) (*model.ExecutionResult, error) {
	startTime := time.Now()

	command, ok := arguments["command"].(string)
	if !ok || command == "" {
		return &model.ExecutionResult{
			Success:       false,
			ToolName:      e.GetToolName(),
			Error:         "缺少command参数",
			ExecutionTime: time.Since(startTime).Milliseconds(),
		}, nil
	}

	target, ok := arguments["target"].(string)
	if !ok || target == "" {
		return &model.ExecutionResult{
			Success:       false,
			ToolName:      e.GetToolName(),
			Error:         "缺少target参数",
			ExecutionTime: time.Since(startTime).Milliseconds(),
		}, nil
	}

	cmdArgs := []string{command, target}

	if username, ok := arguments["username"].(string); ok && username != "" {
		cmdArgs = append(cmdArgs, "-u", username)
	}

	if password, ok := arguments["password"].(string); ok && password != "" {
		cmdArgs = append(cmdArgs, "-p", password)
	}

	cmd := exec.CommandContext(ctx, command, cmdArgs...)
	output, err := cmd.CombinedOutput()

	result := &model.ExecutionResult{
		Success:       err == nil,
		ToolName:      e.GetToolName(),
		Output:        string(output),
		ExecutionTime: time.Since(startTime).Milliseconds(),
		StructuredOutput: map[string]interface{}{
			"command":   command,
			"target":    target,
			"arguments": arguments,
			"output":    string(output),
			"error":     err != nil,
		},
	}

	if err != nil {
		result.Error = fmt.Sprintf("执行Impacket命令失败: %v", err)
	}

	return result, nil
}

func (e *ImpacketExecutor) GetToolName() string {
	return "Impacket"
}
