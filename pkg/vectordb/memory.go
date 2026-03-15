package vectordb

import (
	"context"
	"sync"
)

// MemoryStore implements in-memory vector storage
type MemoryStore struct {
	collections map[string]*memoryCollection
	mu          sync.RWMutex
	config      Config
}

type memoryCollection struct {
	name      string
	dimension int
	vectors   map[string]*Vector
	mu        sync.RWMutex
}

// NewMemoryStore creates a new in-memory vector store
func NewMemoryStore(config Config) *MemoryStore {
	return &MemoryStore{
		collections: make(map[string]*memoryCollection),
		config:      config,
	}
}

// CreateCollection creates a new collection
func (s *MemoryStore) CreateCollection(ctx context.Context, name string, dimension int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.collections[name]; exists {
		return nil // Already exists
	}

	s.collections[name] = &memoryCollection{
		name:      name,
		dimension: dimension,
		vectors:   make(map[string]*Vector),
	}
	return nil
}

// DeleteCollection deletes a collection
func (s *MemoryStore) DeleteCollection(ctx context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.collections, name)
	return nil
}

// ListCollections lists all collections
func (s *MemoryStore) ListCollections(ctx context.Context) ([]CollectionInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	infos := make([]CollectionInfo, 0, len(s.collections))
	for name, coll := range s.collections {
		coll.mu.RLock()
		infos = append(infos, CollectionInfo{
			Name:        name,
			Dimension:   coll.dimension,
			VectorCount: len(coll.vectors),
		})
		coll.mu.RUnlock()
	}
	return infos, nil
}

// Insert inserts vectors into a collection
func (s *MemoryStore) Insert(ctx context.Context, collection string, vectors []Vector) error {
	s.mu.RLock()
	coll, exists := s.collections[collection]
	s.mu.RUnlock()

	if !exists {
		return ErrCollectionNotFound
	}

	coll.mu.Lock()
	defer coll.mu.Unlock()

	for _, vec := range vectors {
		if len(vec.Values) != coll.dimension {
			return ErrInvalidDimension
		}
		coll.vectors[vec.ID] = &vec
	}
	return nil
}

// Update updates vectors in a collection
func (s *MemoryStore) Update(ctx context.Context, collection string, vectors []Vector) error {
	return s.Insert(ctx, collection, vectors)
}

// Delete deletes vectors from a collection
func (s *MemoryStore) Delete(ctx context.Context, collection string, ids []string) error {
	s.mu.RLock()
	coll, exists := s.collections[collection]
	s.mu.RUnlock()

	if !exists {
		return ErrCollectionNotFound
	}

	coll.mu.Lock()
	defer coll.mu.Unlock()

	for _, id := range ids {
		delete(coll.vectors, id)
	}
	return nil
}

// Search searches for similar vectors
func (s *MemoryStore) Search(ctx context.Context, collection string, query []float32, topK int, filter map[string]interface{}) ([]SearchResult, error) {
	s.mu.RLock()
	coll, exists := s.collections[collection]
	s.mu.RUnlock()

	if !exists {
		return nil, ErrCollectionNotFound
	}

	coll.mu.RLock()
	defer coll.mu.RUnlock()

	// Calculate cosine similarity for all vectors
	type scoredVector struct {
		vec   *Vector
		score float32
	}

	var candidates []scoredVector
	for _, vec := range coll.vectors {
		// Apply filter if provided
		if !matchesFilter(vec.Metadata, filter) {
			continue
		}

		score := cosineSimilarity(query, vec.Values)
		candidates = append(candidates, scoredVector{vec: vec, score: score})
	}

	// Sort by score (descending) and return top K
	sortByScore(candidates)
	if topK > len(candidates) {
		topK = len(candidates)
	}

	results := make([]SearchResult, topK)
	for i := 0; i < topK; i++ {
		results[i] = SearchResult{
			ID:       candidates[i].vec.ID,
			Score:    candidates[i].score,
			Vector:   candidates[i].vec,
			Metadata: candidates[i].vec.Metadata,
		}
	}

	return results, nil
}

// Get retrieves vectors by ID
func (s *MemoryStore) Get(ctx context.Context, collection string, ids []string) ([]Vector, error) {
	s.mu.RLock()
	coll, exists := s.collections[collection]
	s.mu.RUnlock()

	if !exists {
		return nil, ErrCollectionNotFound
	}

	coll.mu.RLock()
	defer coll.mu.RUnlock()

	vectors := make([]Vector, 0, len(ids))
	for _, id := range ids {
		if vec, exists := coll.vectors[id]; exists {
			vectors = append(vectors, *vec)
		}
	}
	return vectors, nil
}

// GetStats returns collection statistics
func (s *MemoryStore) GetStats(ctx context.Context, collection string) (*CollectionInfo, error) {
	s.mu.RLock()
	coll, exists := s.collections[collection]
	s.mu.RUnlock()

	if !exists {
		return nil, ErrCollectionNotFound
	}

	coll.mu.RLock()
	defer coll.mu.RUnlock()

	return &CollectionInfo{
		Name:        collection,
		Dimension:   coll.dimension,
		VectorCount: len(coll.vectors),
	}, nil
}

// Name returns the store name
func (s *MemoryStore) Name() string {
	return "memory"
}

// Helper functions

func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float32
	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (sqrt32(normA) * sqrt32(normB))
}

func sqrt32(x float32) float32 {
	return float32(sqrt(float64(x)))
}

func sqrt(x float64) float64 {
	// Newton's method
	z := 1.0
	for i := 0; i < 10; i++ {
		z -= (z*z - x) / (2 * z)
	}
	return z
}

func matchesFilter(metadata map[string]interface{}, filter map[string]interface{}) bool {
	if filter == nil || len(filter) == 0 {
		return true
	}

	for key, value := range filter {
		if metaValue, exists := metadata[key]; !exists || metaValue != value {
			return false
		}
	}
	return true
}

func sortByScore(vectors []struct {
	vec   *Vector
	score float32
}) {
	// Simple insertion sort for small arrays
	for i := 1; i < len(vectors); i++ {
		for j := i; j > 0 && vectors[j].score > vectors[j-1].score; j-- {
			vectors[j], vectors[j-1] = vectors[j-1], vectors[j]
		}
	}
}
