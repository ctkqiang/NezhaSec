package executor

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"nezha_sec/internal/model"
)

type MsfconsoleExecutor struct{}

func NewMsfconsoleExecutor() (*MsfconsoleExecutor, error) {
	return &MsfconsoleExecutor{}, nil
}

func (e *MsfconsoleExecutor) Execute(ctx context.Context, arguments map[string]interface{}) (*model.ExecutionResult, error) {
	startTime := time.Now()

	module, ok := arguments["module"].(string)
	if !ok || module == "" {
		return &model.ExecutionResult{
			Success:       false,
			ToolName:      e.GetToolName(),
			Error:         "缺少module参数",
			ExecutionTime: time.Since(startTime).Milliseconds(),
		}, nil
	}

	cmdArgs := []string{"-x", fmt.Sprintf("use %s; show options; exit", module)}

	cmd := exec.CommandContext(ctx, "msfconsole", cmdArgs...)
	output, err := cmd.CombinedOutput()

	result := &model.ExecutionResult{
		Success:       err == nil,
		ToolName:      e.GetToolName(),
		Output:        string(output),
		ExecutionTime: time.Since(startTime).Milliseconds(),
		StructuredOutput: map[string]interface{}{
			"module":    module,
			"arguments": arguments,
			"output":    string(output),
			"error":     err != nil,
		},
	}

	if err != nil {
		result.Error = fmt.Sprintf("执行Msfconsole命令失败: %v", err)
	}

	return result, nil
}

func (e *MsfconsoleExecutor) GetToolName() string {
	return "msfconsole"
}
