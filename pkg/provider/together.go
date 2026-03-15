package provider

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
)

// Together provider implementation
type Together struct {
	*BaseProvider
	client *resty.Client
}

// NewTogether creates a new Together provider
func NewTogether(apiKey string) *Together {
	return &Together{
		BaseProvider: NewBaseProvider("together", apiKey, "https://api.together.xyz/v1", "meta-llama/Llama-3.3-70B-Instruct-Turbo"),
		client: resty.New().
			SetHeader("Authorization", "Bearer "+apiKey).
			SetHeader("Content-Type", "application/json"),
	}
}

// Chat performs a chat completion
func (p *Together) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	if req.Model == "" {
		req.Model = p.model
	}

	resp, err := p.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&ChatResponse{}).
		Post(p.apiBase + "/chat/completions")

	if err != nil {
		return nil, fmt.Errorf("together chat request failed: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("together chat request failed: %s", resp.String())
	}

	return resp.Result().(*ChatResponse), nil
}

// ChatStream performs a streaming chat completion
func (p *Together) ChatStream(ctx context.Context, req *ChatRequest) (<-chan StreamChunk, error) {
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
func (p *Together) Embed(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {
	if req.Model == "" {
		req.Model = "togethercomputer/m2-bert-80M-8k-retrieval"
	}

	resp, err := p.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&EmbeddingResponse{}).
		Post(p.apiBase + "/embeddings")

	if err != nil {
		return nil, fmt.Errorf("together embedding request failed: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("together embedding request failed: %s", resp.String())
	}

	return resp.Result().(*EmbeddingResponse), nil
}

// Metadata returns provider metadata
func (p *Together) Metadata() Metadata {
	return Metadata{
		Name:         "Together AI",
		APIBase:      p.apiBase,
		DefaultModel: "meta-llama/Llama-3.3-70B-Instruct-Turbo",
		Capabilities: []Capability{CapChat, CapStream, CapEmbedding, CapJSON},
		Region:       "international",
		Documentation: "https://docs.together.ai",
	}
}
