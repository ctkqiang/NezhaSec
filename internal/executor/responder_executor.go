package executor

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"nezha_sec/internal/model"
)

type ResponderExecutor struct{}

func NewResponderExecutor() (*ResponderExecutor, error) {
	return &ResponderExecutor{}, nil
}

func (e *ResponderExecutor) Execute(ctx context.Context, arguments map[string]interface{}) (*model.ExecutionResult, error) {
	startTime := time.Now()

	interfaceName, ok := arguments["interface"].(string)
	if !ok || interfaceName == "" {
		return &model.ExecutionResult{
			Success:       false,
			ToolName:      e.GetToolName(),
			Error:         "缺少interface参数",
			ExecutionTime: time.Since(startTime).Milliseconds(),
		}, nil
	}

	cmdArgs := []string{"-I", interfaceName}

	if verbose, ok := arguments["verbose"].(bool); ok && verbose {
		cmdArgs = append(cmdArgs, "-v")
	}

	cmd := exec.CommandContext(ctx, "responder", cmdArgs...)
	output, err := cmd.CombinedOutput()

	result := &model.ExecutionResult{
		Success:       err == nil,
		ToolName:      e.GetToolName(),
		Output:        string(output),
		ExecutionTime: time.Since(startTime).Milliseconds(),
		StructuredOutput: map[string]interface{}{
			"interface": interfaceName,
			"arguments": arguments,
			"output":    string(output),
			"error":     err != nil,
		},
	}

	if err != nil {
		result.Error = fmt.Sprintf("执行Responder命令失败: %v", err)
	}

	return result, nil
}

func (e *ResponderExecutor) GetToolName() string {
	return "Responder"
}
