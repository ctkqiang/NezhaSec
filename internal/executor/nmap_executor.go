package executor

import (
	"context"
	"fmt"
	"nezha_sec/internal/model"
	"os/exec"
	"time"
)

type NmapExecutor struct{}

func NewNmapExecutor() (*NmapExecutor, error) {
	return &NmapExecutor{}, nil
}

func (e *NmapExecutor) Execute(ctx context.Context, arguments map[string]interface{}) (*model.ExecutionResult, error) {
	startTime := time.Now()

	target, ok := arguments["target"].(string)
	if !ok || target == "" {

		if urlVal, ok := arguments["url"].(string); ok && urlVal != "" {
			target = urlVal
		} else {
			return &model.ExecutionResult{
				Success:       false,
				ToolName:      e.GetToolName(),
				Error:         "缺少target或url参数",
				ExecutionTime: time.Since(startTime).Milliseconds(),
			}, nil
		}
	}

	cmdArgs := []string{target}

	if ports, ok := arguments["ports"].(string); ok && ports != "" {
		cmdArgs = append(cmdArgs, "-p", ports)
	}

	if scanType, ok := arguments["scan_type"].(string); ok && scanType != "" {
		switch scanType {
		case "fast":
			cmdArgs = append(cmdArgs, "-F")
		case "syn":
			cmdArgs = append(cmdArgs, "-sS")
		case "udp":
			cmdArgs = append(cmdArgs, "-sU")
		case "comprehensive":
			cmdArgs = append(cmdArgs, "-sV", "-sC")
		}
	} else {

		cmdArgs = append(cmdArgs, "-F")
	}

	cmd := exec.CommandContext(ctx, "nmap", cmdArgs...)
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
		result.Error = fmt.Sprintf("执行nmap命令失败: %v", err)
	}

	return result, nil
}

func (e *NmapExecutor) GetToolName() string {
	return "nmap"
}
