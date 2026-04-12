package executor

import (
	"context"
	"fmt"
	"nezha_sec/internal/model"
	"os/exec"
	"time"
)

// NmapExecutor nmap工具执行器
// 用于执行nmap命令进行端口扫描
type NmapExecutor struct{}

// NewNmapExecutor 创建一个新的nmap执行器
// 返回值列表：
//   - *NmapExecutor nmap执行器实例
//   - error 初始化过程中可能发生的错误
func NewNmapExecutor() (*NmapExecutor, error) {
	return &NmapExecutor{}, nil
}

// Execute 执行nmap命令
// 参数列表：
//   - ctx 上下文，用于控制执行超时和取消
//   - arguments 执行参数，包含target、ports等
//
// 返回值列表：
//   - *model.ExecutionResult 执行结果
//   - error 执行过程中可能发生的错误
func (e *NmapExecutor) Execute(ctx context.Context, arguments map[string]interface{}) (*model.ExecutionResult, error) {
	startTime := time.Now()

	// 获取目标参数
	target, ok := arguments["target"].(string)
	if !ok || target == "" {
		// 尝试从url参数获取目标
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

	// 构建nmap命令
	cmdArgs := []string{target}

	// 添加端口参数
	if ports, ok := arguments["ports"].(string); ok && ports != "" {
		cmdArgs = append(cmdArgs, "-p", ports)
	}

	// 添加扫描类型参数
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
		// 默认使用快速扫描
		cmdArgs = append(cmdArgs, "-F")
	}

	// 执行nmap命令
	cmd := exec.CommandContext(ctx, "nmap", cmdArgs...)
	output, err := cmd.CombinedOutput()

	// 构建结果
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

// GetToolName 获取工具名称
// 返回值列表：
//   - string 工具名称
func (e *NmapExecutor) GetToolName() string {
	return "nmap"
}
