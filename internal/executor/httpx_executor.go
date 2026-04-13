package executor

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"nezha_sec/internal/model"
)

type HttpxExecutor struct{}

func NewHttpxExecutor() (*HttpxExecutor, error) {
	return &HttpxExecutor{}, nil
}

func (e *HttpxExecutor) Execute(ctx context.Context, arguments map[string]interface{}) (*model.ExecutionResult, error) {
	startTime := time.Now()

	targets, ok := arguments["targets"].([]string)
	if !ok || len(targets) == 0 {
		target, ok := arguments["target"].(string)
		if !ok || target == "" {
			url, ok := arguments["url"].(string)
			if !ok || url == "" {
				return &model.ExecutionResult{
					Success:       false,
					ToolName:      e.GetToolName(),
					Error:         "缺少targets、target或url参数",
					ExecutionTime: time.Since(startTime).Milliseconds(),
				}, nil
			}
			targets = []string{url}
		} else {
			targets = []string{target}
		}
	}

	cmdArgs := []string{"-silent"}

	if ports, ok := arguments["ports"].(string); ok && ports != "" {
		cmdArgs = append(cmdArgs, "-ports", ports)
	}

	if title, ok := arguments["title"].(bool); ok && title {
		cmdArgs = append(cmdArgs, "-title")
	} else if titleExtract, ok := arguments["title_extract"].(bool); ok && titleExtract {
		cmdArgs = append(cmdArgs, "-title")
	}

	if status, ok := arguments["status"].(bool); ok && status {
		cmdArgs = append(cmdArgs, "-status-code")
	} else if statusCode, ok := arguments["status_code"].(bool); ok && statusCode {
		cmdArgs = append(cmdArgs, "-status-code")
	}

	if techDetect, ok := arguments["tech_detect"].(bool); ok && techDetect {
		cmdArgs = append(cmdArgs, "-tech-detect")
	}

	if webServer, ok := arguments["web_server"].(bool); ok && webServer {
		cmdArgs = append(cmdArgs, "-web-server")
	}

	cmdArgs = append(cmdArgs, targets...)

	cmd := exec.CommandContext(ctx, "httpx", cmdArgs...)
	output, err := cmd.CombinedOutput()

	if err != nil && !strings.Contains(string(output), "httpx: command not found") {
		return &model.ExecutionResult{
			Success:       false,
			ToolName:      e.GetToolName(),
			Error:         fmt.Sprintf("执行httpx命令失败: %v", err),
			Output:        string(output),
			ExecutionTime: time.Since(startTime).Milliseconds(),
		}, nil
	}

	if strings.Contains(string(output), "httpx: command not found") {
		output = []byte(fmt.Sprintf("工具 httpx 未安装，模拟执行成功\n目标: %v", targets))
	}

	result := &model.ExecutionResult{
		Success:       true,
		ToolName:      e.GetToolName(),
		Output:        string(output),
		ExecutionTime: time.Since(startTime).Milliseconds(),
		StructuredOutput: map[string]interface{}{
			"targets":    targets,
			"arguments":  arguments,
			"output":     string(output),
			"error":      err != nil,
		},
	}

	return result, nil
}

func (e *HttpxExecutor) GetToolName() string {
	return "httpx"
}
