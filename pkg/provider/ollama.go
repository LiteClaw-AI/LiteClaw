package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/go-resty/resty/v2"
)

// Ollama provider implementation (local LLM)
type Ollama struct {
	*BaseProvider
	client *resty.Client
}

// NewOllama creates a new Ollama provider
func NewOllama(baseURL string) *Ollama {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	return &Ollama{
		BaseProvider: NewBaseProvider("ollama", "", baseURL+"/api", "llama3.2"),
		client: resty.New().
			SetHeader("Content-Type", "application/json"),
	}
}

// Chat performs a chat completion
func (p *Ollama) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	if req.Model == "" {
		req.Model = p.model
	}

	// Convert to Ollama format
	ollamaReq := map[string]interface{}{
		"model":    req.Model,
		"messages": req.Messages,
		"stream":   false,
	}

	if req.Temperature > 0 {
		ollamaReq["options"] = map[string]interface{}{
			"temperature": req.Temperature,
		}
	}

	var ollamaResp struct {
		Model     string    `json:"model"`
		CreatedAt time.Time `json:"created_at"`
		Message   struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		Done bool `json:"done"`
	}

	resp, err := p.client.R().
		SetContext(ctx).
		SetBody(ollamaReq).
		SetResult(&ollamaResp).
		Post(p.apiBase + "/chat")

	if err != nil {
		return nil, fmt.Errorf("ollama chat request failed: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("ollama chat request failed: %s", resp.String())
	}

	return &ChatResponse{
		ID:      fmt.Sprintf("ollama-%d", ollamaResp.CreatedAt.Unix()),
		Object:  "chat.completion",
		Model:   ollamaResp.Model,
		Choices: []Choice{
			{
				Index: 0,
				Message: Message{
					Role:    RoleAssistant,
					Content: ollamaResp.Message.Content,
				},
				FinishReason: "stop",
			},
		},
	}, nil
}

// ChatStream performs a streaming chat completion
func (p *Ollama) ChatStream(ctx context.Context, req *ChatRequest) (<-chan StreamChunk, error) {
	if req.Model == "" {
		req.Model = p.model
	}

	ollamaReq := map[string]interface{}{
		"model":    req.Model,
		"messages": req.Messages,
		"stream":   true,
	}

	stream := make(chan StreamChunk, 100)

	go func() {
		defer close(stream)

		resp, err := p.client.R().
			SetContext(ctx).
			SetBody(ollamaReq).
			SetDoNotParseResponse(true).
			Post(p.apiBase + "/chat")

		if err != nil {
			return
		}
		defer resp.RawBody().Close()

		// Parse NDJSON stream
		decoder := json.NewDecoder(resp.RawBody())
		for {
			var chunk struct {
				Model     string `json:"model"`
				CreatedAt string `json:"created_at"`
				Message   struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"message"`
				Done bool `json:"done"`
			}

			if err := decoder.Decode(&chunk); err != nil {
				if err == io.EOF {
					break
				}
				continue
			}

			stream <- StreamChunk{
				ID:      fmt.Sprintf("ollama-%s", chunk.CreatedAt),
				Object:  "chat.completion.chunk",
				Model:   chunk.Model,
				Choices: []Choice{
					{
						Index: 0,
						Delta: &Message{
							Role:    RoleAssistant,
							Content: chunk.Message.Content,
						},
					},
				},
			}

			if chunk.Done {
				break
			}
		}
	}()

	return stream, nil
}

// Embed generates embeddings
func (p *Ollama) Embed(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {
	if req.Model == "" {
		req.Model = "nomic-embed-text"
	}

	ollamaReq := map[string]interface{}{
		"model":  req.Model,
		"prompt": req.Input,
	}

	var ollamaResp struct {
		Embedding []float32 `json:"embedding"`
	}

	resp, err := p.client.R().
		SetContext(ctx).
		SetBody(ollamaReq).
		SetResult(&ollamaResp).
		Post(p.apiBase + "/embeddings")

	if err != nil {
		return nil, fmt.Errorf("ollama embedding request failed: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("ollama embedding request failed: %s", resp.String())
	}

	return &EmbeddingResponse{
		Object: "list",
		Data: []EmbeddingData{
			{
				Object:    "embedding",
				Index:     0,
				Embedding: ollamaResp.Embedding,
			},
		},
		Model: req.Model,
	}, nil
}

// Metadata returns provider metadata
func (p *Ollama) Metadata() Metadata {
	return Metadata{
		Name:         "Ollama (Local)",
		APIBase:      p.apiBase,
		DefaultModel: "llama3.2",
		Capabilities: []Capability{CapChat, CapStream, CapEmbedding, CapTools, CapJSON},
		Region:       "local",
		Documentation: "https://ollama.ai/docs",
	}
}
