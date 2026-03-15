package vectordb

import (
	"context"
	"errors"
)

// Vector represents a vector with metadata
type Vector struct {
	ID       string                 `json:"id"`
	Values   []float32              `json:"values"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// SearchResult represents a search result
type SearchResult struct {
	ID       string                 `json:"id"`
	Score    float32                `json:"score"`
	Vector   *Vector                `json:"vector,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// CollectionInfo represents collection information
type CollectionInfo struct {
	Name       string `json:"name"`
	Dimension  int    `json:"dimension"`
	VectorCount int   `json:"vector_count"`
}

// VectorStore is the interface for vector stores
type VectorStore interface {
	// CreateCollection creates a new collection
	CreateCollection(ctx context.Context, name string, dimension int) error

	// DeleteCollection deletes a collection
	DeleteCollection(ctx context.Context, name string) error

	// ListCollections lists all collections
	ListCollections(ctx context.Context) ([]CollectionInfo, error)

	// Insert inserts vectors into a collection
	Insert(ctx context.Context, collection string, vectors []Vector) error

	// Update updates vectors in a collection
	Update(ctx context.Context, collection string, vectors []Vector) error

	// Delete deletes vectors from a collection
	Delete(ctx context.Context, collection string, ids []string) error

	// Search searches for similar vectors
	Search(ctx context.Context, collection string, query []float32, topK int, filter map[string]interface{}) ([]SearchResult, error)

	// Get retrieves vectors by ID
	Get(ctx context.Context, collection string, ids []string) ([]Vector, error)

	// GetStats returns collection statistics
	GetStats(ctx context.Context, collection string) (*CollectionInfo, error)

	// Name returns the store name
	Name() string
}

// Config represents vector store configuration
type Config struct {
	// Connection URL or endpoint
	URL string `json:"url"`

	// API key (if required)
	APIKey string `json:"api_key"`

	// Collection name
	Collection string `json:"collection"`

	// Vector dimension
	Dimension int `json:"dimension"`

	// Distance metric (cosine, euclidean, dot)
	DistanceMetric string `json:"distance_metric"`

	// Additional configuration
	Options map[string]interface{} `json:"options"`
}

// DistanceMetric represents distance metric type
type DistanceMetric string

const (
	DistanceCosine   DistanceMetric = "cosine"
	DistanceEuclidean DistanceMetric = "euclidean"
	DistanceDot       DistanceMetric = "dot"
)

// NewVectorStore creates a new vector store based on type
func NewVectorStore(storeType string, config Config) (VectorStore, error) {
	switch storeType {
	case "memory":
		return NewMemoryStore(config), nil
	case "qdrant":
		return NewQdrantStore(config)
	case "pinecone":
		return NewPineconeStore(config)
	case "milvus":
		return NewMilvusStore(config)
	case "pgvector":
		return NewPgVectorStore(config)
	default:
		return nil, ErrUnsupportedStore
	}
}

// Errors

var (
	ErrUnsupportedStore   = errors.New("unsupported vector store type")
	ErrCollectionNotFound = errors.New("collection not found")
	ErrVectorNotFound     = errors.New("vector not found")
	ErrInvalidDimension   = errors.New("invalid vector dimension")
)
