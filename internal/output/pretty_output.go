package output

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
)

// PrettyOutput 美化输出结构体
type PrettyOutput struct {
	lightBlue    *color.Color
	green        *color.Color
	yellow       *color.Color
	cyan         *color.Color
	white        *color.Color
	lightYellow  *color.Color
	lightMagenta *color.Color
	lightGreen   *color.Color
	lightCyan    *color.Color
	red          *color.Color
	lightWhite   *color.Color
	lightBlack   *color.Color
}

// NewPrettyOutput 创建新的美化输出实例
func NewPrettyOutput() *PrettyOutput {
	return &PrettyOutput{
		lightBlue:    color.New(color.FgHiBlue),
		green:        color.New(color.FgGreen),
		yellow:       color.New(color.FgYellow),
		cyan:         color.New(color.FgCyan),
		white:        color.New(color.FgWhite),
		lightYellow:  color.New(color.FgHiYellow),
		lightMagenta: color.New(color.FgHiMagenta),
		lightGreen:   color.New(color.FgHiGreen),
		lightCyan:    color.New(color.FgHiCyan),
		red:          color.New(color.FgRed),
		lightWhite:   color.New(color.FgHiWhite),
		lightBlack:   color.New(color.FgHiBlack),
	}
}

// getAppName 从 .env 文件获取应用名称
func (p *PrettyOutput) getAppName() string {
	_ = godotenv.Load(".env")
	appName := os.Getenv("APP_NAME")
	if appName == "" {
		appName = "哪吒网络安全分析器"
	}
	return appName
}

// getAuthor 获取作者信息
func (p *PrettyOutput) getAuthor() string {
	return "钟智强"
}

// PrintBanner 打印程序横幅
func (p *PrettyOutput) PrintBanner() {
	appName := p.getAppName()
	author := p.getAuthor()

	banner := fmt.Sprintf(`
╔══════════════════════════════════════════════════════════════╗
║                                                              
║     %s - AI 驱动的红队工具              
║                                                              
║     Powered by DeepSeek  •  作者: %s                          
║                                                              
╚══════════════════════════════════════════════════════════════╝
`, appName, author)
	fmt.Println(p.lightBlue.Sprint(banner))
}

// PrintInitialization 打印初始化信息
func (p *PrettyOutput) PrintInitialization(target string) {
	header := p.green.Sprint("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	fmt.Println()
	fmt.Println(header)
	fmt.Println(p.yellow.Sprint("[+] 初始化应用..."))
	fmt.Printf(p.cyan.Sprint("[*] 目标: "))
	fmt.Printf("%s\n", p.white.Sprint(target))
	fmt.Println(header)
	fmt.Println()
}

// PrintAnalysisStart 打印分析开始信息
func (p *PrettyOutput) PrintAnalysisStart() {
	fmt.Println(p.lightYellow.Sprint("[+] 开始分析目标 URL..."))
	fmt.Println(p.lightMagenta.Sprint("[*] 调用 DeepSeek API 进行分析..."))
	fmt.Println(p.lightGreen.Sprint("[>] 发送请求中..."))
	fmt.Println()
}

// PrintAPICall 打印 API 调用成功信息
func (p *PrettyOutput) PrintAPICall(attempt, maxAttempts int, duration time.Duration, tokens int) {
	fmt.Printf("%s", p.lightBlue.Sprintf("[OK] API调用成功 (尝试 %d/%d), ", attempt, maxAttempts))
	fmt.Printf("%s", p.lightGreen.Sprintf("耗时: %v, ", duration))
	fmt.Printf("%s", p.lightYellow.Sprintf("令牌: %d\n", tokens))
}

// PrintAPIError 打印 API 调用错误信息
func (p *PrettyOutput) PrintAPIError(attempt, maxAttempts int, err error) {
	fmt.Printf("%s", p.red.Sprintf("[ERR] API调用失败 (尝试 %d/%d): %v\n", attempt, maxAttempts, err))
	if attempt < maxAttempts {
		fmt.Println(p.yellow.Sprint("[...] 等待重试..."))
		time.Sleep(2 * time.Second)
	}
}

// PrintAIResultHeader 打印 AI 结果头部
func (p *PrettyOutput) PrintAIResultHeader() {
	fmt.Println()
	header := p.lightMagenta.Sprint("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println(p.lightCyan.Sprint("[+] AI 分析结果"))
	fmt.Println(header)
	fmt.Println()
}

// PrintToolCallsStart 打印工具调用开始信息
func (p *PrettyOutput) PrintToolCallsStart() {
	fmt.Println()
	header := p.lightMagenta.Sprint("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println(p.lightGreen.Sprint("[+] 工具调用列表"))
	fmt.Println(header)
}

// PrintToolCall 打印单个工具调用信息
func (p *PrettyOutput) PrintToolCall(index int, toolName string, arguments map[string]interface{}) {
	fmt.Printf("\n")
	fmt.Printf(p.cyan.Sprintf("  ┌─ 工具 #%d\n", index))
	fmt.Printf(p.lightWhite.Sprintf("  | 名称: %s\n", toolName))
	fmt.Printf(p.lightWhite.Sprint("  | 参数: "))

	argsStr := p.formatArguments(arguments)
	fmt.Printf("%s\n", p.yellow.Sprint(argsStr))
	fmt.Printf(p.cyan.Sprint("  └─────────────────────────────────\n"))
}

// formatArguments 格式化参数
func (p *PrettyOutput) formatArguments(args map[string]interface{}) string {
	if len(args) == 0 {
		return "{}"
	}

	var parts []string
	for k, v := range args {
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

// PrintExecutionStart 打印执行开始信息
func (p *PrettyOutput) PrintExecutionStart(totalTools int) {
	fmt.Println()
	header := p.lightGreen.Sprint("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf(p.lightYellow.Sprintf("[>] 开始执行工具调用... (共 %d 个工具)\n", totalTools))
	fmt.Println(header)
	fmt.Println()
}

// PrintToolExecutionStart 打印工具执行开始信息
func (p *PrettyOutput) PrintToolExecutionStart(index, total int, toolName string) {
	fmt.Printf(p.cyan.Sprintf("\n  ┌─ 执行中 [%d/%d]: %s\n", index, total, toolName))
	fmt.Printf(p.lightWhite.Sprint("  |"))
}

// PrintToolSuccess 打印工具执行成功信息
func (p *PrettyOutput) PrintToolSuccess(output string) {
	fmt.Println(p.lightGreen.Sprint(" [OK]"))
	fmt.Println(p.lightGreen.Sprint("  |"))
	fmt.Println(p.lightGreen.Sprint("  |  [OK] 执行成功"))
	fmt.Println(p.lightGreen.Sprint("  |"))
	fmt.Println(p.lightGreen.Sprint("  ├─ 输出:"))

	outputLines := strings.Split(output, "\n")
	for _, line := range outputLines {
		if strings.TrimSpace(line) != "" {
			fmt.Printf(p.white.Sprintf("  |  %s\n", line))
		}
	}
	fmt.Printf(p.cyan.Sprint("  └─────────────────────────────────\n"))
}

// PrintToolError 打印工具执行错误信息
func (p *PrettyOutput) PrintToolError(errMsg string) {
	fmt.Println(p.red.Sprint(" [ERR]"))
	fmt.Println(p.red.Sprint("  |"))
	fmt.Printf(p.red.Sprintf("  |  [ERR] 执行失败: %s\n", errMsg))
	fmt.Printf(p.cyan.Sprint("  └─────────────────────────────────\n"))
}

// PrintProgressBar 打印进度条
func (p *PrettyOutput) PrintProgressBar(current, total int, label string) {
	width := 50
	percent := float64(current) / float64(total)
	filled := int(percent * float64(width))

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)

	fmt.Printf("\r  %s [%s] %d%%", label, bar, int(percent*100))

	if current == total {
		fmt.Println()
	}
}

// PrintCompletion 打印完成信息
func (p *PrettyOutput) PrintCompletion() {
	fmt.Println()
	header := p.lightGreen.Sprint("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println(p.lightGreen.Sprint("[OK] 分析完成"))
	fmt.Println(header)
	fmt.Println()
	fmt.Println(p.lightCyan.Sprint("[*] 提示: 使用 -workflow 参数启动红队工作流模式获取更多功能"))
	fmt.Println()
}

// PrintError 打印错误信息
func (p *PrettyOutput) PrintError(msg string) {
	fmt.Println(p.red.Sprint("[ERR] 错误:") + " " + msg)
}

// PrintWarning 打印警告信息
func (p *PrettyOutput) PrintWarning(msg string) {
	fmt.Println(p.yellow.Sprint("[!] 警告:") + " " + msg)
}

// PrintInfo 打印信息
func (p *PrettyOutput) PrintInfo(msg string) {
	fmt.Println(p.cyan.Sprint("[*] 信息:") + " " + msg)
}

// PrintSuccess 打印成功信息
func (p *PrettyOutput) PrintSuccess(msg string) {
	fmt.Println(p.lightGreen.Sprint("[OK] 成功:") + " " + msg)
}

// PrintSeparator 打印分隔线
func (p *PrettyOutput) PrintSeparator() {
	fmt.Println(p.lightBlack.Sprint("────────────────────────────────────────────────────────"))
}

// PrintNmapResult 打印 Nmap 扫描结果
func (p *PrettyOutput) PrintNmapResult(output string) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.Contains(line, "PORT") || strings.Contains(line, "STATE") {
			fmt.Println(p.lightYellow.Sprint("  " + line))
		} else if strings.Contains(line, "open") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				port := p.lightGreen.Sprint(parts[0])
				state := p.green.Sprint(parts[1])
				service := p.cyan.Sprint(strings.Join(parts[2:], " "))
				fmt.Printf("  %s %s %s\n", port, state, service)
			} else {
				fmt.Println(p.white.Sprint("  " + line))
			}
		} else if strings.Contains(line, "Nmap done") {
			fmt.Println(p.lightMagenta.Sprint("  " + line))
		} else {
			fmt.Println(p.white.Sprint("  " + line))
		}
	}
}

// PrintUsage 打印使用方法
func (p *PrettyOutput) PrintUsage() {
	appName := p.getAppName()
	usage := fmt.Sprintf(`
╔══════════════════════════════════════════════════════════════╗
║                    使用方法                                    ║
╠══════════════════════════════════════════════════════════════╣
║  1. 基础扫描模式:                                              ║
║     ./nezha_sec -u https://example.com                        ║
║                                                                  ║
║  2. 红队工作流模式 (美化界面):                                   ║
║     ./nezha_sec -workflow                                     ║
║                                                                  ║
║  3. MCP 风格界面:                                              ║
║     ./nezha_sec -mcp                                          ║
║                                                                  ║
║  4. 处理输出模式:                                               ║
║     ./nezha_sec -process                                      ║
╚══════════════════════════════════════════════════════════════╝
`, appName)
	fmt.Println(p.lightBlue.Sprint(usage))
}
