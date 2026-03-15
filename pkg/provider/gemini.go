package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

// Gemini provider implementation (Google AI)
type Gemini struct {
	*BaseProvider
	client *resty.Client
}

// NewGemini creates a new Gemini provider
func NewGemini(apiKey string) *Gemini {
	return &Gemini{
		BaseProvider: NewBaseProvider("gemini", apiKey, "https://generativelanguage.googleapis.com/v1beta", "gemini-1.5-flash"),
		client: resty.New().
			SetQueryParam("key", apiKey).
			SetHeader("Content-Type", "application/json"),
	}
}

// Chat performs a chat completion
func (p *Gemini) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	if req.Model == "" {
		req.Model = p.model
	}

	// Convert to Gemini format
	geminiReq := map[string]interface{}{
		"contents": p.convertMessagesToGeminiFormat(req.Messages),
		"generationConfig": map[string]interface{}{
			"temperature": req.Temperature,
		},
	}

	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		UsageMetadata struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
			TotalTokenCount      int `json:"totalTokenCount"`
		} `json:"usageMetadata"`
	}

	resp, err := p.client.R().
		SetContext(ctx).
		SetBody(geminiReq).
		SetResult(&geminiResp).
		Post(fmt.Sprintf("%s/models/%s:generateContent", p.apiBase, req.Model))

	if err != nil {
		return nil, fmt.Errorf("gemini chat request failed: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("gemini chat request failed: %s", resp.String())
	}

	if len(geminiResp.Candidates) == 0 {
		return nil, fmt.Errorf("gemini returned no candidates")
	}

	return &ChatResponse{
		ID:      fmt.Sprintf("gemini-%d", time.Now().Unix()),
		Object:  "chat.completion",
		Model:   req.Model,
		Choices: []Choice{
			{
				Index: 0,
				Message: Message{
					Role:    RoleAssistant,
					Content: geminiResp.Candidates[0].Content.Parts[0].Text,
				},
				FinishReason: "stop",
			},
		},
		Usage: Usage{
			PromptTokens:     geminiResp.UsageMetadata.PromptTokenCount,
			CompletionTokens: geminiResp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      geminiResp.UsageMetadata.TotalTokenCount,
		},
	}, nil
}

// ChatStream performs a streaming chat completion
func (p *Gemini) ChatStream(ctx context.Context, req *ChatRequest) (<-chan StreamChunk, error) {
	// TODO: Implement Gemini streaming
	return nil, fmt.Errorf("streaming not yet implemented for Gemini")
}

// Embed generates embeddings
func (p *Gemini) Embed(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {
	if req.Model == "" {
		req.Model = "text-embedding-004"
	}

	geminiReq := map[string]interface{}{
		"model": fmt.Sprintf("models/%s", req.Model),
		"content": map[string]interface{}{
			"parts": []map[string]string{
				{"text": req.Input},
			},
		},
	}

	var geminiResp struct {
		Embedding struct {
			Values []float32 `json:"values"`
		} `json:"embedding"`
	}

	resp, err := p.client.R().
		SetContext(ctx).
		SetBody(geminiReq).
		SetResult(&geminiResp).
		Post(fmt.Sprintf("%s/models/%s:embedContent", p.apiBase, req.Model))

	if err != nil {
		return nil, fmt.Errorf("gemini embedding request failed: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("gemini embedding request failed: %s", resp.String())
	}

	return &EmbeddingResponse{
		Object: "list",
		Data: []EmbeddingData{
			{
				Object:    "embedding",
				Index:     0,
				Embedding: geminiResp.Embedding.Values,
			},
		},
		Model: req.Model,
	}, nil
}

// Metadata returns provider metadata
func (p *Gemini) Metadata() Metadata {
	return Metadata{
		Name:         "Google Gemini",
		APIBase:      p.apiBase,
		DefaultModel: "gemini-1.5-flash",
		Capabilities: []Capability{CapChat, CapEmbedding, CapVision, CapAudio, CapJSON},
		Region:       "international",
		Documentation: "https://ai.google.dev/docs",
	}
}

// Helper functions
func (p *Gemini) convertMessagesToGeminiFormat(messages []Message) []map[string]interface{} {
	var contents []map[string]interface{}
	for _, msg := range messages {
		role := "user"
		if msg.Role == RoleAssistant {
			role = "model"
		}
		contents = append(contents, map[string]interface{}{
			"role": role,
			"parts": []map[string]string{
				{"text": msg.Content},
			},
		})
	}
	return contents
}
