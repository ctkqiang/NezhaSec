package model

// ExecutionResult 工具执行结果
// 包含执行状态、输出和错误信息
type ExecutionResult struct {
	// 执行状态
	Success bool
	
	// 工具名称
	ToolName string
	
	// 执行输出
	Output string
	
	// 执行错误信息
	Error string
	
	// 执行时间（毫秒）
	ExecutionTime int64
	
	// 结构化结果（如果有）
	StructuredOutput map[string]interface{}
}
