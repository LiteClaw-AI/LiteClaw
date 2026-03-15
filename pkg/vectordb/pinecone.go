package vectordb

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
)

// PineconeStore implements Pinecone vector database
type PineconeStore struct {
	client   *resty.Client
	config   Config
	indexURL string
}

// NewPineconeStore creates a new Pinecone store
func NewPineconeStore(config Config) (*PineconeStore, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("pinecone API key is required")
	}

	// Extract index URL from config or build from environment
	indexURL := config.URL
	if indexURL == "" {
		// Format: https://index-name-project-id.svc.environment.pinecone.io
		indexURL = fmt.Sprintf("https://%s", config.Options["index_host"])
	}

	client := resty.New().
		SetBaseURL(indexURL).
		SetHeader("Content-Type", "application/json").
		SetHeader("Api-Key", config.APIKey)

	return &PineconeStore{
		client:   client,
		config:   config,
		indexURL: indexURL,
	}, nil
}

// CreateCollection creates a new collection (index)
// Note: In Pinecone, collections are called indexes and must be created via API/console
func (s *PineconeStore) CreateCollection(ctx context.Context, name string, dimension int) error {
	// Pinecone indexes are created via control plane API, not data plane
	// This is a placeholder that returns an error
	return fmt.Errorf("pinecone indexes must be created via Pinecone console or control plane API")
}

// DeleteCollection deletes a collection (index)
func (s *PineconeStore) DeleteCollection(ctx context.Context, name string) error {
	// Same as CreateCollection, this must be done via control plane
	return fmt.Errorf("pinecone indexes must be deleted via Pinecone console or control plane API")
}

// ListCollections lists all collections (indexes)
func (s *PineconeStore) ListCollections(ctx context.Context) ([]CollectionInfo, error) {
	// This requires the control plane API
	// Return the configured index name
	return []CollectionInfo{
		{Name: s.config.Collection},
	}, nil
}

// Insert inserts vectors into a collection
func (s *PineconeStore) Insert(ctx context.Context, collection string, vectors []Vector) error {
	vectorsReq := make([]map[string]interface{}, len(vectors))
	for i, vec := range vectors {
		vectorsReq[i] = map[string]interface{}{
			"id":       vec.ID,
			"values":   vec.Values,
			"metadata": vec.Metadata,
		}
	}

	req := map[string]interface{}{
		"vectors": vectorsReq,
	}

	resp, err := s.client.R().
		SetContext(ctx).
		SetBody(req).
		Post("/vectors/upsert")

	if err != nil {
		return fmt.Errorf("failed to insert vectors: %w", err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("failed to insert vectors: %s", resp.String())
	}

	return nil
}

// Update updates vectors in a collection
func (s *PineconeStore) Update(ctx context.Context, collection string, vectors []Vector) error {
	return s.Insert(ctx, collection, vectors)
}

// Delete deletes vectors from a collection
func (s *PineconeStore) Delete(ctx context.Context, collection string, ids []string) error {
	req := map[string]interface{}{
		"ids": ids,
	}

	resp, err := s.client.R().
		SetContext(ctx).
		SetBody(req).
		Post("/vectors/delete")

	if err != nil {
		return fmt.Errorf("failed to delete vectors: %w", err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("failed to delete vectors: %s", resp.String())
	}

	return nil
}

// Search searches for similar vectors
func (s *PineconeStore) Search(ctx context.Context, collection string, query []float32, topK int, filter map[string]interface{}) ([]SearchResult, error) {
	req := map[string]interface{}{
		"vector":         query,
		"topK":          topK,
		"includeMetadata": true,
	}

	if filter != nil {
		req["filter"] = filter
	}

	var result struct {
		Matches []struct {
			ID       string                 `json:"id"`
			Score    float32                `json:"score"`
			Metadata map[string]interface{} `json:"metadata"`
		} `json:"matches"`
	}

	resp, err := s.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&result).
		Post("/query")

	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("failed to search: %s", resp.String())
	}

	results := make([]SearchResult, len(result.Matches))
	for i, m := range result.Matches {
		results[i] = SearchResult{
			ID:       m.ID,
			Score:    m.Score,
			Metadata: m.Metadata,
		}
	}

	return results, nil
}

// Get retrieves vectors by ID
func (s *PineconeStore) Get(ctx context.Context, collection string, ids []string) ([]Vector, error) {
	req := map[string]interface{}{
		"ids": ids,
	}

	var result struct {
		Vectors []struct {
			ID       string                 `json:"id"`
			Values   []float32              `json:"values"`
			Metadata map[string]interface{} `json:"metadata"`
		} `json:"vectors"`
	}

	resp, err := s.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&result).
		Post("/vectors/fetch")

	if err != nil {
		return nil, fmt.Errorf("failed to get vectors: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("failed to get vectors: %s", resp.String())
	}

	vectors := make([]Vector, len(result.Vectors))
	for i, v := range result.Vectors {
		vectors[i] = Vector{
			ID:       v.ID,
			Values:   v.Values,
			Metadata: v.Metadata,
		}
	}

	return vectors, nil
}

// GetStats returns collection statistics
func (s *PineconeStore) GetStats(ctx context.Context, collection string) (*CollectionInfo, error) {
	var result struct {
		TotalVectorCount int `json:"totalVectorCount"`
		Dimension        int `json:"dimension"`
	}

	resp, err := s.client.R().
		SetContext(ctx).
		SetResult(&result).
		Get("/describe-index-stats")

	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("failed to get stats: %s", resp.String())
	}

	return &CollectionInfo{
		Name:        collection,
		Dimension:   result.Dimension,
		VectorCount: result.TotalVectorCount,
	}, nil
}

// Name returns the store name
func (s *PineconeStore) Name() string {
	return "pinecone"
}
