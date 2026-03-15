package rag

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Pipeline represents a RAG pipeline
type Pipeline struct {
	config     Config
	store      VectorStore
	splitter   Splitter
	embedder   *EmbeddingEngine
	stats      Stats
	mu         sync.RWMutex
}

// NewPipeline creates a new RAG pipeline
func NewPipeline(config Config) *Pipeline {
	var splitter Splitter
	switch {
	case strings.Contains(config.ChunkSize, "semantic"):
		splitter = NewSemanticSplitter(config.ChunkSize, config.ChunkOverlap)
	default:
		splitter = NewRecursiveCharacterSplitter(config.ChunkSize, config.ChunkOverlap)
	}

	return &Pipeline{
		config:   config,
		store:    NewInMemoryVectorStore(),
		splitter: splitter,
		embedder: NewEmbeddingEngine(NewPseudoEmbeddingProvider(384), 384),
		stats:    Stats{},
	}
}

// Index indexes a document
func (p *Pipeline) Index(ctx context.Context, doc *Document) error {
	// Split document into chunks
	chunks := doc.Chunk(p.config.ChunkSize, p.config.ChunkOverlap)

	// Generate embeddings for chunks
	texts := make([]string, len(chunks))
	for i, chunk := range chunks {
		texts[i] = chunk.Content
	}

	embeddings, err := p.embedder.Embed(ctx, texts)
	if err != nil {
		return fmt.Errorf("failed to generate embeddings: %w", err)
	}

	// Store vectors with metadata
	for i, chunk := range chunks {
		metadata := map[string]string{
			"document_id": chunk.DocumentID,
			"chunk_index": fmt.Sprintf("%d", chunk.ChunkIndex),
			"source":      chunk.Metadata.Source,
			"title":       chunk.Metadata.Title,
		}

		id := fmt.Sprintf("%s_%d", chunk.DocumentID, chunk.ChunkIndex)
		if err := p.store.Add(id, embeddings[i], metadata); err != nil {
			return fmt.Errorf("failed to store chunk: %w", err)
		}
	}

	// Update stats
	p.mu.Lock()
	p.stats.TotalDocuments++
	p.stats.TotalChunks += len(chunks)
	p.stats.LastIndexed = time.Now().Format(time.RFC3339)
	p.mu.Unlock()

	return nil
}

// IndexBatch indexes multiple documents
func (p *Pipeline) IndexBatch(ctx context.Context, docs []*Document) error {
	for _, doc := range docs {
		if err := p.Index(ctx, doc); err != nil {
			return err
		}
	}
	return nil
}

// Retrieve retrieves relevant documents
func (p *Pipeline) Retrieve(ctx context.Context, query string) (*Response, error) {
	// Generate query embedding
	queryEmbedding, err := p.embedder.EmbedOne(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	// Search vector store
	results, err := p.store.Search(queryEmbedding, p.config.TopK)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}

	// Filter by minimum score
	var filteredResults []RetrievalResult
	for _, result := range results {
		if result.Score >= p.config.MinScore {
			filteredResults = append(filteredResults, RetrievalResult{
				Content:    "", // Will be populated from storage
				Score:      result.Score,
				DocumentID: result.Metadata["document_id"],
				Metadata: DocumentMetadata{
					Source: result.Metadata["source"],
					Title:  result.Metadata["title"],
				},
			})
		}
	}

	p.mu.RLock()
	stats := p.stats
	p.mu.RUnlock()

	return &Response{
		Query:   query,
		Results: filteredResults,
		Stats:   stats,
	}, nil
}

// Stats returns pipeline statistics
func (p *Pipeline) Stats() Stats {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.stats
}

// Clear clears all indexed data
func (p *Pipeline) Clear() error {
	if err := p.store.Clear(); err != nil {
		return err
	}

	p.mu.Lock()
	p.stats = Stats{}
	p.mu.Unlock()

	return nil
}

// Response represents a RAG response
type Response struct {
	Query   string            `json:"query"`
	Results []RetrievalResult `json:"results"`
	Stats   Stats             `json:"stats"`
}

// Context returns combined context from results
func (r *Response) Context() string {
	var contexts []string
	for _, result := range r.Results {
		contexts = append(contexts, result.Content)
	}
	return strings.Join(contexts, "\n\n")
}

// Sources returns unique sources
func (r *Response) Sources() []string {
	sourceSet := make(map[string]bool)
	for _, result := range r.Results {
		if result.Metadata.Source != "" {
			sourceSet[result.Metadata.Source] = true
		}
	}

	sources := make([]string, 0, len(sourceSet))
	for source := range sourceSet {
		sources = append(sources, source)
	}
	return sources
}

// Import strings for Context
import "strings"
