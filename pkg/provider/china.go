package provider

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
)

// AliyunQwen provider implementation (通义千问)
type AliyunQwen struct {
	*BaseProvider
	client *resty.Client
}

// NewAliyunQwen creates a new Aliyun Qwen provider
func NewAliyunQwen(apiKey string) *AliyunQwen {
	return &AliyunQwen{
		BaseProvider: NewBaseProvider("qwen", apiKey, "https://dashscope.aliyuncs.com/api/v1", "qwen-max"),
		client: resty.New().
			SetHeader("Authorization", "Bearer "+apiKey).
			SetHeader("Content-Type", "application/json"),
	}
}

// Chat performs a chat completion
func (p *AliyunQwen) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	if req.Model == "" {
		req.Model = p.model
	}

	// Convert to Qwen format
	qwenReq := map[string]interface{}{
		"model": req.Model,
		"input": map[string]interface{}{
			"messages": req.Messages,
		},
		"parameters": map[string]interface{}{
			"temperature": req.Temperature,
			"max_tokens":  req.MaxTokens,
		},
	}

	resp, err := p.client.R().
		SetContext(ctx).
		SetBody(qwenReq).
		Post(p.apiBase + "/services/aigc/text-generation/generation")

	if err != nil {
		return nil, fmt.Errorf("qwen request failed: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("qwen request failed: %s", resp.String())
	}

	// TODO: Parse response
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

// ChatStream performs streaming chat
func (p *AliyunQwen) ChatStream(ctx context.Context, req *ChatRequest) (<-chan StreamChunk, error) {
	stream := make(chan StreamChunk)
	close(stream)
	return stream, fmt.Errorf("streaming not yet implemented for Qwen")
}

// Embed generates embeddings
func (p *AliyunQwen) Embed(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {
	if req.Model == "" {
		req.Model = "text-embedding-v2"
	}

	qwenReq := map[string]interface{}{
		"model": req.Model,
		"input": map[string]interface{}{
			"texts": []string{req.Input},
		},
	}

	resp, err := p.client.R().
		SetContext(ctx).
		SetBody(qwenReq).
		Post(p.apiBase + "/services/embeddings/text-embedding/text-embedding")

	if err != nil {
		return nil, fmt.Errorf("qwen embedding failed: %w", err)
	}

	// TODO: Parse response
	return &EmbeddingResponse{
		Model: req.Model,
	}, nil
}

// Metadata returns provider metadata
func (p *AliyunQwen) Metadata() Metadata {
	return Metadata{
		Name:         "Aliyun Qwen",
		APIBase:      p.apiBase,
		DefaultModel: "qwen-max",
		Capabilities: []Capability{
			CapChat, CapStream, CapEmbedding, CapVision, CapTools,
		},
		Region:        "china",
		Documentation: "https://help.aliyun.com/document_detail/2712195.html",
	}
}

// Moonshot provider implementation (月之暗面 Kimi)
type Moonshot struct {
	*BaseProvider
	client *resty.Client
}

// NewMoonshot creates a new Moonshot provider
func NewMoonshot(apiKey string) *Moonshot {
	return &Moonshot{
		BaseProvider: NewBaseProvider("moonshot", apiKey, "https://api.moonshot.cn/v1", "moonshot-v1-8k"),
		client: resty.New().
			SetHeader("Authorization", "Bearer "+apiKey).
			SetHeader("Content-Type", "application/json"),
	}
}

// Chat performs a chat completion
func (p *Moonshot) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	if req.Model == "" {
		req.Model = p.model
	}

	resp, err := p.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&ChatResponse{}).
		Post(p.apiBase + "/chat/completions")

	if err != nil {
		return nil, fmt.Errorf("moonshot request failed: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("moonshot request failed: %s", resp.String())
	}

	return resp.Result().(*ChatResponse), nil
}

// ChatStream performs streaming chat
func (p *Moonshot) ChatStream(ctx context.Context, req *ChatRequest) (<-chan StreamChunk, error) {
	if req.Model == "" {
		req.Model = p.model
	}
	req.Stream = true

	// Moonshot uses OpenAI-compatible streaming
	openai := &OpenAI{
		BaseProvider: p.BaseProvider,
		client:       p.client,
	}
	return openai.ChatStream(ctx, req)
}

// Embed generates embeddings
func (p *Moonshot) Embed(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {
	if req.Model == "" {
		req.Model = "moonshot-v1-8k"
	}

	resp, err := p.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&EmbeddingResponse{}).
		Post(p.apiBase + "/embeddings")

	if err != nil {
		return nil, fmt.Errorf("moonshot embedding failed: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("moonshot embedding failed: %s", resp.String())
	}

	return resp.Result().(*EmbeddingResponse), nil
}

// Metadata returns provider metadata
func (p *Moonshot) Metadata() Metadata {
	return Metadata{
		Name:         "Moonshot (Kimi)",
		APIBase:      p.apiBase,
		DefaultModel: "moonshot-v1-8k",
		Capabilities: []Capability{
			CapChat, CapStream, CapEmbedding, CapJSON,
		},
		Region:        "china",
		Documentation: "https://platform.moonshot.cn/docs",
	}
}

