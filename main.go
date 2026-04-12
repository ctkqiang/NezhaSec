package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"nezha_sec/internal/api"
	"nezha_sec/internal/orchestrator"
	"nezha_sec/internal/registry"
	"regexp"
	"strings"
	"time"
)

func main() {
	// 解析命令行参数
	var targetURL string
	flag.StringVar(&targetURL, "u", "", "目标 URL 或 IP 地址")
	flag.Parse()

	// 检查是否提供了目标 URL
	if targetURL == "" {
		fmt.Println("请使用 -u 参数指定目标 URL 或 IP 地址")
		fmt.Println("例如: ./nezha_sec -u https://github.com")
		return
	}

	// 初始化应用
	log.Println("初始化应用...")

	// 创建工具注册表
	toolRegistry, err := registry.NewToolRegistry()
	if err != nil {
		log.Fatalf("初始化工具注册表失败: %v", err)
	}

	// 创建调度器
	orchestratorInstance, err := orchestrator.NewOrchestrator(toolRegistry)
	if err != nil {
		log.Fatalf("初始化调度器失败: %v", err)
	}

	// 设置目标 URL
	allowed, reason := orchestratorInstance.SetTarget(targetURL, true)
	if !allowed {
		log.Fatalf("目标不允许扫描: %s", reason)
	}

	// 开始分析
	log.Println("开始分析目标 URL...")
	progressMsg := orchestratorInstance.StartAnalysis()
	fmt.Println(string(progressMsg))

	// 调用 DeepSeek API 进行分析
	log.Println("调用 DeepSeek API 进行分析...")
	result, err := api.AnalyzeURLWithDeepSeek(targetURL, api.GetDeepSeekAPIKey())
	if err != nil {
		log.Fatalf("API 调用失败: %v", err)
	}

	// 显示分析结果
	fmt.Println("\nAI 分析结果:")
	fmt.Println(result)

	// 解析工具调用
	toolCalls := parseToolCalls(result, targetURL)
	if len(toolCalls) > 0 {
		fmt.Println("\n工具调用列表:")
		for i, toolCall := range toolCalls {
			fmt.Printf("%d. %s\n", i+1, toolCall.ToolName)
			fmt.Printf("   参数: %v\n", toolCall.Arguments)
		}

		// 执行工具调用
		fmt.Println("\n开始执行工具调用...")
		for i, toolCall := range toolCalls {
			fmt.Printf("\n执行工具 %d: %s\n", i+1, toolCall.ToolName)
			result, err := orchestratorInstance.ExecuteTool(toolCall.ToolName, toolCall.Arguments)
			if err != nil {
				fmt.Printf("工具执行失败: %v\n", err)
			} else {
				fmt.Printf("工具执行成功\n")
				fmt.Printf("输出: %s\n", result.Output)
			}
			// 避免过于频繁的 API 调用
			time.Sleep(1 * time.Second)
		}
	} else {
		fmt.Println("\n没有解析到工具调用")
	}

	// 完成分析
	fmt.Println("\n分析完成")
}

// 解析AI响应中的工具调用
func parseToolCalls(response string, targetURL string) []api.ToolCallMsg {
	// 初始化工具调用列表
	var toolCalls []api.ToolCallMsg

	// 尝试多种格式解析工具调用
	// 格式1: tool_calls = [{"tool_name": "...", "arguments": {...}}]
	toolCallsRegex := regexp.MustCompile(`tool_calls\s*=\s*\[(.*?)\]`)
	matches := toolCallsRegex.FindStringSubmatch(response)
	if len(matches) >= 2 {
		toolCallsJSON := "[" + matches[1] + "]"
		var parsedCalls []api.ToolCallMsg
		if err := json.Unmarshal([]byte(toolCallsJSON), &parsedCalls); err == nil {
			return parsedCalls
		}
	}

	// 格式2: 工具调用列表: [{"tool_name": "...", "arguments": {...}}]
	altRegex := regexp.MustCompile(`工具调用列表\s*:\s*\[(.*?)\]`)
	matches = altRegex.FindStringSubmatch(response)
	if len(matches) >= 2 {
		toolCallsJSON := "[" + matches[1] + "]"
		var parsedCalls []api.ToolCallMsg
		if err := json.Unmarshal([]byte(toolCallsJSON), &parsedCalls); err == nil {
			return parsedCalls
		}
	}

	// 格式3: 工具调用: 工具名称 {参数}
	singleToolRegex := regexp.MustCompile(`工具调用:\s*(\w+)\s*\{([^}]+)\}`)
	singleMatches := singleToolRegex.FindAllStringSubmatch(response, -1)
	for _, match := range singleMatches {
		if len(match) >= 3 {
			toolName := match[1]
			argsStr := "{" + match[2] + "}"

			// 清理argsStr
			argsStr = strings.ReplaceAll(argsStr, "url", "\"url\"")
			argsStr = strings.ReplaceAll(argsStr, "=", ":")
			argsStr = strings.ReplaceAll(argsStr, "https://", "\"https://")
			argsStr = strings.ReplaceAll(argsStr, "http://", "\"http://")
			argsStr = strings.ReplaceAll(argsStr, "", "\"")
			argsStr = strings.ReplaceAll(argsStr, ", ", ", ")

			var args map[string]interface{}
			if err := json.Unmarshal([]byte(argsStr), &args); err == nil {
				toolCalls = append(toolCalls, api.ToolCallMsg{
					ToolName:  toolName,
					Arguments: args,
				})
			}
		}
	}

	// 格式4: - 工具名称 {参数}
	listToolRegex := regexp.MustCompile(`-\s*(\w+)\s*\{([^}]+)\}`)
	listMatches := listToolRegex.FindAllStringSubmatch(response, -1)
	for _, match := range listMatches {
		if len(match) >= 3 {
			toolName := match[1]
			argsStr := "{" + match[2] + "}"

			// 清理argsStr
			argsStr = strings.ReplaceAll(argsStr, "url", "\"url\"")
			argsStr = strings.ReplaceAll(argsStr, "=", ":")
			argsStr = strings.ReplaceAll(argsStr, "https://", "\"https://")
			argsStr = strings.ReplaceAll(argsStr, "http://", "\"http://")
			argsStr = strings.ReplaceAll(argsStr, "", "\"")
			argsStr = strings.ReplaceAll(argsStr, ", ", ", ")

			var args map[string]interface{}
			if err := json.Unmarshal([]byte(argsStr), &args); err == nil {
				// 检查是否已经添加过相同的工具调用
				exists := false
				for _, existingCall := range toolCalls {
					if existingCall.ToolName == toolName {
						exists = true
						break
					}
				}
				if !exists {
					toolCalls = append(toolCalls, api.ToolCallMsg{
						ToolName:  toolName,
						Arguments: args,
					})
				}
			}
		}
	}

	// 格式5: 工具名称: run_* 格式
	runToolRegex := regexp.MustCompile(`工具名称:\s*run_(\w+)`)
	runMatches := runToolRegex.FindAllStringSubmatch(response, -1)
	for _, match := range runMatches {
		if len(match) >= 2 {
			toolName := "run_" + match[1]
			// 为工具创建默认参数
			args := make(map[string]interface{})
			// 尝试从响应中提取参数
			paramRegex := regexp.MustCompile(`参数:\s*\{([^}]+)\}`)
			paramMatches := paramRegex.FindStringSubmatch(response)
			if len(paramMatches) >= 2 {
				paramStr := "{" + paramMatches[1] + "}"
				// 清理paramStr
				paramStr = strings.ReplaceAll(paramStr, "target", "\"target\"")
				paramStr = strings.ReplaceAll(paramStr, "ports", "\"ports\"")
				paramStr = strings.ReplaceAll(paramStr, "scan_type", "\"scan_type\"")
				paramStr = strings.ReplaceAll(paramStr, "=", ":")
				paramStr = strings.ReplaceAll(paramStr, "https://", "\"https://")
				paramStr = strings.ReplaceAll(paramStr, "http://", "\"http://")
				paramStr = strings.ReplaceAll(paramStr, "", "\"")
				paramStr = strings.ReplaceAll(paramStr, ", ", ", ")

				var paramArgs map[string]interface{}
				if err := json.Unmarshal([]byte(paramStr), &paramArgs); err == nil {
					args = paramArgs
				}
			}
			// 检查是否已经添加过相同的工具调用
			exists := false
			for _, existingCall := range toolCalls {
				if existingCall.ToolName == toolName {
					exists = true
					break
				}
			}
			if !exists {
				toolCalls = append(toolCalls, api.ToolCallMsg{
					ToolName:  toolName,
					Arguments: args,
				})
			}
		}
	}

	// 如果没有解析到工具调用，尝试从响应中提取常见工具名称
	if len(toolCalls) == 0 {
		// 常见工具名称列表
		commonTools := []string{"curl", "nmap", "sqlmap", "gobuster", "whatweb", "wpscan"}
		for _, tool := range commonTools {
			if strings.Contains(response, tool) {
				// 为常见工具创建默认参数
				args := make(map[string]interface{})
				// 这里可以根据工具类型设置默认参数
				toolCalls = append(toolCalls, api.ToolCallMsg{
					ToolName:  tool,
					Arguments: args,
				})
			}
		}
	}

	// 如果仍然没有解析到工具调用，创建默认的工具调用
	if len(toolCalls) == 0 {
		// 创建一个默认的curl工具调用
		args := make(map[string]interface{})
		args["url"] = targetURL
		toolCalls = append(toolCalls, api.ToolCallMsg{
			ToolName:  "curl",
			Arguments: args,
		})
	}

	return toolCalls
}
