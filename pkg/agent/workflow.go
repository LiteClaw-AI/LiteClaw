package agent

import (
	"context"
	"fmt"
)

// Chain represents a chain of agents
type Chain struct {
	agents []*Agent
	name   string
}

// NewChain creates a new agent chain
func NewChain(name string, agents ...*Agent) *Chain {
	return &Chain{
		agents: agents,
		name:   name,
	}
}

// Execute executes the chain
func (c *Chain) Execute(ctx context.Context, input string) (*ChainResult, error) {
	result := &ChainResult{
		ChainName: c.name,
		Input:     input,
		Steps:     []ChainStep{},
	}

	currentInput := input
	for i, agent := range c.agents {
		// Create executor for each agent
		// TODO: Get registry from context or pass as parameter
		executor := NewExecutor(nil)

		execResult, err := executor.Execute(ctx, agent, currentInput)
		if err != nil {
			return nil, fmt.Errorf("agent %s failed at step %d: %w", agent.Name(), i+1, err)
		}

		step := ChainStep{
			AgentID:   agent.ID(),
			AgentName: agent.Name(),
			Input:     currentInput,
			Output:    execResult.Output,
		}

		result.Steps = append(result.Steps, step)
		currentInput = execResult.Output
	}

	result.Output = currentInput
	return result, nil
}

// ChainResult represents chain execution result
type ChainResult struct {
	ChainName string      `json:"chain_name"`
	Input     string      `json:"input"`
	Output    string      `json:"output"`
	Steps     []ChainStep `json:"steps"`
}

// ChainStep represents a single chain step
type ChainStep struct {
	AgentID   string `json:"agent_id"`
	AgentName string `json:"agent_name"`
	Input     string `json:"input"`
	Output    string `json:"output"`
}

// Router routes input to different agents based on conditions
type Router struct {
	agents   map[string]*Agent
	decision func(input string) string
	name     string
}

// NewRouter creates a new agent router
func NewRouter(name string, decision func(input string) string) *Router {
	return &Router{
		agents:   make(map[string]*Agent),
		decision: decision,
		name:     name,
	}
}

// AddAgent adds an agent to the router
func (r *Router) AddAgent(key string, agent *Agent) {
	r.agents[key] = agent
}

// Execute executes the router
func (r *Router) Execute(ctx context.Context, input string) (*RouterResult, error) {
	// Decide which agent to use
	key := r.decision(input)
	agent, exists := r.agents[key]
	if !exists {
		return nil, fmt.Errorf("no agent found for key: %s", key)
	}

	// Create executor
	executor := NewExecutor(nil)

	// Execute agent
	execResult, err := executor.Execute(ctx, agent, input)
	if err != nil {
		return nil, fmt.Errorf("agent execution failed: %w", err)
	}

	return &RouterResult{
		RouterName: r.name,
		Selected:   key,
		AgentID:    agent.ID(),
		Input:      input,
		Output:     execResult.Output,
	}, nil
}

// RouterResult represents router execution result
type RouterResult struct {
	RouterName string `json:"router_name"`
	Selected   string `json:"selected"`
	AgentID    string `json:"agent_id"`
	Input      string `json:"input"`
	Output     string `json:"output"`
}

// Parallel executes multiple agents in parallel
type Parallel struct {
	agents []*Agent
	name   string
}

// NewParallel creates a new parallel executor
func NewParallel(name string, agents ...*Agent) *Parallel {
	return &Parallel{
		agents: agents,
		name:   name,
	}
}

// Execute executes agents in parallel
func (p *Parallel) Execute(ctx context.Context, input string) (*ParallelResult, error) {
	result := &ParallelResult{
		Name:  p.name,
		Input: input,
		Results: []ParallelStep{},
	}

	// TODO: Implement actual parallel execution with goroutines
	// For now, execute sequentially
	for _, agent := range p.agents {
		executor := NewExecutor(nil)
		execResult, err := executor.Execute(ctx, agent, input)
		if err != nil {
			return nil, fmt.Errorf("agent %s failed: %w", agent.Name(), err)
		}

		result.Results = append(result.Results, ParallelStep{
			AgentID:   agent.ID(),
			AgentName: agent.Name(),
			Output:    execResult.Output,
		})
	}

	return result, nil
}

// ParallelResult represents parallel execution result
type ParallelResult struct {
	Name    string        `json:"name"`
	Input   string        `json:"input"`
	Results []ParallelStep `json:"results"`
}

// ParallelStep represents a single parallel step
type ParallelStep struct {
	AgentID   string `json:"agent_id"`
	AgentName string `json:"agent_name"`
	Output    string `json:"output"`
}
