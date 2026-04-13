package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"nezha_sec/internal/api"
	"nezha_sec/internal/orchestrator"
	"nezha_sec/internal/output"
	"nezha_sec/internal/registry"
	"nezha_sec/internal/views"
	"os"
	"regexp"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func processOutput() {
	fmt.Println("请输入终端输出内容，按 Ctrl+D 结束输入:")

	var terminalOutput strings.Builder
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		terminalOutput.WriteString(scanner.Text())
		terminalOutput.WriteString("\n")
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("读取输入失败: %v\n", err)
		return
	}

	cleanedOutput := terminalOutput.String()
	if cleanedOutput == "" {
		fmt.Println("错误: 未输入任何内容")
		return
	}

	apiKey := api.GetDeepSeekAPIKey()
	if apiKey == "" {
		fmt.Println("错误: 未配置DeepSeek API密钥")
		return
	}

	for i := 1; i <= 5; i++ {
		fmt.Printf("\n===== 第 %d 次迭代 =====\n", i)

		result, err := api.SubmitCleanedOutputToDeepSeek(cleanedOutput, apiKey, i)
		if err != nil {
			fmt.Printf("提交清理后的输出到DeepSeek失败: %v\n", err)
			continue
		}

		fmt.Println("\nDeepSeek 响应:")
		fmt.Println(result)

		cleanedOutput = result

		time.Sleep(2 * time.Second)
	}

	fmt.Println("\n===== 完成 5 次迭代 =====")
	fmt.Println("是否希望继续执行过程？")
	fmt.Println("1. 继续")
	fmt.Println("2. 停止")

	var choice int
	fmt.Scanln(&choice)

	if choice == 1 {
		processOutput()
	} else {
		fmt.Println("执行过程已停止")
	}
}

func main() {
	var targetURL string
	var processMode bool
	var mcpMode bool
	var workflowMode bool
	flag.StringVar(&targetURL, "u", "", "目标 URL 或 IP 地址")
	flag.BoolVar(&processMode, "process", false, "处理输出模式")
	flag.BoolVar(&mcpMode, "mcp", false, "使用 MCP 风格界面")
	flag.BoolVar(&workflowMode, "workflow", false, "使用红队工作流模式")
	flag.Parse()

	if processMode {

		processOutput()
		return
	}

	if targetURL == "" {
		if mcpMode {
			model, err := views.NewMCPModel()
			if err != nil {
				log.Fatalf("初始化 MCP 模型失败: %v", err)
			}

			p := tea.NewProgram(model)

			if err := p.Start(); err != nil {
				log.Fatalf("启动程序失败: %v", err)
			}
			return
		} else if workflowMode {
			toolRegistry, err := registry.NewToolRegistry()
			if err != nil {
				log.Fatalf("初始化工具注册表失败: %v", err)
			}

			orchestratorInstance, err := orchestrator.NewOrchestrator(toolRegistry)
			if err != nil {
				log.Fatalf("初始化调度器失败: %v", err)
			}

			model, err := views.NewCLIWorkflowModel(orchestratorInstance)
			if err != nil {
				log.Fatalf("初始化 CLI 工作流模型失败: %v", err)
			}

			p := tea.NewProgram(model)

			if err := p.Start(); err != nil {
				log.Fatalf("启动程序失败: %v", err)
			}
			return
		} else {
			pretty := output.NewPrettyOutput()
			pretty.PrintBanner()
			pretty.PrintUsage()
			return
		}
	}

	pretty := output.NewPrettyOutput()
	pretty.PrintBanner()
	pretty.PrintInitialization(targetURL)

	toolRegistry, err := registry.NewToolRegistry()
	if err != nil {
		log.Fatalf("初始化工具注册表失败: %v", err)
	}

	orchestratorInstance, err := orchestrator.NewOrchestrator(toolRegistry)
	if err != nil {
		log.Fatalf("初始化调度器失败: %v", err)
	}

	allowed, reason := orchestratorInstance.SetTarget(targetURL, true)
	if !allowed {
		pretty.PrintError(fmt.Sprintf("目标不允许扫描: %s", reason))
		return
	}

	pretty.PrintAnalysisStart()

	var result string
	var apiErr error
	maxRetries := 3

	for attempt := 1; attempt <= maxRetries; attempt++ {
		startTime := time.Now()
		result, apiErr = api.AnalyzeURLWithDeepSeek(targetURL, api.GetDeepSeekAPIKey())
		duration := time.Since(startTime)

		if apiErr != nil {
			pretty.PrintAPIError(attempt, maxRetries, apiErr)
			if attempt == maxRetries {
				log.Fatalf("API 调用失败: %v", apiErr)
			}
		} else {
			tokens := 1637
			pretty.PrintAPICall(attempt, maxRetries, duration, tokens)
			break
		}
	}

	pretty.PrintAIResultHeader()
	fmt.Println(result)

	toolCalls := parseToolCalls(result, targetURL)
	if len(toolCalls) > 0 {
		pretty.PrintToolCallsStart()
		for i, toolCall := range toolCalls {
			pretty.PrintToolCall(i+1, toolCall.ToolName, toolCall.Arguments)
		}

		pretty.PrintExecutionStart(len(toolCalls))
		for i, toolCall := range toolCalls {
			pretty.PrintToolExecutionStart(i+1, len(toolCalls), toolCall.ToolName)

			execResult, err := orchestratorInstance.ExecuteTool(toolCall.ToolName, toolCall.Arguments)
			if err != nil {
				pretty.PrintToolError(err.Error())
			} else if !execResult.Success {
				pretty.PrintToolError(execResult.Error)
			} else {
				if toolCall.ToolName == "nmap" {
					pretty.PrintToolSuccess("")
					pretty.PrintNmapResult(execResult.Output)
				} else {
					pretty.PrintToolSuccess(execResult.Output)
				}
			}

			time.Sleep(1 * time.Second)
		}
	} else {
		pretty.PrintWarning("没有解析到工具调用")
	}

	pretty.PrintCompletion()
}

func parseToolCalls(response string, targetURL string) []api.ToolCallMsg {
	var toolCalls []api.ToolCallMsg

	// 处理 tools JSON 格式（AI 常用的格式）
	toolsJSONRegex := regexp.MustCompile(`"tools"\s*:\s*\[(.*?)\]`)
	toolsMatches := toolsJSONRegex.FindAllStringSubmatch(response, -1)
	for _, match := range toolsMatches {
		if len(match) >= 2 {
			toolsJSON := "[" + match[1] + "]"

			type Tool struct {
				Name       string                 `json:"name"`
				Parameters map[string]interface{} `json:"parameters"`
			}
			var tools []Tool
			if err := json.Unmarshal([]byte(toolsJSON), &tools); err == nil {
				for _, tool := range tools {

					toolName := tool.Name
					if strings.HasPrefix(toolName, "run_") {
						toolName = strings.TrimPrefix(toolName, "run_")
					}

					params := tool.Parameters
					if params == nil || len(params) == 0 {
						params = make(map[string]interface{})
						switch toolName {
						case "curl", "sqlmap", "gobuster", "whatweb", "wpscan", "httpx", "nuclei", "ffuf", "trivy", "garak":
							params["url"] = targetURL
						case "nmap", "naabu":
							target := targetURL
							if strings.HasPrefix(target, "http://") {
								target = strings.TrimPrefix(target, "http://")
							} else if strings.HasPrefix(target, "https://") {
								target = strings.TrimPrefix(target, "https://")
							}
							if strings.Contains(target, "/") {
								target = strings.Split(target, "/")[0]
							}
							params["target"] = target
						}
					}

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
							Arguments: params,
						})
					}
				}
			}
		}
	}

	// 如果已经解析到工具，直接返回
	if len(toolCalls) > 0 {
		return toolCalls
	}

	// 处理 tool_calls 格式
	toolCallsRegex := regexp.MustCompile(`tool_calls\s*=\s*\[(.*?)\]`)
	matches := toolCallsRegex.FindStringSubmatch(response)
	if len(matches) >= 2 {
		toolCallsJSON := "[" + matches[1] + "]"
		var parsedCalls []api.ToolCallMsg
		if err := json.Unmarshal([]byte(toolCallsJSON), &parsedCalls); err == nil {
			for i := range parsedCalls {
				if len(parsedCalls[i].Arguments) == 0 {
					switch parsedCalls[i].ToolName {
					case "curl", "sqlmap", "gobuster", "whatweb", "wpscan", "httpx", "nuclei", "ffuf", "trivy", "garak":
						parsedCalls[i].Arguments = make(map[string]interface{})
						parsedCalls[i].Arguments["url"] = targetURL
					case "nmap":
						parsedCalls[i].Arguments = make(map[string]interface{})
						target := targetURL
						if strings.HasPrefix(target, "http://") {
							target = strings.TrimPrefix(target, "http://")
						} else if strings.HasPrefix(target, "https://") {
							target = strings.TrimPrefix(target, "https://")
						}
						if strings.Contains(target, "/") {
							target = strings.Split(target, "/")[0]
						}
						parsedCalls[i].Arguments["target"] = target
					}
				}
			}
			return parsedCalls
		}
	}

	// 处理工具调用列表格式
	altRegex := regexp.MustCompile(`工具调用列表\s*:\s*\[(.*?)\]`)
	matches = altRegex.FindStringSubmatch(response)
	if len(matches) >= 2 {
		toolCallsJSON := "[" + matches[1] + "]"
		var parsedCalls []api.ToolCallMsg
		if err := json.Unmarshal([]byte(toolCallsJSON), &parsedCalls); err == nil {
			for i := range parsedCalls {
				if len(parsedCalls[i].Arguments) == 0 {
					switch parsedCalls[i].ToolName {
					case "curl", "sqlmap", "gobuster", "whatweb", "wpscan", "httpx", "nuclei", "ffuf", "trivy", "garak":
						parsedCalls[i].Arguments = make(map[string]interface{})
						parsedCalls[i].Arguments["url"] = targetURL
					case "nmap":
						parsedCalls[i].Arguments = make(map[string]interface{})
						target := targetURL
						if strings.HasPrefix(target, "http://") {
							target = strings.TrimPrefix(target, "http://")
						} else if strings.HasPrefix(target, "https://") {
							target = strings.TrimPrefix(target, "https://")
						}
						if strings.Contains(target, "/") {
							target = strings.Split(target, "/")[0]
						}
						parsedCalls[i].Arguments["target"] = target
					}
				}
			}
			return parsedCalls
		}
	}

	singleToolRegex := regexp.MustCompile(`工具调用:\s*(\w+)\s*\{([^}]+)\}`)
	singleMatches := singleToolRegex.FindAllStringSubmatch(response, -1)
	for _, match := range singleMatches {
		if len(match) >= 3 {
			toolName := match[1]
			argsStr := "{" + match[2] + "}"

			argsStr = strings.ReplaceAll(argsStr, "url", "\"url\"")
			argsStr = strings.ReplaceAll(argsStr, "=", ":")
			argsStr = strings.ReplaceAll(argsStr, "https://", "https://")
			argsStr = strings.ReplaceAll(argsStr, "http://", "http://")
			argsStr = strings.ReplaceAll(argsStr, "", "\"")
			argsStr = strings.ReplaceAll(argsStr, ", ", ", ")

			var args map[string]interface{}
			if err := json.Unmarshal([]byte(argsStr), &args); err != nil {
				args = make(map[string]interface{})
				switch toolName {
				case "curl", "sqlmap", "gobuster", "whatweb", "wpscan", "httpx", "nuclei", "ffuf", "trivy", "garak":
					args["url"] = targetURL
				case "nmap":
					target := targetURL
					if strings.HasPrefix(target, "http://") {
						target = strings.TrimPrefix(target, "http://")
					} else if strings.HasPrefix(target, "https://") {
						target = strings.TrimPrefix(target, "https://")
					}
					if strings.Contains(target, "/") {
						target = strings.Split(target, "/")[0]
					}
					args["target"] = target
				}
			}
			toolCalls = append(toolCalls, api.ToolCallMsg{
				ToolName:  toolName,
				Arguments: args,
			})
		}
	}

	listToolRegex := regexp.MustCompile(`-\s*(\w+)\s*\{([^}]+)\}`)
	listMatches := listToolRegex.FindAllStringSubmatch(response, -1)
	for _, match := range listMatches {
		if len(match) >= 3 {
			toolName := match[1]
			argsStr := "{" + match[2] + "}"

			argsStr = strings.ReplaceAll(argsStr, "url", "\"url\"")
			argsStr = strings.ReplaceAll(argsStr, "=", ":")
			argsStr = strings.ReplaceAll(argsStr, "https://", "https://")
			argsStr = strings.ReplaceAll(argsStr, "http://", "http://")
			argsStr = strings.ReplaceAll(argsStr, "", "\"")
			argsStr = strings.ReplaceAll(argsStr, ", ", ", ")

			var args map[string]interface{}
			if err := json.Unmarshal([]byte(argsStr), &args); err != nil {
				args = make(map[string]interface{})
				switch toolName {
				case "curl", "sqlmap", "gobuster", "whatweb", "wpscan", "httpx", "nuclei", "ffuf", "trivy", "garak":
					args["url"] = targetURL
				case "nmap":
					target := targetURL
					if strings.HasPrefix(target, "http://") {
						target = strings.TrimPrefix(target, "http://")
					} else if strings.HasPrefix(target, "https://") {
						target = strings.TrimPrefix(target, "https://")
					}
					if strings.Contains(target, "/") {
						target = strings.Split(target, "/")[0]
					}
					args["target"] = target
				}
			}

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

	toolNameRegex := regexp.MustCompile(`\{[\s\S]*?"tool_name"\s*:\s*"(run_\w+)"[\s\S]*?"parameters"\s*:\s*\{([\s\S]*?)\}\s*\}`)
	toolNameMatches := toolNameRegex.FindAllStringSubmatch(response, -1)
	for _, match := range toolNameMatches {
		if len(match) >= 3 {
			toolName := match[1]
			paramsStr := "{" + match[2] + "}"

			if strings.HasPrefix(toolName, "run_") {
				toolName = strings.TrimPrefix(toolName, "run_")
			}

			var params map[string]interface{}
			if err := json.Unmarshal([]byte(paramsStr), &params); err != nil {
				params = make(map[string]interface{})
				switch toolName {
				case "curl", "sqlmap", "gobuster", "whatweb", "wpscan", "httpx", "nuclei", "ffuf", "trivy", "garak":
					params["url"] = targetURL
				case "nmap":
					target := targetURL
					if strings.HasPrefix(target, "http://") {
						target = strings.TrimPrefix(target, "http://")
					} else if strings.HasPrefix(target, "https://") {
						target = strings.TrimPrefix(target, "https://")
					}
					if strings.Contains(target, "/") {
						target = strings.Split(target, "/")[0]
					}
					params["target"] = target
				}
			}

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
					Arguments: params,
				})
			}
		}
	}

	runToolRegex := regexp.MustCompile(`工具名称:\s*run_(\w+)`)
	runMatches := runToolRegex.FindAllStringSubmatch(response, -1)
	for _, match := range runMatches {
		if len(match) >= 2 {
			toolName := match[1]

			args := make(map[string]interface{})

			paramRegex := regexp.MustCompile(`参数:\s*\{([^}]+)\}`)
			paramMatches := paramRegex.FindStringSubmatch(response)
			if len(paramMatches) >= 2 {
				paramStr := "{" + paramMatches[1] + "}"

				paramStr = strings.ReplaceAll(paramStr, "target", "\"target\"")
				paramStr = strings.ReplaceAll(paramStr, "ports", "\"ports\"")
				paramStr = strings.ReplaceAll(paramStr, "scan_type", "\"scan_type\"")
				paramStr = strings.ReplaceAll(paramStr, "=", ":")
				paramStr = strings.ReplaceAll(paramStr, "https://", "https://")
				paramStr = strings.ReplaceAll(paramStr, "http://", "http://")
				paramStr = strings.ReplaceAll(paramStr, "", "\"")
				paramStr = strings.ReplaceAll(paramStr, ", ", ", ")

				var paramArgs map[string]interface{}
				if err := json.Unmarshal([]byte(paramStr), &paramArgs); err == nil {
					args = paramArgs
				}
			}

			if len(args) == 0 {
				switch toolName {
				case "curl", "sqlmap", "gobuster", "whatweb", "wpscan", "httpx", "nuclei", "ffuf", "trivy", "garak":
					args["url"] = targetURL
				case "nmap":
					target := targetURL
					if strings.HasPrefix(target, "http://") {
						target = strings.TrimPrefix(target, "http://")
					} else if strings.HasPrefix(target, "https://") {
						target = strings.TrimPrefix(target, "https://")
					}
					if strings.Contains(target, "/") {
						target = strings.Split(target, "/")[0]
					}
					args["target"] = target
				}
			}

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

	if len(toolCalls) == 0 {

		commonTools := []string{"curl", "nmap", "sqlmap", "gobuster", "whatweb", "wpscan"}
		for _, tool := range commonTools {
			if strings.Contains(response, tool) {

				args := make(map[string]interface{})

				switch tool {
				case "curl", "sqlmap", "gobuster", "whatweb", "wpscan":
					args["url"] = targetURL
				case "nmap":

					target := targetURL
					if strings.HasPrefix(target, "http://") {
						target = strings.TrimPrefix(target, "http://")
					} else if strings.HasPrefix(target, "https://") {
						target = strings.TrimPrefix(target, "https://")
					}

					if strings.Contains(target, "/") {
						target = strings.Split(target, "/")[0]
					}
					args["target"] = target
				}
				toolCalls = append(toolCalls, api.ToolCallMsg{
					ToolName:  tool,
					Arguments: args,
				})
			}
		}
	}

	if len(toolCalls) == 0 {

		args := make(map[string]interface{})
		args["url"] = targetURL
		toolCalls = append(toolCalls, api.ToolCallMsg{
			ToolName:  "curl",
			Arguments: args,
		})
	}

	return toolCalls
}
