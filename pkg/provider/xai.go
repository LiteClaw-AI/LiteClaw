package provider

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
)

// XAI provider implementation (xAI Grok)
type XAI struct {
	*BaseProvider
	client *resty.Client
}

// NewXAI creates a new xAI provider
func NewXAI(apiKey string) *XAI {
	return &XAI{
		BaseProvider: NewBaseProvider("xai", apiKey, "https://api.x.ai/v1", "grok-2-latest"),
		client: resty.New().
			SetHeader("Authorization", "Bearer "+apiKey).
			SetHeader("Content-Type", "application/json"),
	}
}

// Chat performs a chat completion
func (p *XAI) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	if req.Model == "" {
		req.Model = p.model
	}

	resp, err := p.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&ChatResponse{}).
		Post(p.apiBase + "/chat/completions")

	if err != nil {
		return nil, fmt.Errorf("xai chat request failed: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("xai chat request failed: %s", resp.String())
	}

	return resp.Result().(*ChatResponse), nil
}

// ChatStream performs a streaming chat completion
func (p *XAI) ChatStream(ctx context.Context, req *ChatRequest) (<-chan StreamChunk, error) {
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

// Embed generates embeddings (xAI doesn't support embeddings yet)
func (p *XAI) Embed(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {
	return nil, fmt.Errorf("xai does not support embeddings")
}

// Metadata returns provider metadata
func (p *XAI) Metadata() Metadata {
	return Metadata{
		Name:         "xAI (Grok)",
		APIBase:      p.apiBase,
		DefaultModel: "grok-2-latest",
		Capabilities: []Capability{CapChat, CapStream, CapVision, CapTools, CapJSON},
		Region:       "international",
		Documentation: "https://docs.x.ai",
	}
}
