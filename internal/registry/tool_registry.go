package registry

import (
	"nezha_sec/internal/executor"
)

// ToolRegistry 工具注册表
// 管理所有可用的工具执行器
type ToolRegistry struct {
	executors map[string]executor.ToolExecutorInterface
}

// NewToolRegistry 创建一个新的工具注册表
// 返回值列表：
//   - *ToolRegistry 工具注册表实例
//   - error 初始化过程中可能发生的错误
func NewToolRegistry() (*ToolRegistry, error) {
	registry := &ToolRegistry{
		executors: make(map[string]executor.ToolExecutorInterface),
	}

	// 注册默认工具
	if err := registry.registerDefaultTools(); err != nil {
		return nil, err
	}

	return registry, nil
}

// registerDefaultTools 注册默认工具
// 返回值列表：
//   - error 注册过程中可能发生的错误
func (r *ToolRegistry) registerDefaultTools() error {
	// 注册curl工具
	curlExecutor, err := executor.NewCurlExecutor()
	if err != nil {
		return err
	}
	r.RegisterTool(curlExecutor)

	// 注册nmap工具
	nmapExecutor, err := executor.NewNmapExecutor()
	if err != nil {
		return err
	}
	r.RegisterTool(nmapExecutor)

	// 注册sqlmap工具
	sqlmapExecutor, err := executor.NewSqlmapExecutor()
	if err != nil {
		return err
	}
	r.RegisterTool(sqlmapExecutor)

	// 注册QuestExploit工具
	questExploitExecutor, err := executor.NewQuestExploitExecutor()
	if err != nil {
		return err
	}
	r.RegisterTool(questExploitExecutor)

	// 注册Subfinder工具
	subfinderExecutor, err := executor.NewSubfinderExecutor()
	if err != nil {
		return err
	}
	r.RegisterTool(subfinderExecutor)

	// 注册Amass工具
	amassExecutor, err := executor.NewAmassExecutor()
	if err != nil {
		return err
	}
	r.RegisterTool(amassExecutor)

	// 注册Httpx工具
	httpxExecutor, err := executor.NewHttpxExecutor()
	if err != nil {
		return err
	}
	r.RegisterTool(httpxExecutor)

	// 注册Naabu工具
	naabuExecutor, err := executor.NewNaabuExecutor()
	if err != nil {
		return err
	}
	r.RegisterTool(naabuExecutor)

	// 注册Nuclei工具
	nucleiExecutor, err := executor.NewNucleiExecutor()
	if err != nil {
		return err
	}
	r.RegisterTool(nucleiExecutor)

	// 注册Garak工具
	garakExecutor, err := executor.NewGarakExecutor()
	if err != nil {
		return err
	}
	r.RegisterTool(garakExecutor)

	// 注册Trivy工具
	trivyExecutor, err := executor.NewTrivyExecutor()
	if err != nil {
		return err
	}
	r.RegisterTool(trivyExecutor)

	// 注册Msfconsole工具
	msfconsoleExecutor, err := executor.NewMsfconsoleExecutor()
	if err != nil {
		return err
	}
	r.RegisterTool(msfconsoleExecutor)

	// 注册Commix工具
	commixExecutor, err := executor.NewCommixExecutor()
	if err != nil {
		return err
	}
	r.RegisterTool(commixExecutor)

	// 注册Sliver-cli工具
	sliverCliExecutor, err := executor.NewSliverCliExecutor()
	if err != nil {
		return err
	}
	r.RegisterTool(sliverCliExecutor)

	// 注册Havoc工具
	havocExecutor, err := executor.NewHavocExecutor()
	if err != nil {
		return err
	}
	r.RegisterTool(havocExecutor)

	// 注册Impacket工具
	impacketExecutor, err := executor.NewImpacketExecutor()
	if err != nil {
		return err
	}
	r.RegisterTool(impacketExecutor)

	// 注册Responder工具
	responderExecutor, err := executor.NewResponderExecutor()
	if err != nil {
		return err
	}
	r.RegisterTool(responderExecutor)

	// 注册CrackMapExec工具
	crackMapExecExecutor, err := executor.NewCrackMapExecExecutor()
	if err != nil {
		return err
	}
	r.RegisterTool(crackMapExecExecutor)

	// 注册Chisel工具
	chiselExecutor, err := executor.NewChiselExecutor()
	if err != nil {
		return err
	}
	r.RegisterTool(chiselExecutor)

	// 注册Ffuf工具
	ffufExecutor, err := executor.NewFfufExecutor()
	if err != nil {
		return err
	}
	r.RegisterTool(ffufExecutor)
	
	return nil
}

// RegisterTool 注册工具执行器
// 参数列表：
//   - tool 工具执行器实例
func (r *ToolRegistry) RegisterTool(tool executor.ToolExecutorInterface) {
	r.executors[tool.GetToolName()] = tool
}

// GetTool 获取工具执行器
// 参数列表：
//   - toolName 工具名称
// 返回值列表：
//   - executor.ToolExecutorInterface 工具执行器实例
//   - bool 是否找到工具
func (r *ToolRegistry) GetTool(toolName string) (executor.ToolExecutorInterface, bool) {
	tool, exists := r.executors[toolName]
	return tool, exists
}

// ListTools 列出所有可用工具
// 返回值列表：
//   - []string 工具名称列表
func (r *ToolRegistry) ListTools() []string {
	tools := make([]string, 0, len(r.executors))
	for toolName := range r.executors {
		tools = append(tools, toolName)
	}
	return tools
}
