package model

type ExecutionResult struct {
	Success bool

	ToolName string

	Output string

	Error string

	ExecutionTime int64

	StructuredOutput map[string]interface{}
}
