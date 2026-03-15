package team

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// TaskScheduler manages scheduled tasks
type TaskScheduler struct {
	cron     *cron.Cron
	jobs     map[string]cron.EntryID
	mu       sync.RWMutex
	handlers map[string]TaskHandler
}

// TaskHandler represents a task execution handler
type TaskHandler func(ctx context.Context)

// NewTaskScheduler creates a new task scheduler
func NewTaskScheduler() *TaskScheduler {
	return &TaskScheduler{
		cron:     cron.New(cron.WithSeconds()),
		jobs:     make(map[string]cron.EntryID),
		handlers: make(map[string]TaskHandler),
	}
}

// Schedule schedules a recurring task with cron expression
func (s *TaskScheduler) Schedule(ctx context.Context, agentID, task, cronExpr string, handler TaskHandler) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate job ID
	jobID := fmt.Sprintf("%s-%s", agentID, task)

	// Check if job already exists
	if _, exists := s.jobs[jobID]; exists {
		return fmt.Errorf("job %s already exists", jobID)
	}

	// Add cron job
	entryID, err := s.cron.AddFunc(cronExpr, func() {
		handler(ctx)
	})
	if err != nil {
		return fmt.Errorf("failed to schedule job: %w", err)
	}

	s.jobs[jobID] = entryID
	s.handlers[jobID] = handler

	return nil
}

// ScheduleOnce schedules a one-time task
func (s *TaskScheduler) ScheduleOnce(ctx context.Context, agentID, task string, executeAt time.Time, handler TaskHandler) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	jobID := fmt.Sprintf("%s-%s-once", agentID, task)

	// Calculate duration until execution
	duration := time.Until(executeAt)
	if duration < 0 {
		return fmt.Errorf("execution time is in the past")
	}

	// Schedule with timer
	time.AfterFunc(duration, func() {
		handler(ctx)
	})

	s.handlers[jobID] = handler

	return nil
}

// Unschedule removes a scheduled task
func (s *TaskScheduler) Unschedule(agentID, task string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	jobID := fmt.Sprintf("%s-%s", agentID, task)

	entryID, exists := s.jobs[jobID]
	if !exists {
		return fmt.Errorf("job %s not found", jobID)
	}

	s.cron.Remove(entryID)
	delete(s.jobs, jobID)
	delete(s.handlers, jobID)

	return nil
}

// Start starts the scheduler
func (s *TaskScheduler) Start() {
	s.cron.Start()
}

// Stop stops the scheduler
func (s *TaskScheduler) Stop() {
	s.cron.Stop()
}

// ListJobs lists all scheduled jobs
func (s *TaskScheduler) ListJobs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	jobs := make([]string, 0, len(s.jobs))
	for jobID := range s.jobs {
		jobs = append(jobs, jobID)
	}
	return jobs
}

// GetNextRun returns the next run time for a job
func (s *TaskScheduler) GetNextRun(agentID, task string) (time.Time, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	jobID := fmt.Sprintf("%s-%s", agentID, task)

	entryID, exists := s.jobs[jobID]
	if !exists {
		return time.Time{}, fmt.Errorf("job %s not found", jobID)
	}

	entry := s.cron.Entry(entryID)
	return entry.Next, nil
}

// ScheduleDaily schedules a daily task
func (s *TaskScheduler) ScheduleDaily(ctx context.Context, agentID, task string, hour, minute int, handler TaskHandler) error {
	cronExpr := fmt.Sprintf("0 %d %d * * *", minute, hour)
	return s.Schedule(ctx, agentID, task, cronExpr, handler)
}

// ScheduleHourly schedules an hourly task
func (s *TaskScheduler) ScheduleHourly(ctx context.Context, agentID, task string, minute int, handler TaskHandler) error {
	cronExpr := fmt.Sprintf("0 %d * * * *", minute)
	return s.Schedule(ctx, agentID, task, cronExpr, handler)
}

// ScheduleInterval schedules a task at regular intervals
func (s *TaskScheduler) ScheduleInterval(ctx context.Context, agentID, task string, interval time.Duration, handler TaskHandler) error {
	// Convert duration to cron expression (simplified)
	// For exact intervals, we'd need a ticker-based approach
	// This is a simplified implementation
	cronExpr := fmt.Sprintf("*/%d * * * * *", int(interval.Seconds()))
	return s.Schedule(ctx, agentID, task, cronExpr, handler)
}
