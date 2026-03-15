package agent

import (
	"context"
	"fmt"
	"time"

	"liteclaw/pkg/provider"
)

// Executor executes agent tasks
type Executor struct {
	registry *provider.Registry
	verbose  bool
}

// NewExecutor creates a new executor
func NewExecutor(registry *provider.Registry) *Executor {
	return &Executor{
		registry: registry,
		verbose:  false,
	}
}

// Execute executes an agent task
func (e *Executor) Execute(ctx context.Context, agent *Agent, input string) (*ExecutionResult, error) {
	startTime := time.Now()
	result := &ExecutionResult{
		AgentID:   agent.ID(),
		Input:     input,
		Steps:     []ExecutionStep{},
		StartTime: startTime,
	}

	// Get provider
	prov, err := e.registry.Get(agent.provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	// Build messages
	messages := []provider.Message{
		{Role: provider.RoleSystem, Content: e.buildSystemPrompt(agent)},
		{Role: provider.RoleUser, Content: input},
	}

	// Add memory if enabled
	if agent.config.EnableMemory && agent.memory != nil {
		memMessages, err := agent.memory.Get(ctx, 10)
		if err == nil && len(memMessages) > 0 {
			// Insert memory messages before user input
			memProviderMessages := make([]provider.Message, len(memMessages))
			for i, msg := range memMessages {
				memProviderMessages[i] = provider.Message{
					Role:    provider.MessageRole(msg.Role),
					Content: msg.Content,
				}
			}
			messages = append([]provider.Message{messages[0]}, append(memProviderMessages, messages[1:]...)...)
		}
	}

	// Execute iterations
	for i := 0; i < agent.config.MaxIterations; i++ {
		step := ExecutionStep{
			Iteration: i + 1,
		}

		// Call LLM
		resp, err := prov.Chat(ctx, &provider.ChatRequest{
			Messages:    messages,
			Model:       agent.model,
			Temperature: agent.config.Temperature,
			Tools:       e.buildToolsSchema(agent),
		})
		if err != nil {
			step.Error = err.Error()
			result.Steps = append(result.Steps, step)
			result.Error = err.Error()
			break
		}

		// Extract assistant message
		assistantMsg := resp.Choices[0].Message
		step.Response = assistantMsg.Content
		messages = append(messages, assistantMsg)

		// Check if tool calls are needed
		if len(resp.Choices[0].Message.ToolCalls) > 0 {
			// Convert to our ToolCall format
			toolCalls := make([]ToolCall, len(resp.Choices[0].Message.ToolCalls))
			for i, tc := range resp.Choices[0].Message.ToolCalls {
				toolCalls[i] = ToolCall{
					ID:     tc.ID,
					Name:   tc.Function.Name,
					Params: tc.Function.Parameters,
				}
			}

			// Execute tools
			toolResults, err := e.executeTools(ctx, agent, toolCalls)
			if err != nil {
				step.Error = err.Error()
				result.Steps = append(result.Steps, step)
				continue
			}

			step.ToolCalls = toolCalls
			step.ToolResults = toolResults

			// Add tool results to messages
			for _, res := range toolResults {
				messages = append(messages, provider.Message{
					Role:       provider.RoleFunction,
					Content:    fmt.Sprintf("%v", res.Result),
					Name:       res.Name,
				})
			}
		} else {
			// No tool calls, we have a final answer
			result.Output = assistantMsg.Content
			result.Steps = append(result.Steps, step)
			break
		}

		result.Steps = append(result.Steps, step)
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(startTime)

	// Save to memory
	if agent.config.EnableMemory && agent.memory != nil {
		agent.memory.Add(ctx, Message{
			Role:      "user",
			Content:   input,
			Timestamp: startTime.Unix(),
		})
		agent.memory.Add(ctx, Message{
			Role:      "assistant",
			Content:   result.Output,
			Timestamp: result.EndTime.Unix(),
		})
	}

	return result, nil
}

// ExecutionResult represents execution result
type ExecutionResult struct {
	AgentID   string          `json:"agent_id"`
	Input     string          `json:"input"`
	Output    string          `json:"output"`
	Steps     []ExecutionStep `json:"steps"`
	Error     string          `json:"error,omitempty"`
	StartTime time.Time       `json:"start_time"`
	EndTime   time.Time       `json:"end_time"`
	Duration  time.Duration   `json:"duration"`
}

// ExecutionStep represents a single execution step
type ExecutionStep struct {
	Iteration   int              `json:"iteration"`
	Response    string           `json:"response,omitempty"`
	ToolCalls   []ToolCall       `json:"tool_calls,omitempty"`
	ToolResults []ToolCallResult `json:"tool_results,omitempty"`
	Error       string           `json:"error,omitempty"`
}

// ToolCallResult represents a tool call result
type ToolCallResult struct {
	ID     string      `json:"id"`
	Name   string      `json:"name"`
	Result interface{} `json:"result"`
	Error  string      `json:"error,omitempty"`
}

// buildSystemPrompt builds system prompt for the agent
func (e *Executor) buildSystemPrompt(agent *Agent) string {
	prompt := fmt.Sprintf("You are %s. %s\n\n", agent.name, agent.description)

	if len(agent.tools) > 0 {
		prompt += "You have access to the following tools:\n"
		for _, tool := range agent.GetTools() {
			prompt += fmt.Sprintf("- %s: %s\n", tool.Name(), tool.Description())
		}
		prompt += "\nUse tools when necessary to complete the task."
	}

	if agent.config.SystemPrompt != "" {
		prompt += "\n\n" + agent.config.SystemPrompt
	}

	return prompt
}

// buildToolsSchema builds tools schema for the provider
func (e *Executor) buildToolsSchema(agent *Agent) []provider.Tool {
	tools := make([]provider.Tool, 0, len(agent.tools))
	for _, tool := range agent.GetTools() {
		tools = append(tools, provider.Tool{
			Type: "function",
			Function: provider.ToolFunction{
				Name:        tool.Name(),
				Description: tool.Description(),
				Parameters:  tool.Parameters(),
			},
		})
	}
	return tools
}

// executeTools executes tool calls
func (e *Executor) executeTools(ctx context.Context, agent *Agent, toolCalls []ToolCall) ([]ToolCallResult, error) {
	results := make([]ToolCallResult, 0, len(toolCalls))

	for _, tc := range toolCalls {
		result := ToolCallResult{
			ID:   tc.ID,
			Name: tc.Name,
		}

		tool, exists := agent.tools[tc.Name]
		if !exists {
			result.Error = fmt.Sprintf("tool %s not found", tc.Name)
			results = append(results, result)
			continue
		}

		// Execute tool
		output, err := tool.Execute(ctx, tc.Params)
		if err != nil {
			result.Error = err.Error()
		} else {
			result.Result = output
		}

		results = append(results, result)
	}

	return results, nil
}
