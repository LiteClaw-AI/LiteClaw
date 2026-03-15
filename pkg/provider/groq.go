package provider

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
)

// Groq provider implementation (ultra-fast inference)
type Groq struct {
	*BaseProvider
	client *resty.Client
}

// NewGroq creates a new Groq provider
func NewGroq(apiKey string) *Groq {
	return &Groq{
		BaseProvider: NewBaseProvider("groq", apiKey, "https://api.groq.com/openai/v1", "llama-3.3-70b-versatile"),
		client: resty.New().
			SetHeader("Authorization", "Bearer "+apiKey).
			SetHeader("Content-Type", "application/json"),
	}
}

// Chat performs a chat completion
func (p *Groq) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	if req.Model == "" {
		req.Model = p.model
	}

	resp, err := p.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&ChatResponse{}).
		Post(p.apiBase + "/chat/completions")

	if err != nil {
		return nil, fmt.Errorf("groq chat request failed: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("groq chat request failed: %s", resp.String())
	}

	return resp.Result().(*ChatResponse), nil
}

// ChatStream performs a streaming chat completion
func (p *Groq) ChatStream(ctx context.Context, req *ChatRequest) (<-chan StreamChunk, error) {
	// Reuse OpenAI streaming implementation (Groq is OpenAI-compatible)
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

// Embed generates embeddings (Groq doesn't support embeddings natively)
func (p *Groq) Embed(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {
	return nil, fmt.Errorf("groq does not support embeddings")
}

// Metadata returns provider metadata
func (p *Groq) Metadata() Metadata {
	return Metadata{
		Name:         "Groq",
		APIBase:      p.apiBase,
		DefaultModel: "llama-3.3-70b-versatile",
		Capabilities: []Capability{CapChat, CapStream, CapTools, CapJSON},
		Region:       "international",
		Documentation: "https://console.groq.com/docs",
	}
}
