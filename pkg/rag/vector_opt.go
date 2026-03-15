package rag

import (
	"sync"
	"sync/atomic"
)

// OptimizedVectorStore is a memory-optimized vector store
type OptimizedVectorStore struct {
	// Use separate slices instead of struct for better memory layout
	ids       []string
	vectors   [][]float32
	metadata  []map[string]string
	
	// Index for fast lookups
	idToIndex map[string]int
	
	// Pool for temporary vectors
	vectorPool sync.Pool
	
	// Stats (atomic for thread-safety)
	count int64
	
	mu sync.RWMutex
}

// NewOptimizedVectorStore creates an optimized vector store
func NewOptimizedVectorStore(initialCapacity int) *OptimizedVectorStore {
	return &OptimizedVectorStore{
		ids:       make([]string, 0, initialCapacity),
		vectors:   make([][]float32, 0, initialCapacity),
		metadata:  make([]map[string]string, 0, initialCapacity),
		idToIndex: make(map[string]int, initialCapacity),
		vectorPool: sync.Pool{
			New: func() interface{} {
				return make([]float32, 384) // Typical embedding size
			},
		},
	}
}

// Add adds a vector with metadata
func (s *OptimizedVectorStore) Add(id string, vector []float32, metadata map[string]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if already exists
	if _, exists := s.idToIndex[id]; exists {
		return nil // Or update existing
	}

	// Get vector from pool and copy
	pooledVector := s.vectorPool.Get().([]float32)
	if len(pooledVector) < len(vector) {
		pooledVector = make([]float32, len(vector))
	}
	copy(pooledVector, vector)

	// Append to slices
	index := len(s.ids)
	s.ids = append(s.ids, id)
	s.vectors = append(s.vectors, pooledVector[:len(vector)])
	s.metadata = append(s.metadata, metadata)
	s.idToIndex[id] = index

	atomic.AddInt64(&s.count, 1)
	return nil
}

// Search performs similarity search
func (s *OptimizedVectorStore) Search(query []float32, topK int) ([]SearchResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.vectors) == 0 {
		return nil, nil
	}

	// Pre-allocate results
	type scoreIndex struct {
		score float64
		index int
	}
	scores := make([]scoreIndex, 0, len(s.vectors))

	// Calculate similarities
	for i, vec := range s.vectors {
		score := cosineSimilarity(query, vec)
		if score > 0 {
			scores = append(scores, scoreIndex{score: score, index: i})
		}
	}

	// Sort by score (descending)
	// Use a simple insertion sort for small topK
	if topK < 100 && len(scores) > topK {
		// Partial sort - only need top K
		for i := 0; i < topK && i < len(scores); i++ {
			maxIdx := i
			for j := i + 1; j < len(scores); j++ {
				if scores[j].score > scores[maxIdx].score {
					maxIdx = j
				}
			}
			scores[i], scores[maxIdx] = scores[maxIdx], scores[i]
		}
		scores = scores[:topK]
	} else {
		// Full sort for larger topK
		sortScores(scores)
		if len(scores) > topK {
			scores = scores[:topK]
		}
	}

	// Build results
	results := make([]SearchResult, len(scores))
	for i, s := range scores {
		results[i] = SearchResult{
			ID:       s.ids[s.index],
			Score:    s.score,
			Metadata: s.metadata[s.index],
		}
	}

	return results, nil
}

// Delete removes a vector
func (s *OptimizedVectorStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	index, exists := s.idToIndex[id]
	if !exists {
		return nil
	}

	// Return vector to pool
	s.vectorPool.Put(s.vectors[index])

	// Remove from slices (swap with last)
	lastIndex := len(s.ids) - 1
	if index != lastIndex {
		// Swap
		s.ids[index] = s.ids[lastIndex]
		s.vectors[index] = s.vectors[lastIndex]
		s.metadata[index] = s.metadata[lastIndex]
		s.idToIndex[s.ids[index]] = index
	}

	// Truncate
	s.ids = s.ids[:lastIndex]
	s.vectors = s.vectors[:lastIndex]
	s.metadata = s.metadata[:lastIndex]
	delete(s.idToIndex, id)

	atomic.AddInt64(&s.count, -1)
	return nil
}

// Count returns the number of vectors
func (s *OptimizedVectorStore) Count() int {
	return int(atomic.LoadInt64(&s.count))
}

// Clear removes all vectors
func (s *OptimizedVectorStore) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Return all vectors to pool
	for _, vec := range s.vectors {
		s.vectorPool.Put(vec)
	}

	// Reset
	s.ids = s.ids[:0]
	s.vectors = s.vectors[:0]
	s.metadata = s.metadata[:0]
	s.idToIndex = make(map[string]int)

	atomic.StoreInt64(&s.count, 0)
	return nil
}

// Get retrieves a vector by ID
func (s *OptimizedVectorStore) Get(id string) ([]float32, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	index, exists := s.idToIndex[id]
	if !exists {
		return nil, nil
	}

	// Return a copy to avoid mutation
	vec := s.vectors[index]
	result := make([]float32, len(vec))
	copy(result, vec)
	return result, nil
}

// Helper: simple sort for scores
func sortScores(scores []struct {
	score float64
	index int
}) {
	// Quick sort implementation for better performance
	if len(scores) < 2 {
		return
	}

	// Use built-in sort for simplicity (or implement quicksort for better perf)
	for i := 0; i < len(scores)-1; i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[j].score > scores[i].score {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}
}
