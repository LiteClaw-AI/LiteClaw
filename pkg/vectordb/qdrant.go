package vectordb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-resty/resty/v2"
)

// QdrantStore implements Qdrant vector database
type QdrantStore struct {
	client *resty.Client
	config Config
}

// NewQdrantStore creates a new Qdrant store
func NewQdrantStore(config Config) (*QdrantStore, error) {
	if config.URL == "" {
		config.URL = "http://localhost:6333"
	}

	client := resty.New().
		SetBaseURL(config.URL).
		SetHeader("Content-Type", "application/json")

	if config.APIKey != "" {
		client.SetHeader("api-key", config.APIKey)
	}

	return &QdrantStore{
		client: client,
		config: config,
	}, nil
}

// CreateCollection creates a new collection
func (s *QdrantStore) CreateCollection(ctx context.Context, name string, dimension int) error {
	req := map[string]interface{}{
		"vectors": map[string]interface{}{
			"size":     dimension,
			"distance": s.getDistanceMetric(),
		},
	}

	resp, err := s.client.R().
		SetContext(ctx).
		SetBody(req).
		Put("/collections/" + name)

	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("failed to create collection: %s", resp.String())
	}

	return nil
}

// DeleteCollection deletes a collection
func (s *QdrantStore) DeleteCollection(ctx context.Context, name string) error {
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
func (s *QdrantStore) ListCollections(ctx context.Context) ([]CollectionInfo, error) {
	var result struct {
		Result struct {
			Collections []struct {
				Name string `json:"name"`
			} `json:"collections"`
		} `json:"result"`
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

	infos := make([]CollectionInfo, len(result.Result.Collections))
	for i, coll := range result.Result.Collections {
		infos[i] = CollectionInfo{Name: coll.Name}
	}

	return infos, nil
}

// Insert inserts vectors into a collection
func (s *QdrantStore) Insert(ctx context.Context, collection string, vectors []Vector) error {
	points := make([]map[string]interface{}, len(vectors))
	for i, vec := range vectors {
		points[i] = map[string]interface{}{
			"id":       vec.ID,
			"vector":   vec.Values,
			"payload":  vec.Metadata,
		}
	}

	req := map[string]interface{}{
		"points": points,
	}

	resp, err := s.client.R().
		SetContext(ctx).
		SetBody(req).
		Put("/collections/" + collection + "/points")

	if err != nil {
		return fmt.Errorf("failed to insert vectors: %w", err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("failed to insert vectors: %s", resp.String())
	}

	return nil
}

// Search searches for similar vectors
func (s *QdrantStore) Search(ctx context.Context, collection string, query []float32, topK int, filter map[string]interface{}) ([]SearchResult, error) {
	req := map[string]interface{}{
		"vector":    query,
		"limit":     topK,
		"with_payload": true,
	}

	if filter != nil {
		req["filter"] = s.buildFilter(filter)
	}

	var result struct {
		Result []struct {
			ID      string                 `json:"id"`
			Score   float32                `json:"score"`
			Payload map[string]interface{} `json:"payload"`
		} `json:"result"`
	}

	resp, err := s.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&result).
		Post("/collections/" + collection + "/points/search")

	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("failed to search: %s", resp.String())
	}

	results := make([]SearchResult, len(result.Result))
	for i, r := range result.Result {
		results[i] = SearchResult{
			ID:       r.ID,
			Score:    r.Score,
			Metadata: r.Payload,
		}
	}

	return results, nil
}

// Get retrieves vectors by ID
func (s *QdrantStore) Get(ctx context.Context, collection string, ids []string) ([]Vector, error) {
	req := map[string]interface{}{
		"ids": ids,
	}

	var result struct {
		Result []struct {
			ID     string                 `json:"id"`
			Vector []float32              `json:"vector"`
			Payload map[string]interface{} `json:"payload"`
		} `json:"result"`
	}

	resp, err := s.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&result).
		Post("/collections/" + collection + "/points")

	if err != nil {
		return nil, fmt.Errorf("failed to get vectors: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("failed to get vectors: %s", resp.String())
	}

	vectors := make([]Vector, len(result.Result))
	for i, r := range result.Result {
		vectors[i] = Vector{
			ID:       r.ID,
			Values:   r.Vector,
			Metadata: r.Payload,
		}
	}

	return vectors, nil
}

// Update, Delete, GetStats implementations...

func (s *QdrantStore) Update(ctx context.Context, collection string, vectors []Vector) error {
	return s.Insert(ctx, collection, vectors)
}

func (s *QdrantStore) Delete(ctx context.Context, collection string, ids []string) error {
	req := map[string]interface{}{
		"points": ids,
	}

	resp, err := s.client.R().
		SetContext(ctx).
		SetBody(req).
		Post("/collections/" + collection + "/points/delete")

	if err != nil {
		return fmt.Errorf("failed to delete vectors: %w", err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("failed to delete vectors: %s", resp.String())
	}

	return nil
}

func (s *QdrantStore) GetStats(ctx context.Context, collection string) (*CollectionInfo, error) {
	var result struct {
		Result struct {
			PointsCount int `json:"points_count"`
			VectorsCount int `json:"vectors_count"`
		} `json:"result"`
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
		Name:        collection,
		VectorCount: result.Result.PointsCount,
	}, nil
}

// Name returns the store name
func (s *QdrantStore) Name() string {
	return "qdrant"
}

// Helper functions

func (s *QdrantStore) getDistanceMetric() string {
	switch s.config.DistanceMetric {
	case string(DistanceEuclidean):
		return "Euclid"
	case string(DistanceDot):
		return "Dot"
	default:
		return "Cosine"
	}
}

func (s *QdrantStore) buildFilter(filter map[string]interface{}) map[string]interface{} {
	// Convert simple filter to Qdrant filter format
	var conditions []map[string]interface{}
	for key, value := range filter {
		conditions = append(conditions, map[string]interface{}{
			"key":   key,
			"match": map[string]interface{}{"value": value},
		})
	}

	return map[string]interface{}{
		"must": conditions,
	}
}
