package executor

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"nezha_sec/internal/model"
)

type CurlExecutor struct{}

func NewCurlExecutor() (*CurlExecutor, error) {
	return &CurlExecutor{}, nil
}

func (e *CurlExecutor) Execute(ctx context.Context, arguments map[string]interface{}) (*model.ExecutionResult, error) {
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

	method := "GET"
	if methodVal, ok := arguments["method"].(string); ok && methodVal != "" {
		method = methodVal
	}

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return &model.ExecutionResult{
			Success:       false,
			ToolName:      e.GetToolName(),
			Error:         fmt.Sprintf("创建请求失败: %v", err),
			ExecutionTime: time.Since(startTime).Milliseconds(),
		}, nil
	}

	if headers, ok := arguments["headers"].(map[string]string); ok {
		for key, value := range headers {
			req.Header.Set(key, value)
		}
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return &model.ExecutionResult{
			Success:       false,
			ToolName:      e.GetToolName(),
			Error:         fmt.Sprintf("发送请求失败: %v", err),
			ExecutionTime: time.Since(startTime).Milliseconds(),
		}, nil
	}
	defer resp.Body.Close()

	body := make([]byte, 1024*1024)
	n, _ := resp.Body.Read(body)
	responseBody := string(body[:n])

	result := &model.ExecutionResult{
		Success:       true,
		ToolName:      e.GetToolName(),
		Output:        fmt.Sprintf("HTTP %d\n%s", resp.StatusCode, responseBody),
		ExecutionTime: time.Since(startTime).Milliseconds(),
		StructuredOutput: map[string]interface{}{
			"status_code": resp.StatusCode,
			"headers":     resp.Header,
			"body":        responseBody,
		},
	}

	return result, nil
}

func (e *CurlExecutor) GetToolName() string {
	return "curl"
}
