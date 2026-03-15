package agent

import (
	"context"
	"fmt"
	"sync"
)

// Agent represents an AI agent with tools and memory
type Agent struct {
	id          string
	name        string
	description string
	provider    string
	model       string
	tools       map[string]Tool
	memory      Memory
	config      AgentConfig
	mu          sync.RWMutex
}

// AgentConfig represents agent configuration
type AgentConfig struct {
	MaxIterations  int     `json:"max_iterations"`
	Temperature    float64 `json:"temperature"`
	SystemPrompt   string  `json:"system_prompt"`
	EnableMemory   bool    `json:"enable_memory"`
	EnableThinking bool    `json:"enable_thinking"`
	Verbose        bool    `json:"verbose"`
}

// Tool represents an agent tool
type Tool interface {
	// Name returns the tool name
	Name() string

	// Description returns the tool description
	Description() string

	// Parameters returns the tool parameters schema
	Parameters() map[string]interface{}

	// Execute executes the tool
	Execute(ctx context.Context, params map[string]interface{}) (interface{}, error)
}

// Memory represents agent memory
type Memory interface {
	// Add adds a message to memory
	Add(ctx context.Context, message Message) error

	// Get retrieves messages from memory
	Get(ctx context.Context, limit int) ([]Message, error)

	// Clear clears the memory
	Clear(ctx context.Context) error

	// Search searches in memory
	Search(ctx context.Context, query string, limit int) ([]Message, error)
}

// Message represents an agent message
type Message struct {
	Role       string                 `json:"role"`
	Content    string                 `json:"content"`
	ToolCalls  []ToolCall             `json:"tool_calls,omitempty"`
	ToolCallID string                 `json:"tool_call_id,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Timestamp  int64                  `json:"timestamp"`
}

// ToolCall represents a tool call
type ToolCall struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Params   map[string]interface{} `json:"params"`
	Result   interface{}            `json:"result,omitempty"`
	Error    string                 `json:"error,omitempty"`
	Duration int64                  `json:"duration,omitempty"`
}

// Thought represents agent thinking process
type Thought struct {
	Step      int      `json:"step"`
	Reasoning string   `json:"reasoning"`
	Action    string   `json:"action,omitempty"`
	Input     string   `json:"input,omitempty"`
	Output    string   `json:"output,omitempty"`
	Error     string   `json:"error,omitempty"`
	Complete  bool     `json:"complete"`
	Tags      []string `json:"tags,omitempty"`
}

// NewAgent creates a new agent
func NewAgent(id, name, description string, opts ...AgentOption) *Agent {
	agent := &Agent{
		id:          id,
		name:        name,
		description: description,
		provider:    "openai",
		model:       "gpt-4",
		tools:       make(map[string]Tool),
		config: AgentConfig{
			MaxIterations:  10,
			Temperature:    0.7,
			EnableMemory:   true,
			EnableThinking: false,
			Verbose:        false,
		},
	}

	for _, opt := range opts {
		opt(agent)
	}

	if agent.config.EnableMemory && agent.memory == nil {
		agent.memory = NewInMemoryStore()
	}

	return agent
}

// AgentOption is a function that configures an agent
type AgentOption func(*Agent)

// WithProvider sets the provider
func WithProvider(provider string) AgentOption {
	return func(a *Agent) {
		a.provider = provider
	}
}

// WithModel sets the model
func WithModel(model string) AgentOption {
	return func(a *Agent) {
		a.model = model
	}
}

// WithTools sets the tools
func WithTools(tools ...Tool) AgentOption {
	return func(a *Agent) {
		for _, tool := range tools {
			a.tools[tool.Name()] = tool
		}
	}
}

// WithMemory sets the memory
func WithMemory(memory Memory) AgentOption {
	return func(a *Agent) {
		a.memory = memory
	}
}

// WithConfig sets the configuration
func WithConfig(config AgentConfig) AgentOption {
	return func(a *Agent) {
		a.config = config
	}
}

// AddTool adds a tool to the agent
func (a *Agent) AddTool(tool Tool) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if _, exists := a.tools[tool.Name()]; exists {
		return fmt.Errorf("tool %s already exists", tool.Name())
	}

	a.tools[tool.Name()] = tool
	return nil
}

// RemoveTool removes a tool from the agent
func (a *Agent) RemoveTool(name string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.tools, name)
}

// GetTools returns all tools
func (a *Agent) GetTools() []Tool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	tools := make([]Tool, 0, len(a.tools))
	for _, tool := range a.tools {
		tools = append(tools, tool)
	}
	return tools
}

// ID returns the agent ID
func (a *Agent) ID() string {
	return a.id
}

// Name returns the agent name
func (a *Agent) Name() string {
	return a.name
}

// Description returns the agent description
func (a *Agent) Description() string {
	return a.description
}
