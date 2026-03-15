// Package rag implements Retrieval-Augmented Generation
package rag

import (
	"time"
)

// Document represents a document with content and metadata
type Document struct {
	ID        string            `json:"id"`
	Content   string            `json:"content"`
	Metadata  DocumentMetadata  `json:"metadata"`
	Embedding []float32         `json:"embedding,omitempty"`
}

// DocumentMetadata represents document metadata
type DocumentMetadata struct {
	Source     string            `json:"source,omitempty"`
	Title      string            `json:"title,omitempty"`
	Author     string            `json:"author,omitempty"`
	CreatedAt  string            `json:"created_at,omitempty"`
	ModifiedAt string            `json:"modified_at,omitempty"`
	DocType    string            `json:"doc_type,omitempty"`
	Language   string            `json:"language,omitempty"`
	Custom     map[string]string `json:"custom,omitempty"`
}

// DocumentChunk represents a document chunk
type DocumentChunk struct {
	DocumentID string           `json:"document_id"`
	ChunkIndex int              `json:"chunk_index"`
	Content    string           `json:"content"`
	Metadata   DocumentMetadata `json:"metadata"`
}

// NewDocument creates a new document
func NewDocument(content string) *Document {
	return &Document{
		ID:       generateID(),
		Content:  content,
		Metadata: DocumentMetadata{CreatedAt: time.Now().Format(time.RFC3339)},
	}
}

// WithSource sets the document source
func (d *Document) WithSource(source string) *Document {
	d.Metadata.Source = source
	return d
}

// WithTitle sets the document title
func (d *Document) WithTitle(title string) *Document {
	d.Metadata.Title = title
	return d
}

// WithMetadata adds custom metadata
func (d *Document) WithMetadata(key, value string) *Document {
	if d.Metadata.Custom == nil {
		d.Metadata.Custom = make(map[string]string)
	}
	d.Metadata.Custom[key] = value
	return d
}

// Chunk splits the document into chunks
func (d *Document) Chunk(chunkSize, overlap int) []DocumentChunk {
	var chunks []DocumentChunk
	content := d.Content
	start := 0
	chunkIndex := 0

	for start < len(content) {
		end := start + chunkSize
		if end > len(content) {
			end = len(content)
		}

		chunkContent := content[start:end]
		chunks = append(chunks, DocumentChunk{
			DocumentID: d.ID,
			ChunkIndex: chunkIndex,
			Content:    chunkContent,
			Metadata:   d.Metadata,
		})

		chunkIndex++
		if end >= len(content) {
			break
		}

		// Move start with overlap
		start = end - overlap
		if start < 0 {
			start = 0
		}
	}

	return chunks
}

// Stats represents RAG system statistics
type Stats struct {
	TotalDocuments    int     `json:"total_documents"`
	TotalChunks       int     `json:"total_chunks"`
	AvgChunkSize      float64 `json:"avg_chunk_size"`
	EmbeddingDim      int     `json:"embedding_dimension"`
	LastIndexed       string  `json:"last_indexed,omitempty"`
	VectorStoreSize   int     `json:"vector_store_size"`
}

// RetrievalResult represents a retrieval result
type RetrievalResult struct {
	Content    string           `json:"content"`
	Score      float64          `json:"score"`
	DocumentID string           `json:"document_id"`
	ChunkIndex int              `json:"chunk_index"`
	Metadata   DocumentMetadata `json:"metadata"`
}

// Config represents RAG configuration
type Config struct {
	EmbeddingModel string  `yaml:"embedding_model" json:"embedding_model"`
	ChunkSize      int     `yaml:"chunk_size" json:"chunk_size"`
	ChunkOverlap   int     `yaml:"chunk_overlap" json:"chunk_overlap"`
	TopK           int     `yaml:"top_k" json:"top_k"`
	MinScore       float64 `yaml:"min_score" json:"min_score"`
	HybridSearch   bool    `yaml:"hybrid_search" json:"hybrid_search"`
	VectorStore    string  `yaml:"vector_store" json:"vector_store"`
}

// DefaultConfig returns default RAG configuration
func DefaultConfig() Config {
	return Config{
		EmbeddingModel: "text-embedding-3-small",
		ChunkSize:      1000,
		ChunkOverlap:   200,
		TopK:           5,
		MinScore:       0.5,
		HybridSearch:   true,
		VectorStore:    "memory",
	}
}

// generateID generates a unique ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// Import fmt for ID generation
import "fmt"
