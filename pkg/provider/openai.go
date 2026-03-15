package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/go-resty/resty/v2"
)

// OpenAI provider implementation
type OpenAI struct {
	*BaseProvider
	client *resty.Client
}

// NewOpenAI creates a new OpenAI provider
func NewOpenAI(apiKey string) *OpenAI {
	return &OpenAI{
		BaseProvider: NewBaseProvider("openai", apiKey, "https://api.openai.com/v1", "gpt-4"),
		client: resty.New().
			SetHeader("Authorization", "Bearer "+apiKey).
			SetHeader("Content-Type", "application/json"),
	}
}

// Chat performs a chat completion
func (p *OpenAI) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	if req.Model == "" {
		req.Model = p.model
	}

	resp, err := p.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&ChatResponse{}).
		Post(p.apiBase + "/chat/completions")

	if err != nil {
		return nil, fmt.Errorf("openai chat request failed: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("openai chat request failed: %s", resp.String())
	}

	return resp.Result().(*ChatResponse), nil
}

// ChatStream performs a streaming chat completion
func (p *OpenAI) ChatStream(ctx context.Context, req *ChatRequest) (<-chan StreamChunk, error) {
	if req.Model == "" {
		req.Model = p.model
	}
	req.Stream = true

	stream := make(chan StreamChunk, 100)

	go func() {
		defer close(stream)

		resp, err := p.client.R().
			SetContext(ctx).
			SetBody(req).
			SetDoNotParseResponse(true).
			Post(p.apiBase + "/chat/completions")

		if err != nil {
			return
		}
		defer resp.RawBody().Close()

		decoder := json.NewDecoder(resp.RawBody())
		for {
			var chunk struct {
				ID      string   `json:"id"`
				Object  string   `json:"object"`
				Created int64    `json:"created"`
				Model   string   `json:"model"`
				Choices []Choice `json:"choices"`
			}

			// Read "data: " prefix
			var line string
			if err := decoder.Decode(&line); err != nil {
				if err == io.EOF {
					break
				}
				continue
			}

			if line == "[DONE]" {
				break
			}

			if len(line) > 6 && line[:6] == "data: " {
				if err := json.Unmarshal([]byte(line[6:]), &chunk); err != nil {
					continue
				}
				stream <- StreamChunk{
					ID:      chunk.ID,
					Object:  chunk.Object,
					Created: chunk.Created,
					Model:   chunk.Model,
					Choices: chunk.Choices,
				}
			}
		}
	}()

	return stream, nil
}

// Embed generates embeddings
func (p *OpenAI) Embed(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {
	if req.Model == "" {
		req.Model = "text-embedding-3-small"
	}

	resp, err := p.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&EmbeddingResponse{}).
		Post(p.apiBase + "/embeddings")

	if err != nil {
		return nil, fmt.Errorf("openai embedding request failed: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("openai embedding request failed: %s", resp.String())
	}

	return resp.Result().(*EmbeddingResponse), nil
}

// Metadata returns provider metadata
func (p *OpenAI) Metadata() Metadata {
	return Metadata{
		Name:         "OpenAI",
		APIBase:      p.apiBase,
		DefaultModel: "gpt-4",
		Capabilities: []Capability{
			CapChat, CapStream, CapEmbedding, CapVision, CapTools, CapJSON,
		},
		Region:        "international",
		Documentation: "https://platform.openai.com/docs",
	}
}
