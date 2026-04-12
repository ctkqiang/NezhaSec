package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// 定义消息类型
type (
	ProgressMsg         string
	DeepSeekResponseMsg struct {
		Result string
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

// GetDeepSeekAPIKey 从环境变量获取DeepSeek API密钥
func GetDeepSeekAPIKey() string {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		// 尝试从.env文件读取
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

// CallDeepSeekAPI 调用DeepSeek API分析URL
// 参数列表：
//   - url 待分析的URL地址
//
// 返回值列表：
//   - tea.Cmd 一个命令，执行后会返回DeepSeekResponseMsg消息
func CallDeepSeekAPI(url string) tea.Cmd {
	// 创建一个默认的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	
	return CallDeepSeekAPIWithContext(ctx, url)
}

// CallDeepSeekAPIWithContext 调用DeepSeek API分析URL（带上下文）
// 参数列表：
//   - ctx 上下文，用于控制取消
//   - url 待分析的URL地址
//
// 返回值列表：
//   - tea.Cmd 一个命令，执行后会返回DeepSeekResponseMsg消息
func CallDeepSeekAPIWithContext(ctx context.Context, url string) tea.Cmd {
	return func() tea.Msg {
		apiKey := GetDeepSeekAPIKey()
		if apiKey == "" {
			return ProgressMsg("错误: 未配置DeepSeek API密钥")
		}

		// 启动API调用
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

		// 等待结果或取消
		select {
		case msg := <-resultChan:
			return msg
		case <-ctx.Done():
			return ProgressMsg("API调用已取消")
		}
	}
}

// progressWrapper 包装API调用，添加进度消息
func progressWrapper(fn func() tea.Msg) tea.Msg {
	// 执行API调用
	resultMsg := fn()

	// 如果是错误消息，直接返回
	if _, ok := resultMsg.(ProgressMsg); ok {
		return resultMsg
	}

	return resultMsg
}

// AnalyzeURLWithDeepSeek 使用DeepSeek API分析URL
// 参数列表：
//   - url 待分析的URL地址
//   - apiKey DeepSeek API密钥
//
// 返回值列表：
//   - string 分析结果
//   - error 可能的错误
func AnalyzeURLWithDeepSeek(url string, apiKey string) (string, error) {
	const maxRetries = 3
	const retryDelay = 2 * time.Second

	log.Printf("开始分析URL: %s", url)

	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Printf("API调用尝试 %d/%d", attempt, maxRetries)
		
		// 构建请求
		request := DeepSeekRequest{
			Model: "deepseek-chat",
			Messages: []Message{
				{
					Role:    "system",
					Content: "你是一个网络安全分析助手，负责分析给定URL的安全性。请提供详细的安全分析报告，包括潜在的漏洞、安全配置问题等。",
				},
				{
					Role:    "user",
					Content: fmt.Sprintf("请分析以下URL的安全性：%s", url),
				},
			},
			Stream: false,
		}

		// 序列化请求
		requestBody, err := json.Marshal(request)
		if err != nil {
			log.Printf("序列化请求失败: %v", err)
			return "", fmt.Errorf("序列化请求失败: %w", err)
		}

		// 创建HTTP请求
		client := &http.Client{
			Timeout: 60 * time.Second, // 增加超时时间
		}

		req, err := http.NewRequest("POST", "https://api.deepseek.com/chat/completions", bytes.NewBuffer(requestBody))
		if err != nil {
			log.Printf("创建请求失败: %v", err)
			return "", fmt.Errorf("创建请求失败: %w", err)
		}

		// 设置请求头
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

		// 发送请求
		log.Printf("发送API请求到: %s", req.URL.String())
		startTime := time.Now()
		resp, err := client.Do(req)
		duration := time.Since(startTime)
		
		if err != nil {
			log.Printf("API请求失败 (尝试 %d/%d): %v (耗时: %v)", attempt, maxRetries, err, duration)
			if attempt < maxRetries {
				log.Printf("等待 %v 后重试...", retryDelay)
				time.Sleep(retryDelay)
				continue
			}
			return "", fmt.Errorf("发送请求失败: %w", err)
		}
		defer resp.Body.Close()

		// 读取响应
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("读取响应失败 (尝试 %d/%d): %v", attempt, maxRetries, err)
			if attempt < maxRetries {
				time.Sleep(retryDelay)
				continue
			}
			return "", fmt.Errorf("读取响应失败: %w", err)
		}

		// 检查响应状态码
		if resp.StatusCode != http.StatusOK {
			log.Printf("API返回错误状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
			return "", fmt.Errorf("API返回错误状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
		}

		// 解析响应
		var deepSeekResp DeepSeekResponse
		if err := json.Unmarshal(respBody, &deepSeekResp); err != nil {
			log.Printf("解析响应失败 (尝试 %d/%d): %v", attempt, maxRetries, err)
			if attempt < maxRetries {
				time.Sleep(retryDelay)
				continue
			}
			return "", fmt.Errorf("解析响应失败: %w", err)
		}

		// 提取结果
		if len(deepSeekResp.Choices) == 0 {
			log.Printf("API返回空结果")
			return "", fmt.Errorf("API返回空结果")
		}

		log.Printf("API调用成功 (尝试 %d/%d), 耗时: %v, 令牌使用: %d (提示) + %d (完成) = %d (总)", 
			attempt, maxRetries, duration, 
			deepSeekResp.Usage.PromptTokens, 
			deepSeekResp.Usage.CompletionTokens, 
			deepSeekResp.Usage.TotalTokens)

		return deepSeekResp.Choices[0].Message.Content, nil
	}

	log.Printf("API调用失败：达到最大重试次数")
	return "", fmt.Errorf("API调用失败：达到最大重试次数")
}
