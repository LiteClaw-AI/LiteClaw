package provider

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
)

// Cohere provider implementation
type Cohere struct {
	*BaseProvider
	client *resty.Client
}

// NewCohere creates a new Cohere provider
func NewCohere(apiKey string) *Cohere {
	return &Cohere{
		BaseProvider: NewBaseProvider("cohere", apiKey, "https://api.cohere.ai/v1", "command-r-plus"),
		client: resty.New().
			SetHeader("Authorization", "Bearer "+apiKey).
			SetHeader("Content-Type", "application/json"),
	}
}

// Chat performs a chat completion
func (p *Cohere) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	if req.Model == "" {
		req.Model = p.model
	}

	// Convert to Cohere format
	cohereReq := map[string]interface{}{
		"model":    req.Model,
		"message":  p.extractLastUserMessage(req.Messages),
		"chat_history": p.convertMessagesToCohereFormat(req.Messages),
	}

	if req.Temperature > 0 {
		cohereReq["temperature"] = req.Temperature
	}

	var cohereResp struct {
		Text          string `json:"text"`
		GenerationID  string `json:"generation_id"`
		Meta          struct {
			Tokens struct {
				InputTokens  int `json:"input_tokens"`
				OutputTokens int `json:"output_tokens"`
			} `json:"tokens"`
		} `json:"meta"`
	}

	resp, err := p.client.R().
		SetContext(ctx).
		SetBody(cohereReq).
		SetResult(&cohereResp).
		Post(p.apiBase + "/chat")

	if err != nil {
		return nil, fmt.Errorf("cohere chat request failed: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("cohere chat request failed: %s", resp.String())
	}

	return &ChatResponse{
		ID:      cohereResp.GenerationID,
		Object:  "chat.completion",
		Model:   req.Model,
		Choices: []Choice{
			{
				Index: 0,
				Message: Message{
					Role:    RoleAssistant,
					Content: cohereResp.Text,
				},
				FinishReason: "stop",
			},
		},
		Usage: Usage{
			PromptTokens:     cohereResp.Meta.Tokens.InputTokens,
			CompletionTokens: cohereResp.Meta.Tokens.OutputTokens,
			TotalTokens:      cohereResp.Meta.Tokens.InputTokens + cohereResp.Meta.Tokens.OutputTokens,
		},
	}, nil
}

// ChatStream performs a streaming chat completion
func (p *Cohere) ChatStream(ctx context.Context, req *ChatRequest) (<-chan StreamChunk, error) {
	// TODO: Implement Cohere streaming
	return nil, fmt.Errorf("streaming not yet implemented for Cohere")
}

// Embed generates embeddings
func (p *Cohere) Embed(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {
	if req.Model == "" {
		req.Model = "embed-english-v3.0"
	}

	cohereReq := map[string]interface{}{
		"model": req.Model,
		"texts": []string{req.Input},
	}

	var cohereResp struct {
		Embeddings [][]float32 `json:"embeddings"`
		Meta       struct {
			Tokens struct {
				InputTokens int `json:"input_tokens"`
			} `json:"tokens"`
		} `json:"meta"`
	}

	resp, err := p.client.R().
		SetContext(ctx).
		SetBody(cohereReq).
		SetResult(&cohereResp).
		Post(p.apiBase + "/embed")

	if err != nil {
		return nil, fmt.Errorf("cohere embedding request failed: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("cohere embedding request failed: %s", resp.String())
	}

	return &EmbeddingResponse{
		Object: "list",
		Data: []EmbeddingData{
			{
				Object:    "embedding",
				Index:     0,
				Embedding: cohereResp.Embeddings[0],
			},
		},
		Model: req.Model,
		Usage: Usage{
			PromptTokens: cohereResp.Meta.Tokens.InputTokens,
		},
	}, nil
}

// Metadata returns provider metadata
func (p *Cohere) Metadata() Metadata {
	return Metadata{
		Name:         "Cohere",
		APIBase:      p.apiBase,
		DefaultModel: "command-r-plus",
		Capabilities: []Capability{CapChat, CapEmbedding, CapTools, CapJSON},
		Region:       "international",
		Documentation: "https://docs.cohere.com",
	}
}

// Helper functions
func (p *Cohere) extractLastUserMessage(messages []Message) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == RoleUser {
			return messages[i].Content
		}
	}
	return ""
}

func (p *Cohere) convertMessagesToCohereFormat(messages []Message) []map[string]string {
	var history []map[string]string
	for i := 0; i < len(messages)-1; i++ {
		msg := messages[i]
		role := "USER"
		if msg.Role == RoleAssistant {
			role = "CHATBOT"
		}
		history = append(history, map[string]string{
			"role":    role,
			"message": msg.Content,
		})
	}
	return history
}
