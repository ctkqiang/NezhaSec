package executor

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"nezha_sec/internal/model"
)

type AmassExecutor struct{}

func NewAmassExecutor() (*AmassExecutor, error) {
	return &AmassExecutor{}, nil
}

func (e *AmassExecutor) Execute(ctx context.Context, arguments map[string]interface{}) (*model.ExecutionResult, error) {
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

	cmdArgs := []string{"enum", "-d", domain}

	if passive, ok := arguments["passive"].(bool); ok && passive {
		cmdArgs = append(cmdArgs, "-passive")
	}

	if active, ok := arguments["active"].(bool); ok && active {
		cmdArgs = append(cmdArgs, "-active")
	}

	cmd := exec.CommandContext(ctx, "amass", cmdArgs...)
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
		result.Error = fmt.Sprintf("执行Amass命令失败: %v", err)
	}

	return result, nil
}

func (e *AmassExecutor) GetToolName() string {
	return "amass"
}
