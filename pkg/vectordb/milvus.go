package vectordb

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
)

// MilvusStore implements Milvus vector database
type MilvusStore struct {
	client *resty.Client
	config Config
}

// NewMilvusStore creates a new Milvus store
func NewMilvusStore(config Config) (*MilvusStore, error) {
	if config.URL == "" {
		config.URL = "http://localhost:19530"
	}

	client := resty.New().
		SetBaseURL(config.URL + "/v2/vectordb").
		SetHeader("Content-Type", "application/json")

	return &MilvusStore{
		client: client,
		config: config,
	}, nil
}

// CreateCollection creates a new collection
func (s *MilvusStore) CreateCollection(ctx context.Context, name string, dimension int) error {
	req := map[string]interface{}{
		"collectionName": name,
		"dimension":      dimension,
		"metricType":     s.getMetricType(),
	}

	resp, err := s.client.R().
		SetContext(ctx).
		SetBody(req).
		Post("/collections")

	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("failed to create collection: %s", resp.String())
	}

	return nil
}

// DeleteCollection deletes a collection
func (s *MilvusStore) DeleteCollection(ctx context.Context, name string) error {
	resp, err := s.client.R().
		SetContext(ctx).
		Delete("/collections/" + name)

	if err != nil {
		return fmt.Errorf("failed to delete collection: %w", err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("failed to delete collection: %s", resp.String())
	}

	return nil
}

// ListCollections lists all collections
func (s *MilvusStore) ListCollections(ctx context.Context) ([]CollectionInfo, error) {
	var result struct {
		Data []struct {
			CollectionName string `json:"collectionName"`
		} `json:"data"`
	}

	resp, err := s.client.R().
		SetContext(ctx).
		SetResult(&result).
		Get("/collections")

	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("failed to list collections: %s", resp.String())
	}

	infos := make([]CollectionInfo, len(result.Data))
	for i, coll := range result.Data {
		infos[i] = CollectionInfo{Name: coll.CollectionName}
	}

	return infos, nil
}

// Insert inserts vectors into a collection
func (s *MilvusStore) Insert(ctx context.Context, collection string, vectors []Vector) error {
	data := make([]map[string]interface{}, len(vectors))
	for i, vec := range vectors {
		data[i] = map[string]interface{}{
			"id":       vec.ID,
			"vector":   vec.Values,
		}
		// Add metadata fields
		for k, v := range vec.Metadata {
			data[i][k] = v
		}
	}

	req := map[string]interface{}{
		"collectionName": collection,
		"data":          data,
	}

	resp, err := s.client.R().
		SetContext(ctx).
		SetBody(req).
		Post("/entities")

	if err != nil {
		return fmt.Errorf("failed to insert vectors: %w", err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("failed to insert vectors: %s", resp.String())
	}

	return nil
}

// Update updates vectors in a collection
func (s *MilvusStore) Update(ctx context.Context, collection string, vectors []Vector) error {
	// Milvus uses upsert for update
	return s.Insert(ctx, collection, vectors)
}

// Delete deletes vectors from a collection
func (s *MilvusStore) Delete(ctx context.Context, collection string, ids []string) error {
	req := map[string]interface{}{
		"collectionName": collection,
		"id":            ids,
	}

	resp, err := s.client.R().
		SetContext(ctx).
		SetBody(req).
		Delete("/entities")

	if err != nil {
		return fmt.Errorf("failed to delete vectors: %w", err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("failed to delete vectors: %s", resp.String())
	}

	return nil
}

// Search searches for similar vectors
func (s *MilvusStore) Search(ctx context.Context, collection string, query []float32, topK int, filter map[string]interface{}) ([]SearchResult, error) {
	req := map[string]interface{}{
		"collectionName": collection,
		"vector":         query,
		"limit":          topK,
		"outputFields":   []string{"*"},
	}

	if filter != nil {
		req["filter"] = s.buildFilter(filter)
	}

	var result struct {
		Data []struct {
			ID     string                 `json:"id"`
			Distance float32              `json:"distance"`
			Fields map[string]interface{} `json:"fields"`
		} `json:"data"`
	}

	resp, err := s.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&result).
		Post("/search")

	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("failed to search: %s", resp.String())
	}

	results := make([]SearchResult, len(result.Data))
	for i, r := range result.Data {
		results[i] = SearchResult{
			ID:       r.ID,
			Score:    r.Distance,
			Metadata: r.Fields,
		}
	}

	return results, nil
}

// Get retrieves vectors by ID
func (s *MilvusStore) Get(ctx context.Context, collection string, ids []string) ([]Vector, error) {
	req := map[string]interface{}{
		"collectionName": collection,
		"id":            ids,
		"outputFields":   []string{"*"},
	}

	var result struct {
		Data []struct {
			ID     string                 `json:"id"`
			Vector []float32              `json:"vector"`
			Fields map[string]interface{} `json:"fields"`
		} `json:"data"`
	}

	resp, err := s.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&result).
		Get("/entities")

	if err != nil {
		return nil, fmt.Errorf("failed to get vectors: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("failed to get vectors: %s", resp.String())
	}

	vectors := make([]Vector, len(result.Data))
	for i, r := range result.Data {
		vectors[i] = Vector{
			ID:       r.ID,
			Values:   r.Vector,
			Metadata: r.Fields,
		}
	}

	return vectors, nil
}

// GetStats returns collection statistics
func (s *MilvusStore) GetStats(ctx context.Context, collection string) (*CollectionInfo, error) {
	var result struct {
		Data struct {
			CollectionName string `json:"collectionName"`
			RowCount       int    `json:"rowCount"`
		} `json:"data"`
	}

	resp, err := s.client.R().
		SetContext(ctx).
		SetResult(&result).
		Get("/collections/" + collection)

	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("failed to get stats: %s", resp.String())
	}

	return &CollectionInfo{
		Name:        result.Data.CollectionName,
		VectorCount: result.Data.RowCount,
	}, nil
}

// Name returns the store name
func (s *MilvusStore) Name() string {
	return "milvus"
}

// Helper functions

func (s *MilvusStore) getMetricType() string {
	switch s.config.DistanceMetric {
	case string(DistanceEuclidean):
		return "L2"
	case string(DistanceDot):
		return "IP"
	default:
		return "COSINE"
	}
}

func (s *MilvusStore) buildFilter(filter map[string]interface{}) string {
	// Build Milvus filter expression
	// Example: "color == 'red' and price > 100"
	var expr string
	first := true
	for key, value := range filter {
		if !first {
			expr += " and "
		}
		expr += fmt.Sprintf("%s == '%v'", key, value)
		first = false
	}
	return expr
}
