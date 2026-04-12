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

func processOutput() {

	terminalOutput := `开始执行工具调用...

执行工具 1: nmap
工具执行成功
输出: Starting Nmap 7.94 ( https://nmap.org ) at 2026-04-12 20:36 +08
Nmap scan report for localhost (127.0.0.1)
Host is up (0.000073s latency).
Other addresses for localhost (not scanned): ::1
Not shown: 90 closed tcp ports (conn-refused)
PORT     STATE SERVICE
22/tcp   open  ssh
53/tcp   open  domain
80/tcp   open  http
445/tcp  open  microsoft-ds
3306/tcp open  mysql
5000/tcp open  upnp
5432/tcp open  postgresql
5900/tcp open  vnc
8000/tcp open  http-alt
8080/tcp open  http-proxy

Nmap done: 1 IP address (1 host up) scanned in 0.31 seconds


执行工具 2: sqlmap
工具执行成功
输出:         ___
       __H__
 ___ ___[,]_____ ___ ___  {1.10.3#pip}
|_ -| . [)]     | .'| . |
|___|_  [)]_|_|_|__,|  _|
      |_|V...       |_|   https://sqlmap.org

[!] legal disclaimer: Usage of sqlmap for attacking targets without prior mutual consent is illegal. It is the end user's responsibility to obey all applicable local, state and federal laws. Developers assume no liability and are not responsible for any misuse or damage caused by this program

[*] starting @ 20:36:37 /2026-04-12/

[1/1] URL:
GET http://localhost
do you want to test this URL? [Y/n/q]
> Y
[20:36:37] [INFO] testing URL 'http://localhost'
[20:36:37] [INFO] using '/Users/johnmelodyme/.local/share/sqlmap/output/results-04122026_0836pm.csv' as the CSV results file in multiple targets mode
[20:36:37] [INFO] testing connection to the target URL
[20:36:38] [WARNING] the web server responded with an HTTP error code (502) which could interfere with the results of the tests
[20:36:38] [INFO] checking if the target is protected by some kind of WAF/IPS
[20:36:38] [INFO] testing if the target URL content is stable
[20:36:38] [INFO] target URL content is stable
[20:36:38] [ERROR] all tested parameters do not appear to be injectable. Try to increase values for '--level'/'--risk' options if you wish to perform more tests. If you suspect that there is some kind of protection mechanism involved (e.g. WAF) maybe you could try to use option '--tamper' (e.g. '--tamper=space2comment') and/or switch '--random-agent', skipping to the next target
[20:36:38] [WARNING] HTTP error codes detected during run:
502 (Bad Gateway) - 3 times
[20:36:38] [INFO] you can find results of scanning in multiple targets mode inside the CSV file '/Users/johnmelodyme/.local/share/sqlmap/output/results-04122026_0836pm.csv'

[*] ending @ 20:36:38 /2026-04-12/


分析完成`

	cleanedOutput := terminalOutput

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
	flag.StringVar(&targetURL, "u", "", "目标 URL 或 IP 地址")
	flag.BoolVar(&processMode, "process", false, "处理输出模式")
	flag.Parse()

	if processMode {

		processOutput()
		return
	}

	if targetURL == "" {
		fmt.Println("请使用 -u 参数指定目标 URL 或 IP 地址")
		fmt.Println("例如: ./nezha_sec -u https://example.com")
		return
	}

	log.Println("初始化应用...")

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
		log.Fatalf("目标不允许扫描: %s", reason)
	}

	log.Println("开始分析目标 URL...")
	progressMsg := orchestratorInstance.StartAnalysis()
	fmt.Println(string(progressMsg))

	log.Println("调用 DeepSeek API 进行分析...")
	result, err := api.AnalyzeURLWithDeepSeek(targetURL, api.GetDeepSeekAPIKey())
	if err != nil {
		log.Fatalf("API 调用失败: %v", err)
	}

	fmt.Println("\nAI 分析结果:")
	fmt.Println(result)

	toolCalls := parseToolCalls(result, targetURL)
	if len(toolCalls) > 0 {
		fmt.Println("\n工具调用列表:")
		for i, toolCall := range toolCalls {
			fmt.Printf("%d. %s\n", i+1, toolCall.ToolName)
			fmt.Printf("   参数: %v\n", toolCall.Arguments)
		}

		fmt.Println("\n开始执行工具调用...")
		for i, toolCall := range toolCalls {
			fmt.Printf("\n执行工具 %d: %s\n", i+1, toolCall.ToolName)
			result, err := orchestratorInstance.ExecuteTool(toolCall.ToolName, toolCall.Arguments)
			if err != nil {
				fmt.Printf("工具执行失败: %v\n", err)
			} else if !result.Success {
				fmt.Printf("工具执行失败: %s\n", result.Error)
			} else {
				fmt.Printf("工具执行成功\n")
				fmt.Printf("输出: %s\n", result.Output)
			}

			time.Sleep(1 * time.Second)
		}
	} else {
		fmt.Println("\n没有解析到工具调用")
	}

	fmt.Println("\n分析完成")
}

func parseToolCalls(response string, targetURL string) []api.ToolCallMsg {

	var toolCalls []api.ToolCallMsg

	toolCallsRegex := regexp.MustCompile(`tool_calls\s*=\s*\[(.*?)\]`)
	matches := toolCallsRegex.FindStringSubmatch(response)
	if len(matches) >= 2 {
		toolCallsJSON := "[" + matches[1] + "]"
		var parsedCalls []api.ToolCallMsg
		if err := json.Unmarshal([]byte(toolCallsJSON), &parsedCalls); err == nil {
			return parsedCalls
		}
	}

	altRegex := regexp.MustCompile(`工具调用列表\s*:\s*\[(.*?)\]`)
	matches = altRegex.FindStringSubmatch(response)
	if len(matches) >= 2 {
		toolCallsJSON := "[" + matches[1] + "]"
		var parsedCalls []api.ToolCallMsg
		if err := json.Unmarshal([]byte(toolCallsJSON), &parsedCalls); err == nil {
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
			if err := json.Unmarshal([]byte(argsStr), &args); err == nil {
				toolCalls = append(toolCalls, api.ToolCallMsg{
					ToolName:  toolName,
					Arguments: args,
				})
			}
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
			if err := json.Unmarshal([]byte(argsStr), &args); err == nil {

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
			if err := json.Unmarshal([]byte(paramsStr), &params); err == nil {

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

	toolsJSONRegex := regexp.MustCompile(`\{[\s\S]*?"tools"\s*:\s*\[(.*?)\]\s*\}`)
	toolsMatches := toolsJSONRegex.FindStringSubmatch(response)
	if len(toolsMatches) >= 2 {
		toolsJSON := "[" + toolsMatches[1] + "]"

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
						Arguments: tool.Parameters,
					})
				}
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
