package agent

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
)

// BashTool implements a bash execution tool
type BashTool struct {
	name        string
	description string
	timeout     int
}

// NewBashTool creates a new bash tool
func NewBashTool() *BashTool {
	return &BashTool{
		name:        "bash",
		description: "Execute bash commands on the system. Use with caution.",
		timeout:     30,
	}
}

// Name returns the tool name
func (t *BashTool) Name() string {
	return t.name
}

// Description returns the tool description
func (t *BashTool) Description() string {
	return t.description
}

// Parameters returns the tool parameters schema
func (t *BashTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command": map[string]interface{}{
				"type":        "string",
				"description": "The bash command to execute",
			},
		},
		"required": []string{"command"},
	}
}

// Execute executes the tool
func (t *BashTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	command, ok := params["command"].(string)
	if !ok {
		return nil, fmt.Errorf("command parameter must be a string")
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/C", command)
	} else {
		cmd = exec.CommandContext(ctx, "bash", "-c", command)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return map[string]interface{}{
			"error":  err.Error(),
			"output": string(output),
		}, nil
	}

	return map[string]interface{}{
		"output": string(output),
	}, nil
}

// FileReadTool implements a file reading tool
type FileReadTool struct{}

// NewFileReadTool creates a new file read tool
func NewFileReadTool() *FileReadTool {
	return &FileReadTool{}
}

// Name returns the tool name
func (t *FileReadTool) Name() string {
	return "file_read"
}

// Description returns the tool description
func (t *FileReadTool) Description() string {
	return "Read content from a file"
}

// Parameters returns the tool parameters schema
func (t *FileReadTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file_path": map[string]interface{}{
				"type":        "string",
				"description": "The path to the file to read",
			},
		},
		"required": []string{"file_path"},
	}
}

// Execute executes the tool
func (t *FileReadTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	filePath, ok := params["file_path"].(string)
	if !ok {
		return nil, fmt.Errorf("file_path parameter must be a string")
	}

	data, err := exec.Command("cat", filePath).Output()
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return map[string]interface{}{
		"content": string(data),
		"path":    filePath,
	}, nil
}

// WebSearchTool implements a web search tool (mock)
type WebSearchTool struct{}

// NewWebSearchTool creates a new web search tool
func NewWebSearchTool() *WebSearchTool {
	return &WebSearchTool{}
}

// Name returns the tool name
func (t *WebSearchTool) Name() string {
	return "web_search"
}

// Description returns the tool description
func (t *WebSearchTool) Description() string {
	return "Search the web for information"
}

// Parameters returns the tool parameters schema
func (t *WebSearchTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "The search query",
			},
			"num_results": map[string]interface{}{
				"type":        "integer",
				"description": "Number of results to return",
				"default":     5,
			},
		},
		"required": []string{"query"},
	}
}

// Execute executes the tool
func (t *WebSearchTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	query, ok := params["query"].(string)
	if !ok {
		return nil, fmt.Errorf("query parameter must be a string")
	}

	// TODO: Implement actual web search
	// For now, return mock results
	return map[string]interface{}{
		"query": query,
		"results": []map[string]string{
			{
				"title":   "Example Result 1",
				"url":     "https://example.com/1",
				"snippet": "This is an example search result...",
			},
			{
				"title":   "Example Result 2",
				"url":     "https://example.com/2",
				"snippet": "Another example result...",
			},
		},
	}, nil
}

// CalculatorTool implements a calculator tool
type CalculatorTool struct{}

// NewCalculatorTool creates a new calculator tool
func NewCalculatorTool() *CalculatorTool {
	return &CalculatorTool{}
}

// Name returns the tool name
func (t *CalculatorTool) Name() string {
	return "calculator"
}

// Description returns the tool description
func (t *CalculatorTool) Description() string {
	return "Perform mathematical calculations"
}

// Parameters returns the tool parameters schema
func (t *CalculatorTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"expression": map[string]interface{}{
				"type":        "string",
				"description": "The mathematical expression to evaluate",
			},
		},
		"required": []string{"expression"},
	}
}

// Execute executes the tool
func (t *CalculatorTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	expr, ok := params["expression"].(string)
	if !ok {
		return nil, fmt.Errorf("expression parameter must be a string")
	}

	// Use bc command for calculation
	output, err := exec.Command("echo", expr, "|", "bc").Output()
	if err != nil {
		return nil, fmt.Errorf("calculation failed: %w", err)
	}

	return map[string]interface{}{
		"expression": expr,
		"result":     string(output),
	}, nil
}
