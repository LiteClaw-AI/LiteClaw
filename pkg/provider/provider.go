// Package provider defines the LLM provider interface and implementations
package provider

import (
	"context"
	"io"
)

// MessageRole represents the role of a message
type MessageRole string

const (
	RoleSystem    MessageRole = "system"
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
	RoleFunction  MessageRole = "function"
)

// Message represents a chat message
type Message struct {
	Role    MessageRole `json:"role"`
	Content string      `json:"content"`
	Name    string      `json:"name,omitempty"`
}

// ChatRequest represents a chat completion request
type ChatRequest struct {
	Messages    []Message `json:"messages"`
	Model       string    `json:"model,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
	Tools       []Tool    `json:"tools,omitempty"`
}

// ChatResponse represents a chat completion response
type ChatResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice represents a completion choice
type Choice struct {
	Index        int      `json:"index"`
	Message      Message  `json:"message"`
	FinishReason string   `json:"finish_reason"`
	Delta        *Message `json:"delta,omitempty"` // For streaming
}

// Usage represents token usage
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Tool represents a function tool
type Tool struct {
	Type     string      `json:"type"`
	Function ToolFunction `json:"function"`
}

// ToolFunction represents a function definition
type ToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// EmbeddingRequest represents an embedding request
type EmbeddingRequest struct {
	Input string `json:"input"`
	Model string `json:"model,omitempty"`
}

// EmbeddingResponse represents an embedding response
type EmbeddingResponse struct {
	Object string          `json:"object"`
	Data   []EmbeddingData `json:"data"`
	Model  string          `json:"model"`
	Usage  Usage           `json:"usage"`
}

// EmbeddingData represents embedding data
type EmbeddingData struct {
	Object    string    `json:"object"`
	Index     int       `json:"index"`
	Embedding []float32 `json:"embedding"`
}

// StreamChunk represents a streaming response chunk
type StreamChunk struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
}

// Capability represents a provider capability
type Capability string

const (
	CapChat       Capability = "chat"
	CapStream     Capability = "stream"
	CapEmbedding  Capability = "embedding"
	CapVision     Capability = "vision"
	CapAudio      Capability = "audio"
	CapTools      Capability = "tools"
	CapJSON       Capability = "json"
)

// Metadata represents provider metadata
type Metadata struct {
	Name           string       `json:"name"`
	APIBase        string       `json:"api_base"`
	DefaultModel   string       `json:"default_model"`
	Capabilities   []Capability `json:"capabilities"`
	Region         string       `json:"region"`
	Documentation  string       `json:"documentation"`
}

// Provider is the interface for LLM providers
type Provider interface {
	// Chat performs a chat completion
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)

	// ChatStream performs a streaming chat completion
	ChatStream(ctx context.Context, req *ChatRequest) (<-chan StreamChunk, error)

	// Embed generates embeddings
	Embed(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error)

	// Metadata returns provider metadata
	Metadata() Metadata

	// Name returns the provider name
	Name() string
}

// StreamProvider extends Provider with streaming support
type StreamProvider interface {
	Provider
	ChatStream(ctx context.Context, req *ChatRequest) (<-chan StreamChunk, error)
}

// VisionProvider extends Provider with vision support
type VisionProvider interface {
	Provider
	ChatWithImage(ctx context.Context, req *ChatRequest, imageURL string) (*ChatResponse, error)
}

// BaseProvider provides common functionality for providers
type BaseProvider struct {
	name    string
	apiKey  string
	apiBase string
	model   string
}

// NewBaseProvider creates a new base provider
func NewBaseProvider(name, apiKey, apiBase, model string) *BaseProvider {
	return &BaseProvider{
		name:    name,
		apiKey:  apiKey,
		apiBase: apiBase,
		model:   model,
	}
}

// Name returns the provider name
func (p *BaseProvider) Name() string {
	return p.name
}

// Model returns the default model
func (p *BaseProvider) Model() string {
	return p.model
}
