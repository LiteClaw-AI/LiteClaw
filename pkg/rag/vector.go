package rag

import (
	"context"
	"math"
	"sort"
	"sync"
)

// VectorStore is the interface for vector storage
type VectorStore interface {
	Add(id string, vector []float32, metadata map[string]string) error
	Search(query []float32, topK int) ([]SearchResult, error)
	Delete(id string) error
	Get(id string) ([]float32, error)
	Count() int
	Clear() error
}

// SearchResult represents a search result
type SearchResult struct {
	ID       string            `json:"id"`
	Score    float64           `json:"score"`
	Metadata map[string]string `json:"metadata"`
}

// InMemoryVectorStore stores vectors in memory
type InMemoryVectorStore struct {
	vectors map[string]VectorEntry
	mu      sync.RWMutex
}

// VectorEntry represents a stored vector
type VectorEntry struct {
	Vector   []float32
	Metadata map[string]string
}

// NewInMemoryVectorStore creates a new in-memory store
func NewInMemoryVectorStore() *InMemoryVectorStore {
	return &InMemoryVectorStore{
		vectors: make(map[string]VectorEntry),
	}
}

// Add adds a vector to the store
func (s *InMemoryVectorStore) Add(id string, vector []float32, metadata map[string]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.vectors[id] = VectorEntry{
		Vector:   vector,
		Metadata: metadata,
	}
	return nil
}

// Search searches for similar vectors
func (s *InMemoryVectorStore) Search(query []float32, topK int) ([]SearchResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []SearchResult
	for id, entry := range s.vectors {
		score := cosineSimilarity(query, entry.Vector)
		if score > 0 {
			results = append(results, SearchResult{
				ID:       id,
				Score:    score,
				Metadata: entry.Metadata,
			})
		}
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Limit to topK
	if len(results) > topK {
		results = results[:topK]
	}

	return results, nil
}

// Delete removes a vector from the store
func (s *InMemoryVectorStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.vectors, id)
	return nil
}

// Get retrieves a vector by ID
func (s *InMemoryVectorStore) Get(id string) ([]float32, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.vectors[id]
	if !exists {
		return nil, nil
	}
	return entry.Vector, nil
}

// Count returns the number of stored vectors
func (s *InMemoryVectorStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.vectors)
}

// Clear removes all vectors
func (s *InMemoryVectorStore) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.vectors = make(map[string]VectorEntry)
	return nil
}

// cosineSimilarity computes cosine similarity between two vectors
func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

// EmbeddingEngine generates embeddings
type EmbeddingEngine struct {
	provider EmbeddingProvider
	dimension int
}

// EmbeddingProvider provides embedding functionality
type EmbeddingProvider interface {
	Embed(ctx context.Context, texts []string) ([][]float32, error)
}

// NewEmbeddingEngine creates a new embedding engine
func NewEmbeddingEngine(provider EmbeddingProvider, dimension int) *EmbeddingEngine {
	return &EmbeddingEngine{
		provider: provider,
		dimension: dimension,
	}
}

// Embed generates embeddings for texts
func (e *EmbeddingEngine) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	return e.provider.Embed(ctx, texts)
}

// EmbedOne generates embedding for a single text
func (e *EmbeddingEngine) EmbedOne(ctx context.Context, text string) ([]float32, error) {
	embeddings, err := e.provider.Embed(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, nil
	}
	return embeddings[0], nil
}

// PseudoEmbeddingProvider generates pseudo-embeddings for testing
type PseudoEmbeddingProvider struct {
	dimension int
}

// NewPseudoEmbeddingProvider creates a pseudo embedding provider
func NewPseudoEmbeddingProvider(dimension int) *PseudoEmbeddingProvider {
	return &PseudoEmbeddingProvider{dimension: dimension}
}

// Embed generates pseudo-embeddings
func (p *PseudoEmbeddingProvider) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))

	for i, text := range texts {
		embedding := make([]float32, p.dimension)
		runes := []rune(text)

		for j, r := range runes {
			idx := (j*7 + int(r)) % p.dimension
			embedding[idx] += 1.0
		}

		// Normalize
		var norm float32
		for _, v := range embedding {
			norm += v * v
		}
		if norm > 0 {
			norm = float32(math.Sqrt(float64(norm)))
			for j := range embedding {
				embedding[j] /= norm
			}
		}

		embeddings[i] = embedding
	}

	return embeddings, nil
}
