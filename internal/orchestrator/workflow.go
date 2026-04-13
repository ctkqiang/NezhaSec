package orchestrator

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"nezha_sec/internal/model"
	"strings"
	"time"
)

type WorkflowPhase string

const (
	PhaseReconnaissance      WorkflowPhase = "侦察阶段"
	PhaseVulnerabilityScan   WorkflowPhase = "漏洞扫描阶段"
	PhaseExploitation        WorkflowPhase = "利用阶段"
	PhasePrivilegeEscalation WorkflowPhase = "权限提升阶段"
	PhaseLateralMovement     WorkflowPhase = "横向移动阶段"
	PhasePostExploitation    WorkflowPhase = "后利用阶段"
	PhaseReporting           WorkflowPhase = "报告阶段"
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
	messageChan  chan interface{}
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

func (wm *WorkflowManager) SetMessageChannel(ch chan interface{}) {
	wm.messageChan = ch
}

func (wm *WorkflowManager) sendPhaseChange(phase WorkflowPhase) {
	if wm.messageChan != nil {
		wm.messageChan <- struct {
			Phase string
		}{Phase: string(phase)}
	}
}

func (wm *WorkflowManager) sendLog(logMessage string) {
	if wm.messageChan != nil {
		wm.messageChan <- struct {
			Log string
		}{Log: logMessage}
	}
}

func (wm *WorkflowManager) sendSuccess(message string) {
	if wm.messageChan != nil {
		wm.messageChan <- struct {
			Message string
		}{Message: message}
	}
}

func (wm *WorkflowManager) sendError(err error) {
	if wm.messageChan != nil {
		wm.messageChan <- struct {
			Error string
		}{Error: err.Error()}
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
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		wm.state.CurrentPhase = phase
		wm.sendPhaseChange(phase)
		log.Printf("开始阶段: %s", phase)

		err := wm.executePhase(ctx, phase)
		if err != nil {
			log.Printf("阶段执行失败 %s: %v", phase, err)
			wm.sendError(fmt.Errorf("阶段 %s 执行失败: %w", phase, err))
			continue
		}

		if wm.shouldSkipNextPhases() {
			log.Println("根据当前结果，跳过后续阶段")
			wm.sendLog("根据当前结果，跳过后续阶段")
			break
		}

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
	wm.sendLog("执行侦察阶段...")

	subfinderArgs := map[string]interface{}{
		"domain": wm.extractDomain(wm.state.Target),
	}
	subfinderResult, err := wm.orchestrator.ExecuteTool("subfinder", subfinderArgs)
	if err != nil {
		log.Printf("subfinder 执行失败: %v", err)
		wm.sendLog(fmt.Sprintf("subfinder 执行失败: %v", err))
	} else {
		wm.state.Results[PhaseReconnaissance] = subfinderResult
		wm.state.Context["subdomains"] = subfinderResult.Output
		wm.sendSuccess("子域名发现完成")
	}

	amassArgs := map[string]interface{}{
		"domain":  wm.extractDomain(wm.state.Target),
		"passive": true,
	}
	amassResult, err := wm.orchestrator.ExecuteTool("amass", amassArgs)
	if err != nil {
		log.Printf("amass 执行失败: %v", err)
		wm.sendLog(fmt.Sprintf("amass 执行失败: %v", err))
	} else {
		wm.state.Context["assets"] = amassResult.Output
		wm.sendSuccess("资产发现完成")
	}

	httpxArgs := map[string]interface{}{
		"targets":     wm.state.Target,
		"probe":       true,
		"status_code": true,
		"tech_detect": true,
	}
	httpxResult, err := wm.orchestrator.ExecuteTool("httpx", httpxArgs)
	if err != nil {
		log.Printf("httpx 执行失败: %v", err)
		wm.sendLog(fmt.Sprintf("httpx 执行失败: %v", err))
	} else {
		wm.state.Context["http_info"] = httpxResult.Output
		wm.sendSuccess("HTTP 探测完成")
	}

	nmapArgs := map[string]interface{}{
		"target":    wm.extractDomain(wm.state.Target),
		"ports":     "1-1000",
		"scan_type": "service",
	}
	nmapResult, err := wm.orchestrator.ExecuteTool("nmap", nmapArgs)
	if err != nil {
		log.Printf("nmap 执行失败: %v", err)
		wm.sendLog(fmt.Sprintf("nmap 执行失败: %v", err))
	} else {
		wm.state.Context["ports"] = nmapResult.Output
		wm.sendSuccess("端口扫描完成")
	}

	return nil
}

func (wm *WorkflowManager) executeVulnerabilityScan(ctx context.Context) error {
	log.Println("执行漏洞扫描阶段...")
	wm.sendLog("执行漏洞扫描阶段...")

	nucleiArgs := map[string]interface{}{
		"target":    wm.state.Target,
		"templates": "cves,exposures,misconfigurations",
	}
	nucleiResult, err := wm.orchestrator.ExecuteTool("nuclei", nucleiArgs)
	if err != nil {
		log.Printf("nuclei 执行失败: %v", err)
		wm.sendLog(fmt.Sprintf("nuclei 执行失败: %v", err))
	} else {
		wm.state.Results[PhaseVulnerabilityScan] = nucleiResult
		wm.state.Context["vulnerabilities"] = nucleiResult.Output
		wm.sendSuccess("漏洞扫描完成")
	}

	ffufArgs := map[string]interface{}{
		"url":        wm.state.Target + "/FUZZ",
		"wordlist":   "common.txt",
		"extensions": "php,html,js,json,txt",
	}
	ffufResult, err := wm.orchestrator.ExecuteTool("ffuf", ffufArgs)
	if err != nil {
		log.Printf("ffuf 执行失败: %v", err)
		wm.sendLog(fmt.Sprintf("ffuf 执行失败: %v", err))
	} else {
		wm.state.Context["directories"] = ffufResult.Output
		wm.sendSuccess("目录爆破完成")
	}

	sqlmapArgs := map[string]interface{}{
		"url":   wm.buildTargetURL(wm.state.Target),
		"crawl": 2,
		"risk":  1,
		"level": 1,
	}
	sqlmapResult, err := wm.orchestrator.ExecuteTool("sqlmap", sqlmapArgs)
	if err != nil {
		log.Printf("sqlmap 执行失败: %v", err)
		wm.sendLog(fmt.Sprintf("sqlmap 执行失败: %v", err))
	} else {
		wm.state.Context["sql_injection"] = sqlmapResult.Output
		wm.sendSuccess("SQL 注入检测完成")
	}

	return nil
}

func (wm *WorkflowManager) executeExploitation(ctx context.Context) error {
	log.Println("执行利用阶段...")
	wm.sendLog("执行利用阶段...")

	vulnerabilities, ok := wm.state.Context["vulnerabilities"].(string)
	if ok && vulnerabilities != "" {
		log.Println("基于漏洞扫描结果执行利用...")
		wm.sendLog("基于漏洞扫描结果执行利用...")
	}

	commixArgs := map[string]interface{}{
		"url":   wm.buildTargetURL(wm.state.Target),
		"level": 1,
	}
	commixResult, err := wm.orchestrator.ExecuteTool("commix", commixArgs)
	if err != nil {
		log.Printf("commix 执行失败: %v", err)
		wm.sendLog(fmt.Sprintf("commix 执行失败: %v", err))
	} else {
		wm.state.Results[PhaseExploitation] = commixResult
		wm.state.Context["command_injection"] = commixResult.Output
		wm.sendSuccess("命令注入检测完成")
	}

	return nil
}

func (wm *WorkflowManager) executePrivilegeEscalation(ctx context.Context) error {
	log.Println("执行权限提升阶段...")
	wm.sendLog("执行权限提升阶段...")

	msfArgs := map[string]interface{}{
		"command": "search exploit/unix",
	}
	msfResult, err := wm.orchestrator.ExecuteTool("msfconsole", msfArgs)
	if err != nil {
		log.Printf("msfconsole 执行失败: %v", err)
		wm.sendLog(fmt.Sprintf("msfconsole 执行失败: %v", err))
	} else {
		wm.state.Results[PhasePrivilegeEscalation] = msfResult
		wm.state.Context["privilege_escalation"] = msfResult.Output
		wm.sendSuccess("权限提升模块搜索完成")
	}

	return nil
}

func (wm *WorkflowManager) executeLateralMovement(ctx context.Context) error {
	log.Println("执行横向移动阶段...")
	wm.sendLog("执行横向移动阶段...")

	impacketArgs := map[string]interface{}{
		"command": "smbclient -L //127.0.0.1",
	}
	impacketResult, err := wm.orchestrator.ExecuteTool("impacket", impacketArgs)
	if err != nil {
		log.Printf("impacket 执行失败: %v", err)
		wm.sendLog(fmt.Sprintf("impacket 执行失败: %v", err))
	} else {
		wm.state.Results[PhaseLateralMovement] = impacketResult
		wm.state.Context["lateral_movement"] = impacketResult.Output
		wm.sendSuccess("横向移动探测完成")
	}

	return nil
}

func (wm *WorkflowManager) executePostExploitation(ctx context.Context) error {
	log.Println("执行后利用阶段...")
	wm.sendLog("执行后利用阶段...")

	sliverArgs := map[string]interface{}{
		"command": "sessions",
	}
	sliverResult, err := wm.orchestrator.ExecuteTool("sliver-cli", sliverArgs)
	if err != nil {
		log.Printf("sliver-cli 执行失败: %v", err)
		wm.sendLog(fmt.Sprintf("sliver-cli 执行失败: %v", err))
	} else {
		wm.state.Results[PhasePostExploitation] = sliverResult
		wm.state.Context["post_exploitation"] = sliverResult.Output
		wm.sendSuccess("后利用模块检查完成")
	}

	return nil
}

func (wm *WorkflowManager) executeReporting(ctx context.Context) error {
	log.Println("执行报告阶段...")
	wm.sendLog("执行报告阶段...")

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
	wm.sendSuccess("报告生成完成")
	fmt.Println(report)

	return nil
}

func (wm *WorkflowManager) shouldSkipNextPhases() bool {
	vulnerabilities, ok := wm.state.Context["vulnerabilities"].(string)
	if ok && vulnerabilities == "" {
		return true
	}
	return false
}

func (wm *WorkflowManager) extractDomain(target string) string {
	return target
}

// buildTargetURL 构建完整的目标URL
// 如果目标已经是完整URL则直接返回，否则默认添加 http:// 前缀
func (wm *WorkflowManager) buildTargetURL(target string) string {
	if target == "" {
		return ""
	}

	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		return target
	}

	parsedURL, err := url.Parse("http://" + target)
	if err != nil {
		return "http://" + target
	}

	return parsedURL.String()
}
