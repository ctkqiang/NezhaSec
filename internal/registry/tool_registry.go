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
