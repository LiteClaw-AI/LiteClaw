package agent

import (
	"context"
	"sync"
	"time"
)

// InMemoryStore implements in-memory storage for agent memory
type InMemoryStore struct {
	messages []Message
	mu       sync.RWMutex
	maxSize  int
}

// NewInMemoryStore creates a new in-memory store
func NewInMemoryStore(maxSize ...int) *InMemoryStore {
	size := 1000
	if len(maxSize) > 0 {
		size = maxSize[0]
	}
	return &InMemoryStore{
		messages: make([]Message, 0),
		maxSize:  size,
	}
}

// Add adds a message to memory
func (s *InMemoryStore) Add(ctx context.Context, message Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	message.Timestamp = time.Now().Unix()
	s.messages = append(s.messages, message)

	// Trim if exceeds max size
	if len(s.messages) > s.maxSize {
		s.messages = s.messages[len(s.messages)-s.maxSize:]
	}

	return nil
}

// Get retrieves messages from memory
func (s *InMemoryStore) Get(ctx context.Context, limit int) ([]Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 || limit > len(s.messages) {
		limit = len(s.messages)
	}

	// Return last N messages
	start := len(s.messages) - limit
	if start < 0 {
		start = 0
	}

	result := make([]Message, limit)
	copy(result, s.messages[start:])
	return result, nil
}

// Clear clears the memory
func (s *InMemoryStore) Clear(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = make([]Message, 0)
	return nil
}

// Search searches in memory (simple substring search)
func (s *InMemoryStore) Search(ctx context.Context, query string, limit int) ([]Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []Message
	for i := len(s.messages) - 1; i >= 0 && len(results) < limit; i-- {
		if contains(s.messages[i].Content, query) {
			results = append(results, s.messages[i])
		}
	}

	return results, nil
}

// contains checks if s contains substr (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
