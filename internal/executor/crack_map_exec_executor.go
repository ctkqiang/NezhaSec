package executor

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"nezha_sec/internal/model"
)

type CrackMapExecExecutor struct{}

func NewCrackMapExecExecutor() (*CrackMapExecExecutor, error) {
	return &CrackMapExecExecutor{}, nil
}

func (e *CrackMapExecExecutor) Execute(ctx context.Context, arguments map[string]interface{}) (*model.ExecutionResult, error) {
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

	cmdArgs := []string{"smb", target}

	if username, ok := arguments["username"].(string); ok && username != "" {
		cmdArgs = append(cmdArgs, "-u", username)
	}

	if password, ok := arguments["password"].(string); ok && password != "" {
		cmdArgs = append(cmdArgs, "-p", password)
	}

	cmd := exec.CommandContext(ctx, "crackmapexec", cmdArgs...)
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
		result.Error = fmt.Sprintf("执行CrackMapExec命令失败: %v", err)
	}

	return result, nil
}

func (e *CrackMapExecExecutor) GetToolName() string {
	return "crackmapexec"
}
