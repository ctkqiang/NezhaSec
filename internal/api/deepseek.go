package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type (
	ProgressMsg         string
	DeepSeekResponseMsg struct {
		Result string
	}
	ToolCallMsg struct {
		ToolName  string                 `json:"tool_name"`
		Arguments map[string]interface{} `json:"arguments"`
	}
)

type DeepSeekRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type DeepSeekResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func GetDeepSeekAPIKey() string {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {

		dotenv, err := os.ReadFile(".env")
		if err == nil {
			lines := bytes.Split(dotenv, []byte("\n"))
			for _, line := range lines {
				if bytes.HasPrefix(line, []byte("DEEPSEEK_API_KEY=")) {
					apiKey = string(bytes.TrimPrefix(line, []byte("DEEPSEEK_API_KEY=")))
					break
				}
			}
		}
	}
	return apiKey
}

func CallDeepSeekAPI(url string) tea.Cmd {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	return CallDeepSeekAPIWithContext(ctx, url)
}

func CallDeepSeekAPIWithContext(ctx context.Context, url string) tea.Cmd {
	return func() tea.Msg {
		apiKey := GetDeepSeekAPIKey()
		if apiKey == "" {
			return ProgressMsg("错误: 未配置DeepSeek API密钥")
		}

		resultChan := make(chan tea.Msg, 1)
		go func() {
			result, err := AnalyzeURLWithDeepSeek(url, apiKey)
			if err != nil {
				select {
				case resultChan <- ProgressMsg(fmt.Sprintf("API调用失败: %v", err)):
				case <-ctx.Done():
					return
				}
			} else {
				select {
				case resultChan <- DeepSeekResponseMsg{Result: result}:
				case <-ctx.Done():
					return
				}
			}
		}()

		select {
		case msg := <-resultChan:
			return msg
		case <-ctx.Done():
			return ProgressMsg("API调用已取消")
		}
	}
}

func progressWrapper(fn func() tea.Msg) tea.Msg {
	resultMsg := fn()

	if _, ok := resultMsg.(ProgressMsg); ok {
		return resultMsg
	}

	return resultMsg
}

func AnalyzeURLWithDeepSeek(url string, apiKey string) (string, error) {
	const maxRetries = 3
	const retryDelay = 2 * time.Second

	fmt.Fprintf(os.Stderr, "开始分析URL: %s\n", url)

	rulesContent, err := os.ReadFile(".deepseekrules")
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取.deepseekrules文件失败: %v\n", err)
		return "", fmt.Errorf("读取规则文件失败: %w", err)
	}

	systemPrompt := string(rulesContent)

	for attempt := 1; attempt <= maxRetries; attempt++ {
		fmt.Fprintf(os.Stderr, "API调用尝试 %d/%d\n", attempt, maxRetries)

		request := DeepSeekRequest{
			Model: "deepseek-chat",
			Messages: []Message{
				{
					Role:    "system",
					Content: systemPrompt,
				},
				{
					Role: "user",
					Content: fmt.Sprintf(`请分析以下URL的安全性并规划渗透测试步骤：%s

请返回详细的分析报告和具体的工具调用列表。`, url),
				},
			},
			Stream: false,
		}

		requestBody, err := json.Marshal(request)
		if err != nil {
			fmt.Fprintf(os.Stderr, "序列化请求失败: %v\n", err)
			return "", fmt.Errorf("序列化请求失败: %w", err)
		}

		client := &http.Client{
			Timeout: 60 * time.Second,
		}

		req, err := http.NewRequest("POST", "https://api.deepseek.com/chat/completions", bytes.NewBuffer(requestBody))
		if err != nil {
			fmt.Fprintf(os.Stderr, "创建请求失败: %v\n", err)
			return "", fmt.Errorf("创建请求失败: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

		fmt.Fprintf(os.Stderr, "发送API请求到: %s\n", req.URL.String())
		startTime := time.Now()
		resp, err := client.Do(req)
		duration := time.Since(startTime)

		if err != nil {
			fmt.Fprintf(os.Stderr, "API请求失败 (尝试 %d/%d): %v (耗时: %v)\n", attempt, maxRetries, err, duration)
			if attempt < maxRetries {
				fmt.Fprintf(os.Stderr, "等待 %v 后重试...\n", retryDelay)
				time.Sleep(retryDelay)
				continue
			}
			return "", fmt.Errorf("发送请求失败: %w", err)
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "读取响应失败 (尝试 %d/%d): %v\n", attempt, maxRetries, err)
			if attempt < maxRetries {
				time.Sleep(retryDelay)
				continue
			}
			return "", fmt.Errorf("读取响应失败: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			fmt.Fprintf(os.Stderr, "API返回错误状态码: %d, 响应: %s\n", resp.StatusCode, string(respBody))
			return "", fmt.Errorf("API返回错误状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
		}

		var deepSeekResp DeepSeekResponse
		if err := json.Unmarshal(respBody, &deepSeekResp); err != nil {
			fmt.Fprintf(os.Stderr, "解析响应失败 (尝试 %d/%d): %v\n", attempt, maxRetries, err)
			if attempt < maxRetries {
				time.Sleep(retryDelay)
				continue
			}
			return "", fmt.Errorf("解析响应失败: %w", err)
		}

		if len(deepSeekResp.Choices) == 0 {
			fmt.Fprintf(os.Stderr, "API返回空结果\n")
			return "", fmt.Errorf("API返回空结果")
		}

		fmt.Fprintf(os.Stderr, "API调用成功 (尝试 %d/%d), 耗时: %v, 令牌使用: %d (提示) + %d (完成) = %d (总)\n",
			attempt, maxRetries, duration,
			deepSeekResp.Usage.PromptTokens,
			deepSeekResp.Usage.CompletionTokens,
			deepSeekResp.Usage.TotalTokens)

		return deepSeekResp.Choices[0].Message.Content, nil
	}

	fmt.Fprintf(os.Stderr, "API调用失败：达到最大重试次数\n")
	return "", fmt.Errorf("API调用失败：达到最大重试次数")
}

func SubmitCleanedOutputToDeepSeek(output string, apiKey string, iteration int) (string, error) {
	const maxRetries = 3
	const retryDelay = 2 * time.Second

	fmt.Fprintf(os.Stderr, "提交清理后的输出到DeepSeek，迭代 %d\n", iteration)

	for attempt := 1; attempt <= maxRetries; attempt++ {
		fmt.Fprintf(os.Stderr, "API调用尝试 %d/%d\n", attempt, maxRetries)

		request := DeepSeekRequest{
			Model: "deepseek-chat",
			Messages: []Message{
				{
					Role:    "user",
					Content: fmt.Sprintf(`请分析以下清理后的输出，并继续完成任务：\n\n%s\n\n这是第 %d 次迭代，请注意使用与之前不同的策略。`, output, iteration),
				},
			},
			Stream: false,
		}

		requestBody, err := json.Marshal(request)
		if err != nil {
			fmt.Fprintf(os.Stderr, "序列化请求失败: %v\n", err)
			return "", fmt.Errorf("序列化请求失败: %w", err)
		}

		client := &http.Client{
			Timeout: 60 * time.Second,
		}

		req, err := http.NewRequest("POST", "https://api.deepseek.com/chat/completions", bytes.NewBuffer(requestBody))
		if err != nil {
			fmt.Fprintf(os.Stderr, "创建请求失败: %v\n", err)
			return "", fmt.Errorf("创建请求失败: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

		fmt.Fprintf(os.Stderr, "发送API请求到: %s\n", req.URL.String())
		startTime := time.Now()
		resp, err := client.Do(req)
		duration := time.Since(startTime)

		if err != nil {
			fmt.Fprintf(os.Stderr, "API请求失败 (尝试 %d/%d): %v (耗时: %v)\n", attempt, maxRetries, err, duration)
			if attempt < maxRetries {
				fmt.Fprintf(os.Stderr, "等待 %v 后重试...\n", retryDelay)
				time.Sleep(retryDelay)
				continue
			}
			return "", fmt.Errorf("发送请求失败: %w", err)
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "读取响应失败 (尝试 %d/%d): %v\n", attempt, maxRetries, err)
			if attempt < maxRetries {
				time.Sleep(retryDelay)
				continue
			}
			return "", fmt.Errorf("读取响应失败: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			fmt.Fprintf(os.Stderr, "API返回错误状态码: %d, 响应: %s\n", resp.StatusCode, string(respBody))
			return "", fmt.Errorf("API返回错误状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
		}

		var deepSeekResp DeepSeekResponse
		if err := json.Unmarshal(respBody, &deepSeekResp); err != nil {
			fmt.Fprintf(os.Stderr, "解析响应失败 (尝试 %d/%d): %v\n", attempt, maxRetries, err)
			if attempt < maxRetries {
				time.Sleep(retryDelay)
				continue
			}
			return "", fmt.Errorf("解析响应失败: %w", err)
		}

		if len(deepSeekResp.Choices) == 0 {
			fmt.Fprintf(os.Stderr, "API返回空结果\n")
			return "", fmt.Errorf("API返回空结果")
		}

		fmt.Fprintf(os.Stderr, "API调用成功 (尝试 %d/%d), 耗时: %v, 令牌使用: %d (提示) + %d (完成) = %d (总)\n",
			attempt, maxRetries, duration,
			deepSeekResp.Usage.PromptTokens,
			deepSeekResp.Usage.CompletionTokens,
			deepSeekResp.Usage.TotalTokens)

		return deepSeekResp.Choices[0].Message.Content, nil
	}

	fmt.Fprintf(os.Stderr, "API调用失败：达到最大重试次数\n")
	return "", fmt.Errorf("API调用失败：达到最大重试次数")
}
