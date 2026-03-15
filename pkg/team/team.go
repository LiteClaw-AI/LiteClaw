package team

import (
	"context"
	"fmt"
	"sync"
	"time"

	"liteclaw/pkg/agent"
	"liteclaw/pkg/provider"
)

// Team represents a team of AI employees (agents)
type Team struct {
	id          string
	name        string
	description string
	agents      map[string]*agent.Agent
	communication *CommunicationHub
	workflow    *WorkflowEngine
	scheduler   *TaskScheduler
	registry    *provider.Registry
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

// TeamConfig represents team configuration
type TeamConfig struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MaxAgents   int    `json:"max_agents"`
}

// NewTeam creates a new AI employee team
func NewTeam(config TeamConfig, registry *provider.Registry) *Team {
	ctx, cancel := context.WithCancel(context.Background())
	return &Team{
		id:          config.ID,
		name:        config.Name,
		description: config.Description,
		agents:      make(map[string]*agent.Agent),
		communication: NewCommunicationHub(),
		workflow:    NewWorkflowEngine(),
		scheduler:   NewTaskScheduler(),
		registry:    registry,
		ctx:         ctx,
		cancel:      cancel,
	}
}

// AddAgent adds an agent to the team
func (t *Team) AddAgent(a *agent.Agent) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, exists := t.agents[a.ID()]; exists {
		return fmt.Errorf("agent %s already exists in team", a.ID())
	}

	t.agents[a.ID()] = a
	t.communication.RegisterAgent(a.ID())
	return nil
}

// RemoveAgent removes an agent from the team
func (t *Team) RemoveAgent(agentID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.agents, agentID)
	t.communication.UnregisterAgent(agentID)
}

// GetAgent retrieves an agent by ID
func (t *Team) GetAgent(agentID string) (*agent.Agent, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	a, exists := t.agents[agentID]
	if !exists {
		return nil, fmt.Errorf("agent %s not found", agentID)
	}
	return a, nil
}

// ListAgents lists all agents in the team
func (t *Team) ListAgents() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	agentIDs := make([]string, 0, len(t.agents))
	for id := range t.agents {
		agentIDs = append(agentIDs, id)
	}
	return agentIDs
}

// Getters
func (t *Team) GetID() string               { return t.id }
func (t *Team) GetName() string             { return t.name }
func (t *Team) GetDescription() string      { return t.description }
func (t *Team) GetWorkflowEngine() *WorkflowEngine { return t.workflow }
func (t *Team) GetScheduler() *TaskScheduler { return t.scheduler }
func (t *Team) GetCommunicationHub() *CommunicationHub { return t.communication }

// ExecuteTask executes a task with specified agent
func (t *Team) ExecuteTask(ctx context.Context, agentID string, task string) (*TaskResult, error) {
	a, err := t.GetAgent(agentID)
	if err != nil {
		return nil, err
	}

	executor := agent.NewExecutor(t.registry)
	result, err := executor.Execute(ctx, a, task)
	if err != nil {
		return nil, err
	}

	return &TaskResult{
		AgentID:  agentID,
		Task:     task,
		Output:   result.Output,
		Steps:    result.Steps,
		Duration: result.Duration,
	}, nil
}

// ExecuteWorkflow executes a multi-agent workflow
func (t *Team) ExecuteWorkflow(ctx context.Context, workflowID string, input map[string]interface{}) (*WorkflowResult, error) {
	return t.workflow.Execute(ctx, t, workflowID, input)
}

// SendMessage sends a message from one agent to another
func (t *Team) SendMessage(from, to string, message string, data map[string]interface{}) error {
	return t.communication.Send(from, to, message, data)
}

// Broadcast broadcasts a message to all agents
func (t *Team) Broadcast(from string, message string, data map[string]interface{}) error {
	return t.communication.Broadcast(from, message, data)
}

// ScheduleTask schedules a recurring task
func (t *Team) ScheduleTask(agentID string, task string, cronExpr string) error {
	return t.scheduler.Schedule(t.ctx, agentID, task, cronExpr, func(ctx context.Context) {
		t.ExecuteTask(ctx, agentID, task)
	})
}

// Stop stops the team
func (t *Team) Stop() {
	t.cancel()
	t.scheduler.Stop()
}

// ID returns the team ID
func (t *Team) ID() string {
	return t.id
}

// Name returns the team name
func (t *Team) Name() string {
	return t.name
}

// TaskResult represents a task execution result
type TaskResult struct {
	AgentID  string      `json:"agent_id"`
	Task     string      `json:"task"`
	Output   string      `json:"output"`
	Steps    interface{} `json:"steps"`
	Duration time.Duration `json:"duration"`
	Error    string      `json:"error,omitempty"`
}

// WorkflowResult represents a workflow execution result
type WorkflowResult struct {
	WorkflowID string                 `json:"workflow_id"`
	Status     string                 `json:"status"`
	Steps      []WorkflowStepResult   `json:"steps"`
	Output     map[string]interface{} `json:"output"`
	Duration   time.Duration          `json:"duration"`
	Error      string                 `json:"error,omitempty"`
}

// WorkflowStepResult represents a single workflow step result
type WorkflowStepResult struct {
	StepID   string    `json:"step_id"`
	AgentID  string    `json:"agent_id"`
	Task     string    `json:"task"`
	Output   string    `json:"output"`
	Duration time.Duration `json:"duration"`
	Error    string    `json:"error,omitempty"`
}
