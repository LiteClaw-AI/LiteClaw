package team

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// WorkflowEngine manages multi-agent workflows
type WorkflowEngine struct {
	workflows map[string]*WorkflowDefinition
	running   map[string]*WorkflowInstance
	mu        sync.RWMutex
}

// WorkflowDefinition represents a workflow definition
type WorkflowDefinition struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Steps       []WorkflowStep `json:"steps"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
}

// WorkflowStep represents a single workflow step
type WorkflowStep struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	AgentID     string                 `json:"agent_id"`
	Task        string                 `json:"task"`
	Input       map[string]interface{} `json:"input,omitempty"`
	Condition   string                 `json:"condition,omitempty"`
	Parallel    bool                   `json:"parallel,omitempty"`
	DependsOn   []string               `json:"depends_on,omitempty"`
	OnError     string                 `json:"on_error,omitempty"` // "continue", "abort", "retry"
	MaxRetries  int                    `json:"max_retries,omitempty"`
}

// WorkflowInstance represents a running workflow instance
type WorkflowInstance struct {
	ID           string
	DefinitionID string
	Status       string
	CurrentStep  int
	Variables    map[string]interface{}
	Results      map[string]interface{}
	StartTime    time.Time
	EndTime      *time.Time
	Error        string
}

// NewWorkflowEngine creates a new workflow engine
func NewWorkflowEngine() *WorkflowEngine {
	return &WorkflowEngine{
		workflows: make(map[string]*WorkflowDefinition),
		running:   make(map[string]*WorkflowInstance),
	}
}

// RegisterWorkflow registers a workflow definition
func (e *WorkflowEngine) RegisterWorkflow(def *WorkflowDefinition) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.workflows[def.ID]; exists {
		return fmt.Errorf("workflow %s already exists", def.ID)
	}

	e.workflows[def.ID] = def
	return nil
}

// Execute executes a workflow
func (e *WorkflowEngine) Execute(ctx context.Context, team *Team, workflowID string, input map[string]interface{}) (*WorkflowResult, error) {
	// Get workflow definition
	e.mu.RLock()
	def, exists := e.workflows[workflowID]
	e.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("workflow %s not found", workflowID)
	}

	// Create instance
	instance := &WorkflowInstance{
		ID:           fmt.Sprintf("instance-%d", time.Now().UnixNano()),
		DefinitionID: workflowID,
		Status:       "running",
		Variables:    mergeMaps(def.Variables, input),
		Results:      make(map[string]interface{}),
		StartTime:    time.Now(),
	}

	// Register running instance
	e.mu.Lock()
	e.running[instance.ID] = instance
	e.mu.Unlock()

	// Execute workflow
	result := &WorkflowResult{
		WorkflowID: workflowID,
		Status:     "success",
		Steps:      []WorkflowStepResult{},
	}

	startTime := time.Now()

	// Execute steps
	for _, step := range def.Steps {
		// Check context cancellation
		select {
		case <-ctx.Done():
			instance.Status = "cancelled"
			result.Status = "cancelled"
			result.Error = "workflow cancelled"
			break
		default:
		}

		// Check dependencies
		if !e.checkDependencies(step, result) {
			continue
		}

		// Check condition
		if step.Condition != "" && !e.evaluateCondition(step.Condition, instance.Variables) {
			continue
		}

		// Execute step
		stepResult := e.executeStep(ctx, team, step, instance)
		result.Steps = append(result.Steps, stepResult)

		// Handle errors
		if stepResult.Error != "" {
			if step.OnError == "abort" {
				instance.Status = "failed"
				result.Status = "failed"
				result.Error = stepResult.Error
				break
			}
		}

		// Update variables
		instance.Results[step.ID] = stepResult.Output
	}

	instance.EndTime = ptrTime(time.Now())
	instance.Status = result.Status
	result.Duration = time.Since(startTime)
	result.Output = instance.Results

	return result, nil
}

// executeStep executes a single workflow step
func (e *WorkflowEngine) executeStep(ctx context.Context, team *Team, step WorkflowStep, instance *WorkflowInstance) WorkflowStepResult {
	result := WorkflowStepResult{
		StepID:  step.ID,
		AgentID: step.AgentID,
		Task:    step.Task,
	}

	startTime := time.Now()

	// Replace variables in task
	task := e.replaceVariables(step.Task, instance.Variables, instance.Results)

	// Execute task
	taskResult, err := team.ExecuteTask(ctx, step.AgentID, task)
	if err != nil {
		result.Error = err.Error()
		result.Duration = time.Since(startTime)
		return result
	}

	result.Output = taskResult.Output
	result.Duration = time.Since(startTime)
	return result
}

// checkDependencies checks if all dependencies are satisfied
func (e *WorkflowEngine) checkDependencies(step WorkflowStep, result *WorkflowResult) bool {
	if len(step.DependsOn) == 0 {
		return true
	}

	completedSteps := make(map[string]bool)
	for _, stepResult := range result.Steps {
		completedSteps[stepResult.StepID] = true
	}

	for _, dep := range step.DependsOn {
		if !completedSteps[dep] {
			return false
		}
	}

	return true
}

// evaluateCondition evaluates a simple condition
func (e *WorkflowEngine) evaluateCondition(condition string, variables map[string]interface{}) bool {
	// Simple condition evaluation (can be extended)
	// Examples: "${success} == true", "${count} > 5"
	// For now, just return true
	return true
}

// replaceVariables replaces variables in a string
func (e *WorkflowEngine) replaceVariables(template string, variables, results map[string]interface{}) string {
	// Simple variable replacement
	// Examples: "${variables.topic}", "${results.step1.output}"
	// For now, just return the template as-is
	return template
}

// Helper functions

func mergeMaps(m1, m2 map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m1 {
		result[k] = v
	}
	for k, v := range m2 {
		result[k] = v
	}
	return result
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
