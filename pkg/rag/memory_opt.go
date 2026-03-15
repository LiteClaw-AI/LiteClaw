package rag

import (
	"sync"
)

// OptimizedDocumentPool pools Document objects to reduce allocations
var OptimizedDocumentPool = sync.Pool{
	New: func() interface{} {
		return &Document{
			Metadata: DocumentMetadata{
				Custom: make(map[string]string, 4), // Pre-allocate small map
			},
		}
	},
}

// OptimizedChunkPool pools chunk slices
var OptimizedChunkPool = sync.Pool{
	New: func() interface{} {
		// Pre-allocate typical chunk size
		return make([]byte, 0, 1024)
	},
}

// GetDocument gets a document from the pool
func GetDocument() *Document {
	return OptimizedDocumentPool.Get().(*Document)
}

// PutDocument returns a document to the pool
func PutDocument(doc *Document) {
	// Reset document
	doc.ID = ""
	doc.Content = ""
	doc.Embedding = nil
	doc.Metadata = DocumentMetadata{
		Custom: make(map[string]string, 4),
	}
	OptimizedDocumentPool.Put(doc)
}

// Interned strings for common metadata keys
const (
	MetaSource     = "source"
	MetaTitle      = "title"
	MetaAuthor     = "author"
	MetaDocType    = "doc_type"
	MetaLanguage   = "language"
	MetaExtension  = "extension"
	MetaType       = "type"
	MetaFileCount  = "file_count"
)

// Interned string map for reducing string allocations
var internedStrings = struct {
	sync.RWMutex
	cache map[string]string
}{
	cache: make(map[string]string, 256),
}

// InternString returns an interned copy of the string
func InternString(s string) string {
	if len(s) == 0 {
		return s
	}

	internedStrings.RLock()
	if cached, exists := internedStrings.cache[s]; exists {
		internedStrings.RUnlock()
		return cached
	}
	internedStrings.RUnlock()

	internedStrings.Lock()
	internedStrings.cache[s] = s
	internedStrings.Unlock()

	return s
}

// OptimizedDocumentCollection uses pre-allocation
type OptimizedDocumentCollection struct {
	docs     []*Document
	docPool  *sync.Pool
	capacity int
}

// NewOptimizedDocumentCollection creates a collection with pre-allocated capacity
func NewOptimizedDocumentCollection(capacity int) *OptimizedDocumentCollection {
	return &OptimizedDocumentCollection{
		docs:     make([]*Document, 0, capacity),
		docPool:  &OptimizedDocumentPool,
		capacity: capacity,
	}
}

// Add adds a document from the pool
func (c *OptimizedDocumentCollection) Add(content string) *Document {
	doc := c.docPool.Get().(*Document)
	doc.Content = content
	c.docs = append(c.docs, doc)
	return doc
}

// Release returns all documents to the pool
func (c *OptimizedDocumentCollection) Release() {
	for _, doc := range c.docs {
		PutDocument(doc)
	}
	c.docs = c.docs[:0]
}

// OptimizedChunkResult represents a chunk with minimal allocations
type OptimizedChunkResult struct {
	Content    string
	Score      float64
	DocumentID string
	ChunkIndex int
	Source     string
	Title      string
}

// ToRetrievalResult converts to standard RetrievalResult
func (r *OptimizedChunkResult) ToRetrievalResult() RetrievalResult {
	return RetrievalResult{
		Content:    r.Content,
		Score:      r.Score,
		DocumentID: r.DocumentID,
		ChunkIndex: r.ChunkIndex,
		Metadata: DocumentMetadata{
			Source: r.Source,
			Title:  r.Title,
		},
	}
}
