package provider

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
)

// Anthropic provider implementation
type Anthropic struct {
	*BaseProvider
	client *resty.Client
}

// NewAnthropic creates a new Anthropic provider
func NewAnthropic(apiKey string) *Anthropic {
	return &Anthropic{
		BaseProvider: NewBaseProvider("anthropic", apiKey, "https://api.anthropic.com/v1", "claude-3-5-sonnet-20241022"),
		client: resty.New().
			SetHeader("x-api-key", apiKey).
			SetHeader("anthropic-version", "2023-06-01").
			SetHeader("Content-Type", "application/json"),
	}
}

// Chat performs a chat completion
func (p *Anthropic) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	if req.Model == "" {
		req.Model = p.model
	}

	// Convert to Anthropic format
	anthropicReq := map[string]interface{}{
		"model":      req.Model,
		"max_tokens": req.MaxTokens,
		"messages":   req.Messages,
	}

	if req.Temperature > 0 {
		anthropicReq["temperature"] = req.Temperature
	}

	if req.System != "" {
		anthropicReq["system"] = req.System
	}

	resp, err := p.client.R().
		SetContext(ctx).
		SetBody(anthropicReq).
		Post(p.apiBase + "/messages")

	if err != nil {
		return nil, fmt.Errorf("anthropic request failed: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("anthropic request failed: %s", resp.String())
	}

	// Parse Anthropic response and convert to standard format
	// TODO: Implement proper response parsing
	return &ChatResponse{
		Model: req.Model,
		Choices: []Choice{
			{
				Message: Message{
					Role:    RoleAssistant,
					Content: string(resp.Body()),
				},
			},
		},
	}, nil
}

// ChatStream performs a streaming chat completion
func (p *Anthropic) ChatStream(ctx context.Context, req *ChatRequest) (<-chan StreamChunk, error) {
	// TODO: Implement streaming
	stream := make(chan StreamChunk)
	close(stream)
	return stream, fmt.Errorf("streaming not yet implemented for Anthropic")
}

// Embed generates embeddings (not supported by Anthropic)
func (p *Anthropic) Embed(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {
	return nil, fmt.Errorf("Anthropic does not support embeddings")
}

// Metadata returns provider metadata
func (p *Anthropic) Metadata() Metadata {
	return Metadata{
		Name:         "Anthropic",
		APIBase:      p.apiBase,
		DefaultModel: "claude-3-5-sonnet-20241022",
		Capabilities: []Capability{
			CapChat, CapStream, CapVision, CapTools,
		},
		Region:        "international",
		Documentation: "https://docs.anthropic.com",
	}
}

// System prompt helper
func (req *ChatRequest) GetSystemPrompt() string {
	for _, msg := range req.Messages {
		if msg.Role == RoleSystem {
			return msg.Content
		}
	}
	return ""
}
