package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

// OpenRouter provider implementation (aggregator)
type OpenRouter struct {
	*BaseProvider
	client *resty.Client
}

// NewOpenRouter creates a new OpenRouter provider
func NewOpenRouter(apiKey string) *OpenRouter {
	return &OpenRouter{
		BaseProvider: NewBaseProvider("openrouter", apiKey, "https://openrouter.ai/api/v1", "anthropic/claude-3.5-sonnet"),
		client: resty.New().
			SetHeader("Authorization", "Bearer "+apiKey).
			SetHeader("Content-Type", "application/json").
			SetHeader("HTTP-Referer", "https://liteclaw.dev").
			SetHeader("X-Title", "LiteClaw"),
	}
}

// Chat performs a chat completion
func (p *OpenRouter) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	if req.Model == "" {
		req.Model = p.model
	}

	resp, err := p.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&ChatResponse{}).
		Post(p.apiBase + "/chat/completions")

	if err != nil {
		return nil, fmt.Errorf("openrouter chat request failed: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("openrouter chat request failed: %s", resp.String())
	}

	return resp.Result().(*ChatResponse), nil
}

// ChatStream performs a streaming chat completion
func (p *OpenRouter) ChatStream(ctx context.Context, req *ChatRequest) (<-chan StreamChunk, error) {
	if req.Model == "" {
		req.Model = p.model
	}
	req.Stream = true

	openai := &OpenAI{
		BaseProvider: p.BaseProvider,
		client:       p.client,
	}
	return openai.ChatStream(ctx, req)
}

// Embed generates embeddings (OpenRouter doesn't support embeddings)
func (p *OpenRouter) Embed(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {
	return nil, fmt.Errorf("openrouter does not support embeddings")
}

// Metadata returns provider metadata
func (p *OpenRouter) Metadata() Metadata {
	return Metadata{
		Name:         "OpenRouter",
		APIBase:      p.apiBase,
		DefaultModel: "anthropic/claude-3.5-sonnet",
		Capabilities: []Capability{CapChat, CapStream, CapVision, CapTools, CapJSON},
		Region:       "international",
		Documentation: "https://openrouter.ai/docs",
	}
}
