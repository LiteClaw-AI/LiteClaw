package team

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// CommunicationHub manages agent-to-agent communication
type CommunicationHub struct {
	agents    map[string]*AgentMailbox
	mu        sync.RWMutex
	broadcast chan *Message
}

// AgentMailbox represents an agent's message mailbox
type AgentMailbox struct {
	agentID string
	inbox   chan *Message
}

// Message represents a message between agents
type Message struct {
	ID        string                 `json:"id"`
	From      string                 `json:"from"`
	To        string                 `json:"to"`
	Content   string                 `json:"content"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp int64                  `json:"timestamp"`
	Type      MessageType            `json:"type"`
}

// MessageType represents message type
type MessageType string

const (
	MessageTypeDirect    MessageType = "direct"
	MessageTypeBroadcast MessageType = "broadcast"
	MessageTypeTask      MessageType = "task"
	MessageTypeResult    MessageType = "result"
	MessageTypeEvent     MessageType = "event"
)

// NewCommunicationHub creates a new communication hub
func NewCommunicationHub() *CommunicationHub {
	hub := &CommunicationHub{
		agents:    make(map[string]*AgentMailbox),
		broadcast: make(chan *Message, 1000),
	}
	
	// Start broadcast dispatcher
	go hub.dispatchBroadcast()
	
	return hub
}

// RegisterAgent registers an agent in the communication hub
func (h *CommunicationHub) RegisterAgent(agentID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.agents[agentID] = &AgentMailbox{
		agentID: agentID,
		inbox:   make(chan *Message, 100),
	}
}

// UnregisterAgent unregisters an agent from the communication hub
func (h *CommunicationHub) UnregisterAgent(agentID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if mailbox, exists := h.agents[agentID]; exists {
		close(mailbox.inbox)
		delete(h.agents, agentID)
	}
}

// Send sends a message from one agent to another
func (h *CommunicationHub) Send(from, to, content string, data map[string]interface{}) error {
	h.mu.RLock()
	mailbox, exists := h.agents[to]
	h.mu.RUnlock()

	if !exists {
		return fmt.Errorf("agent %s not found", to)
	}

	msg := &Message{
		ID:        generateMessageID(),
		From:      from,
		To:        to,
		Content:   content,
		Data:      data,
		Timestamp: time.Now().Unix(),
		Type:      MessageTypeDirect,
	}

	select {
	case mailbox.inbox <- msg:
		return nil
	default:
		return fmt.Errorf("agent %s inbox is full", to)
	}
}

// Broadcast broadcasts a message to all agents
func (h *CommunicationHub) Broadcast(from, content string, data map[string]interface{}) error {
	msg := &Message{
		ID:        generateMessageID(),
		From:      from,
		To:        "all",
		Content:   content,
		Data:      data,
		Timestamp: time.Now().Unix(),
		Type:      MessageTypeBroadcast,
	}

	select {
	case h.broadcast <- msg:
		return nil
	default:
		return fmt.Errorf("broadcast channel is full")
	}
}

// Receive receives a message for an agent (non-blocking)
func (h *CommunicationHub) Receive(agentID string) (*Message, error) {
	h.mu.RLock()
	mailbox, exists := h.agents[agentID]
	h.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("agent %s not found", agentID)
	}

	select {
	case msg := <-mailbox.inbox:
		return msg, nil
	default:
		return nil, nil // No message available
	}
}

// ReceiveBlocking receives a message for an agent (blocking with timeout)
func (h *CommunicationHub) ReceiveBlocking(agentID string, timeout time.Duration) (*Message, error) {
	h.mu.RLock()
	mailbox, exists := h.agents[agentID]
	h.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("agent %s not found", agentID)
	}

	select {
	case msg := <-mailbox.inbox:
		return msg, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout waiting for message")
	}
}

// dispatchBroadcast dispatches broadcast messages to all agents
func (h *CommunicationHub) dispatchBroadcast() {
	for msg := range h.broadcast {
		h.mu.RLock()
		for agentID, mailbox := range h.agents {
			if agentID != msg.From {
				select {
				case mailbox.inbox <- msg:
				default:
					// Skip if inbox is full
				}
			}
		}
		h.mu.RUnlock()
	}
}

// GetInboxSize returns the inbox size for an agent
func (h *CommunicationHub) GetInboxSize(agentID string) (int, error) {
	h.mu.RLock()
	mailbox, exists := h.agents[agentID]
	h.mu.RUnlock()

	if !exists {
		return 0, fmt.Errorf("agent %s not found", agentID)
	}

	return len(mailbox.inbox), nil
}

// generateMessageID generates a unique message ID
func generateMessageID() string {
	return fmt.Sprintf("msg-%d", time.Now().UnixNano())
}
