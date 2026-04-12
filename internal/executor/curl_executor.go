package executor

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"nezha_sec/internal/model"
)

// CurlExecutor curl工具执行器
// 用于发送HTTP请求并获取响应
type CurlExecutor struct {}

// NewCurlExecutor 创建一个新的curl执行器
// 返回值列表：
//   - *CurlExecutor curl执行器实例
//   - error 初始化过程中可能发生的错误
func NewCurlExecutor() (*CurlExecutor, error) {
	return &CurlExecutor{}, nil
}

// Execute 执行curl命令
// 参数列表：
//   - ctx 上下文，用于控制执行超时和取消
//   - arguments 执行参数，包含url、method等
// 返回值列表：
//   - *model.ExecutionResult 执行结果
//   - error 执行过程中可能发生的错误
func (e *CurlExecutor) Execute(ctx context.Context, arguments map[string]interface{}) (*model.ExecutionResult, error) {
	startTime := time.Now()

	// 获取URL参数
	url, ok := arguments["url"].(string)
	if !ok || url == "" {
		return &model.ExecutionResult{
			Success: false,
			ToolName: e.GetToolName(),
			Error: "缺少url参数",
			ExecutionTime: time.Since(startTime).Milliseconds(),
		}, nil
	}

	// 获取HTTP方法
	method := "GET"
	if methodVal, ok := arguments["method"].(string); ok && methodVal != "" {
		method = methodVal
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return &model.ExecutionResult{
			Success: false,
			ToolName: e.GetToolName(),
			Error: fmt.Sprintf("创建请求失败: %v", err),
			ExecutionTime: time.Since(startTime).Milliseconds(),
		}, nil
	}

	// 设置请求头
	if headers, ok := arguments["headers"].(map[string]string); ok {
		for key, value := range headers {
			req.Header.Set(key, value)
		}
	}

	// 发送请求 - 使用上下文的超时设置
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return &model.ExecutionResult{
			Success: false,
			ToolName: e.GetToolName(),
			Error: fmt.Sprintf("发送请求失败: %v", err),
			ExecutionTime: time.Since(startTime).Milliseconds(),
		}, nil
	}
	defer resp.Body.Close()

	// 读取响应
	body := make([]byte, 1024*1024) // 1MB缓冲区
	n, _ := resp.Body.Read(body)
	responseBody := string(body[:n])

	// 构建结果
	result := &model.ExecutionResult{
		Success: true,
		ToolName: e.GetToolName(),
		Output: fmt.Sprintf("HTTP %d\n%s", resp.StatusCode, responseBody),
		ExecutionTime: time.Since(startTime).Milliseconds(),
		StructuredOutput: map[string]interface{}{
			"status_code": resp.StatusCode,
			"headers": resp.Header,
			"body": responseBody,
		},
	}

	return result, nil
}

// GetToolName 获取工具名称
// 返回值列表：
//   - string 工具名称
func (e *CurlExecutor) GetToolName() string {
	return "curl"
}
