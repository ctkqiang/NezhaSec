package registry

import (
	"nezha_sec/internal/executor"
)

type ToolRegistry struct {
	executors map[string]executor.ToolExecutorInterface
}

func NewToolRegistry() (*ToolRegistry, error) {
	registry := &ToolRegistry{
		executors: make(map[string]executor.ToolExecutorInterface),
	}

	if err := registry.registerDefaultTools(); err != nil {
		return nil, err
	}

	return registry, nil
}

func (r *ToolRegistry) registerDefaultTools() error {

	curlExecutor, err := executor.NewCurlExecutor()
	if err != nil {
		return err
	}
	r.RegisterTool(curlExecutor)

	nmapExecutor, err := executor.NewNmapExecutor()
	if err != nil {
		return err
	}
	r.RegisterTool(nmapExecutor)

	sqlmapExecutor, err := executor.NewSqlmapExecutor()
	if err != nil {
		return err
	}
	r.RegisterTool(sqlmapExecutor)

	questExploitExecutor, err := executor.NewQuestExploitExecutor()
	if err != nil {
		return err
	}
	r.RegisterTool(questExploitExecutor)

	subfinderExecutor, err := executor.NewSubfinderExecutor()
	if err != nil {
		return err
	}
	r.RegisterTool(subfinderExecutor)

	amassExecutor, err := executor.NewAmassExecutor()
	if err != nil {
		return err
	}
	r.RegisterTool(amassExecutor)

	httpxExecutor, err := executor.NewHttpxExecutor()
	if err != nil {
		return err
	}
	r.RegisterTool(httpxExecutor)

	naabuExecutor, err := executor.NewNaabuExecutor()
	if err != nil {
		return err
	}
	r.RegisterTool(naabuExecutor)

	nucleiExecutor, err := executor.NewNucleiExecutor()
	if err != nil {
		return err
	}
	r.RegisterTool(nucleiExecutor)

	garakExecutor, err := executor.NewGarakExecutor()
	if err != nil {
		return err
	}
	r.RegisterTool(garakExecutor)

	trivyExecutor, err := executor.NewTrivyExecutor()
	if err != nil {
		return err
	}
	r.RegisterTool(trivyExecutor)

	msfconsoleExecutor, err := executor.NewMsfconsoleExecutor()
	if err != nil {
		return err
	}
	r.RegisterTool(msfconsoleExecutor)

	commixExecutor, err := executor.NewCommixExecutor()
	if err != nil {
		return err
	}
	r.RegisterTool(commixExecutor)

	sliverCliExecutor, err := executor.NewSliverCliExecutor()
	if err != nil {
		return err
	}
	r.RegisterTool(sliverCliExecutor)

	havocExecutor, err := executor.NewHavocExecutor()
	if err != nil {
		return err
	}
	r.RegisterTool(havocExecutor)

	impacketExecutor, err := executor.NewImpacketExecutor()
	if err != nil {
		return err
	}
	r.RegisterTool(impacketExecutor)

	responderExecutor, err := executor.NewResponderExecutor()
	if err != nil {
		return err
	}
	r.RegisterTool(responderExecutor)

	crackMapExecExecutor, err := executor.NewCrackMapExecExecutor()
	if err != nil {
		return err
	}
	r.RegisterTool(crackMapExecExecutor)

	chiselExecutor, err := executor.NewChiselExecutor()
	if err != nil {
		return err
	}
	r.RegisterTool(chiselExecutor)

	ffufExecutor, err := executor.NewFfufExecutor()
	if err != nil {
		return err
	}
	r.RegisterTool(ffufExecutor)

	return nil
}

func (r *ToolRegistry) RegisterTool(tool executor.ToolExecutorInterface) {
	r.executors[tool.GetToolName()] = tool
}

func (r *ToolRegistry) GetTool(toolName string) (executor.ToolExecutorInterface, bool) {
	tool, exists := r.executors[toolName]
	return tool, exists
}

func (r *ToolRegistry) ListTools() []string {
	tools := make([]string, 0, len(r.executors))
	for toolName := range r.executors {
		tools = append(tools, toolName)
	}
	return tools
}
