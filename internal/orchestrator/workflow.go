package orchestrator

import (
	"context"
	"fmt"
	"log"
	"nezha_sec/internal/model"
	"time"
)

type WorkflowPhase string

const (
	PhaseReconnaissance      WorkflowPhase = "reconnaissance"
	PhaseVulnerabilityScan   WorkflowPhase = "vulnerability_scan"
	PhaseExploitation        WorkflowPhase = "exploitation"
	PhasePrivilegeEscalation WorkflowPhase = "privilege_escalation"
	PhaseLateralMovement     WorkflowPhase = "lateral_movement"
	PhasePostExploitation    WorkflowPhase = "post_exploitation"
	PhaseReporting           WorkflowPhase = "reporting"
)

type WorkflowState struct {
	CurrentPhase WorkflowPhase
	Target       string
	Results      map[WorkflowPhase]*model.ExecutionResult
	Context      map[string]interface{}
	StartTime    time.Time
	EndTime      time.Time
}

type WorkflowManager struct {
	orchestrator *Orchestrator
	state        *WorkflowState
	config       map[string]interface{}
}

func NewWorkflowManager(orchestrator *Orchestrator) *WorkflowManager {
	return &WorkflowManager{
		orchestrator: orchestrator,
		state: &WorkflowState{
			CurrentPhase: PhaseReconnaissance,
			Results:      make(map[WorkflowPhase]*model.ExecutionResult),
			Context:      make(map[string]interface{}),
			StartTime:    time.Now(),
		},
		config: make(map[string]interface{}),
	}
}

func (wm *WorkflowManager) SetTarget(target string) error {
	wm.state.Target = target
	return nil
}

func (wm *WorkflowManager) SetConfig(config map[string]interface{}) {
	wm.config = config
}

func (wm *WorkflowManager) GetState() *WorkflowState {
	return wm.state
}

func (wm *WorkflowManager) ExecuteWorkflow(ctx context.Context) error {
	phases := []WorkflowPhase{
		PhaseReconnaissance,
		PhaseVulnerabilityScan,
		PhaseExploitation,
		PhasePrivilegeEscalation,
		PhaseLateralMovement,
		PhasePostExploitation,
		PhaseReporting,
	}

	for _, phase := range phases {
		wm.state.CurrentPhase = phase
		log.Printf("开始阶段: %s", phase)

		err := wm.executePhase(ctx, phase)
		if err != nil {
			log.Printf("阶段执行失败 %s: %v", phase, err)
			// 继续执行下一阶段，不中断整个工作流
			continue
		}

		// 检查是否需要跳过后续阶段
		if wm.shouldSkipNextPhases() {
			log.Println("根据当前结果，跳过后续阶段")
			break
		}

		// 短暂延迟，避免工具执行过于密集
		time.Sleep(1 * time.Second)
	}

	wm.state.EndTime = time.Now()
	log.Println("工作流执行完成")
	return nil
}

func (wm *WorkflowManager) executePhase(ctx context.Context, phase WorkflowPhase) error {
	switch phase {
	case PhaseReconnaissance:
		return wm.executeReconnaissance(ctx)
	case PhaseVulnerabilityScan:
		return wm.executeVulnerabilityScan(ctx)
	case PhaseExploitation:
		return wm.executeExploitation(ctx)
	case PhasePrivilegeEscalation:
		return wm.executePrivilegeEscalation(ctx)
	case PhaseLateralMovement:
		return wm.executeLateralMovement(ctx)
	case PhasePostExploitation:
		return wm.executePostExploitation(ctx)
	case PhaseReporting:
		return wm.executeReporting(ctx)
	default:
		return fmt.Errorf("未知阶段: %s", phase)
	}
}

func (wm *WorkflowManager) executeReconnaissance(ctx context.Context) error {
	log.Println("执行侦察阶段...")

	// 1. 使用 subfinder 发现子域名
	subfinderArgs := map[string]interface{}{
		"domain": wm.extractDomain(wm.state.Target),
	}
	subfinderResult, err := wm.orchestrator.ExecuteTool("subfinder", subfinderArgs)
	if err != nil {
		log.Printf("subfinder 执行失败: %v", err)
	} else {
		wm.state.Results[PhaseReconnaissance] = subfinderResult
		wm.state.Context["subdomains"] = subfinderResult.Output
	}

	// 2. 使用 amass 进行深度资产发现
	amassArgs := map[string]interface{}{
		"domain":  wm.extractDomain(wm.state.Target),
		"passive": true,
	}
	amassResult, err := wm.orchestrator.ExecuteTool("amass", amassArgs)
	if err != nil {
		log.Printf("amass 执行失败: %v", err)
	} else {
		wm.state.Context["assets"] = amassResult.Output
	}

	// 3. 使用 httpx 进行 HTTP 探测
	httpxArgs := map[string]interface{}{
		"targets":     wm.state.Target,
		"probe":       true,
		"status_code": true,
		"tech_detect": true,
	}
	httpxResult, err := wm.orchestrator.ExecuteTool("httpx", httpxArgs)
	if err != nil {
		log.Printf("httpx 执行失败: %v", err)
	} else {
		wm.state.Context["http_info"] = httpxResult.Output
	}

	// 4. 使用 nmap 进行端口扫描
	nmapArgs := map[string]interface{}{
		"target":    wm.extractDomain(wm.state.Target),
		"ports":     "1-1000",
		"scan_type": "service",
	}
	nmapResult, err := wm.orchestrator.ExecuteTool("nmap", nmapArgs)
	if err != nil {
		log.Printf("nmap 执行失败: %v", err)
	} else {
		wm.state.Context["ports"] = nmapResult.Output
	}

	return nil
}

func (wm *WorkflowManager) executeVulnerabilityScan(ctx context.Context) error {
	log.Println("执行漏洞扫描阶段...")

	// 1. 使用 nuclei 进行漏洞扫描
	nucleiArgs := map[string]interface{}{
		"target":    wm.state.Target,
		"templates": "cves,exposures,misconfigurations",
	}
	nucleiResult, err := wm.orchestrator.ExecuteTool("nuclei", nucleiArgs)
	if err != nil {
		log.Printf("nuclei 执行失败: %v", err)
	} else {
		wm.state.Results[PhaseVulnerabilityScan] = nucleiResult
		wm.state.Context["vulnerabilities"] = nucleiResult.Output
	}

	// 2. 使用 ffuf 进行目录爆破
	ffufArgs := map[string]interface{}{
		"url":        wm.state.Target + "/FUZZ",
		"wordlist":   "common.txt",
		"extensions": "php,html,js,json,txt",
	}
	ffufResult, err := wm.orchestrator.ExecuteTool("ffuf", ffufArgs)
	if err != nil {
		log.Printf("ffuf 执行失败: %v", err)
	} else {
		wm.state.Context["directories"] = ffufResult.Output
	}

	// 3. 使用 sqlmap 进行 SQL 注入检测
	sqlmapArgs := map[string]interface{}{
		"url":   wm.state.Target,
		"crawl": 2,
		"risk":  1,
		"level": 1,
	}
	sqlmapResult, err := wm.orchestrator.ExecuteTool("sqlmap", sqlmapArgs)
	if err != nil {
		log.Printf("sqlmap 执行失败: %v", err)
	} else {
		wm.state.Context["sql_injection"] = sqlmapResult.Output
	}

	return nil
}

func (wm *WorkflowManager) executeExploitation(ctx context.Context) error {
	log.Println("执行利用阶段...")

	// 基于之前的扫描结果，选择合适的利用工具
	vulnerabilities, ok := wm.state.Context["vulnerabilities"].(string)
	if ok && vulnerabilities != "" {
		// 这里可以根据漏洞类型选择对应的利用工具
		log.Println("基于漏洞扫描结果执行利用...")
	}

	// 示例：使用 commix 进行命令注入检测
	commixArgs := map[string]interface{}{
		"url":   wm.state.Target,
		"level": 1,
	}
	commixResult, err := wm.orchestrator.ExecuteTool("commix", commixArgs)
	if err != nil {
		log.Printf("commix 执行失败: %v", err)
	} else {
		wm.state.Results[PhaseExploitation] = commixResult
		wm.state.Context["command_injection"] = commixResult.Output
	}

	return nil
}

func (wm *WorkflowManager) executePrivilegeEscalation(ctx context.Context) error {
	log.Println("执行权限提升阶段...")

	// 这里可以集成权限提升工具，如 Metasploit
	msfArgs := map[string]interface{}{
		"command": "search exploit/unix",
	}
	msfResult, err := wm.orchestrator.ExecuteTool("msfconsole", msfArgs)
	if err != nil {
		log.Printf("msfconsole 执行失败: %v", err)
	} else {
		wm.state.Results[PhasePrivilegeEscalation] = msfResult
		wm.state.Context["privilege_escalation"] = msfResult.Output
	}

	return nil
}

func (wm *WorkflowManager) executeLateralMovement(ctx context.Context) error {
	log.Println("执行横向移动阶段...")

	// 这里可以集成横向移动工具，如 impacket、responder 等
	impacketArgs := map[string]interface{}{
		"command": "smbclient -L //127.0.0.1",
	}
	impacketResult, err := wm.orchestrator.ExecuteTool("impacket", impacketArgs)
	if err != nil {
		log.Printf("impacket 执行失败: %v", err)
	} else {
		wm.state.Results[PhaseLateralMovement] = impacketResult
		wm.state.Context["lateral_movement"] = impacketResult.Output
	}

	return nil
}

func (wm *WorkflowManager) executePostExploitation(ctx context.Context) error {
	log.Println("执行后利用阶段...")

	// 这里可以集成后利用工具，如 sliver-cli、havoc 等
	sliverArgs := map[string]interface{}{
		"command": "sessions",
	}
	sliverResult, err := wm.orchestrator.ExecuteTool("sliver-cli", sliverArgs)
	if err != nil {
		log.Printf("sliver-cli 执行失败: %v", err)
	} else {
		wm.state.Results[PhasePostExploitation] = sliverResult
		wm.state.Context["post_exploitation"] = sliverResult.Output
	}

	return nil
}

func (wm *WorkflowManager) executeReporting(ctx context.Context) error {
	log.Println("执行报告阶段...")

	// 生成综合报告
	report := fmt.Sprintf(`# 渗透测试报告

## 执行摘要
- 目标: %s
- 开始时间: %s
- 结束时间: %s
- 持续时间: %s

## 侦察阶段
- 子域名: %v
- 资产信息: %v
- HTTP 信息: %v
- 开放端口: %v

## 漏洞扫描阶段
- 发现的漏洞: %v
- 目录结构: %v
- SQL 注入: %v

## 利用阶段
- 命令注入: %v

## 权限提升阶段
- 权限提升结果: %v

## 横向移动阶段
- 横向移动结果: %v

## 后利用阶段
- 后利用结果: %v

## 结论与建议
- 建议修复发现的所有漏洞
- 加强网络安全防护措施
- 定期进行安全评估
`,
		wm.state.Target,
		wm.state.StartTime.Format(time.RFC3339),
		wm.state.EndTime.Format(time.RFC3339),
		wm.state.EndTime.Sub(wm.state.StartTime),
		wm.state.Context["subdomains"],
		wm.state.Context["assets"],
		wm.state.Context["http_info"],
		wm.state.Context["ports"],
		wm.state.Context["vulnerabilities"],
		wm.state.Context["directories"],
		wm.state.Context["sql_injection"],
		wm.state.Context["command_injection"],
		wm.state.Context["privilege_escalation"],
		wm.state.Context["lateral_movement"],
		wm.state.Context["post_exploitation"],
	)

	wm.state.Results[PhaseReporting] = &model.ExecutionResult{
		Success:       true,
		ToolName:      "reporting",
		Output:        report,
		ExecutionTime: time.Since(wm.state.StartTime).Milliseconds(),
	}

	log.Println("报告生成完成")
	fmt.Println(report)

	return nil
}

func (wm *WorkflowManager) shouldSkipNextPhases() bool {
	// 根据当前结果判断是否需要跳过后续阶段
	// 例如，如果没有发现任何漏洞，可以跳过利用阶段
	vulnerabilities, ok := wm.state.Context["vulnerabilities"].(string)
	if ok && vulnerabilities == "" {
		return true
	}
	return false
}

func (wm *WorkflowManager) extractDomain(target string) string {
	// 从 URL 中提取域名
	// 简单实现，实际项目中可能需要更复杂的解析
	return target
}
