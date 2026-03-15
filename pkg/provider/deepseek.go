package provider

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
)

// DeepSeek provider implementation
type DeepSeek struct {
	*BaseProvider
	client *resty.Client
}

// NewDeepSeek creates a new DeepSeek provider
func NewDeepSeek(apiKey string) *DeepSeek {
	return &DeepSeek{
		BaseProvider: NewBaseProvider("deepseek", apiKey, "https://api.deepseek.com/v1", "deepseek-chat"),
		client: resty.New().
			SetHeader("Authorization", "Bearer "+apiKey).
			SetHeader("Content-Type", "application/json"),
	}
}

// Chat performs a chat completion
func (p *DeepSeek) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	if req.Model == "" {
		req.Model = p.model
	}

	resp, err := p.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&ChatResponse{}).
		Post(p.apiBase + "/chat/completions")

	if err != nil {
		return nil, fmt.Errorf("deepseek chat request failed: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("deepseek chat request failed: %s", resp.String())
	}

	return resp.Result().(*ChatResponse), nil
}

// ChatStream performs a streaming chat completion
func (p *DeepSeek) ChatStream(ctx context.Context, req *ChatRequest) (<-chan StreamChunk, error) {
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

// Embed generates embeddings (DeepSeek doesn't support embeddings)
func (p *DeepSeek) Embed(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {
	return nil, fmt.Errorf("deepseek does not support embeddings")
}

// Metadata returns provider metadata
func (p *DeepSeek) Metadata() Metadata {
	return Metadata{
		Name:         "DeepSeek",
		APIBase:      p.apiBase,
		DefaultModel: "deepseek-chat",
		Capabilities: []Capability{CapChat, CapStream, CapTools, CapJSON},
		Region:       "china",
		Documentation: "https://platform.deepseek.com/docs",
	}
}
