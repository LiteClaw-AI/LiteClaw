package provider

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
)

// Mistral provider implementation
type Mistral struct {
	*BaseProvider
	client *resty.Client
}

// NewMistral creates a new Mistral provider
func NewMistral(apiKey string) *Mistral {
	return &Mistral{
		BaseProvider: NewBaseProvider("mistral", apiKey, "https://api.mistral.ai/v1", "mistral-large-latest"),
		client: resty.New().
			SetHeader("Authorization", "Bearer "+apiKey).
			SetHeader("Content-Type", "application/json"),
	}
}

// Chat performs a chat completion
func (p *Mistral) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	if req.Model == "" {
		req.Model = p.model
	}

	resp, err := p.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&ChatResponse{}).
		Post(p.apiBase + "/chat/completions")

	if err != nil {
		return nil, fmt.Errorf("mistral chat request failed: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("mistral chat request failed: %s", resp.String())
	}

	return resp.Result().(*ChatResponse), nil
}

// ChatStream performs a streaming chat completion
func (p *Mistral) ChatStream(ctx context.Context, req *ChatRequest) (<-chan StreamChunk, error) {
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

// Embed generates embeddings
func (p *Mistral) Embed(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {
	if req.Model == "" {
		req.Model = "mistral-embed"
	}

	resp, err := p.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&EmbeddingResponse{}).
		Post(p.apiBase + "/embeddings")

	if err != nil {
		return nil, fmt.Errorf("mistral embedding request failed: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("mistral embedding request failed: %s", resp.String())
	}

	return resp.Result().(*EmbeddingResponse), nil
}

// Metadata returns provider metadata
func (p *Mistral) Metadata() Metadata {
	return Metadata{
		Name:         "Mistral AI",
		APIBase:      p.apiBase,
		DefaultModel: "mistral-large-latest",
		Capabilities: []Capability{CapChat, CapStream, CapEmbedding, CapTools, CapJSON},
		Region:       "international",
		Documentation: "https://docs.mistral.ai",
	}
}
