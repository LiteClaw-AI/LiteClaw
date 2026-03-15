package main

import (
	"context"
	"fmt"
	"log"

	"github.com/liteclaw/liteclaw/pkg/provider"
	"github.com/liteclaw/liteclaw/pkg/rag"
)

func main() {
	fmt.Println("=== LiteClaw Go Example ===\n")

	// Example 1: Provider Usage
	exampleProviderUsage()

	// Example 2: RAG System
	exampleRAGSystem()

	// Example 3: Document Processing
	exampleDocumentProcessing()
}

func exampleProviderUsage() {
	fmt.Println("📝 Example 1: Provider Usage\n")

	// Create registry
	registry := provider.NewRegistry()

	// Register providers
	if apiKey := getEnv("OPENAI_API_KEY"); apiKey != "" {
		registry.Register("openai", provider.NewOpenAI(apiKey))
		fmt.Println("✓ OpenAI provider registered")
	}

	if apiKey := getEnv("ANTHROPIC_API_KEY"); apiKey != "" {
		registry.Register("anthropic", provider.NewAnthropic(apiKey))
		fmt.Println("✓ Anthropic provider registered")
	}

	if apiKey := getEnv("DASHSCOPE_API_KEY"); apiKey != "" {
		registry.Register("qwen", provider.NewAliyunQwen(apiKey))
		fmt.Println("✓ Aliyun Qwen provider registered")
	}

	if apiKey := getEnv("DEEPSEEK_API_KEY"); apiKey != "" {
		registry.Register("deepseek", provider.NewDeepSeek(apiKey))
		fmt.Println("✓ DeepSeek provider registered")
	}

	// List providers
	fmt.Printf("\nRegistered providers: %v\n\n", registry.List())

	// Use provider
	if p, err := registry.Get("openai"); err == nil {
		fmt.Printf("Using: %s\n", p.Metadata().Name)
		fmt.Printf("Default model: %s\n", p.Metadata().DefaultModel)
		fmt.Printf("Capabilities: %v\n\n", p.Metadata().Capabilities)
	}
}

func exampleRAGSystem() {
	fmt.Println("📚 Example 2: RAG System\n")

	// Create pipeline with config
	config := rag.Config{
		ChunkSize:    1000,
		ChunkOverlap: 200,
		TopK:         5,
		MinScore:     0.5,
		HybridSearch: true,
	}

	pipeline := rag.NewPipeline(config)
	ctx := context.Background()

	// Index documents
	docs := []*rag.Document{
		rag.NewDocument("Rust is a systems programming language focused on safety and performance.").
			WithTitle("Rust Overview").
			WithSource("rust.md"),
		rag.NewDocument("Python is a high-level language known for simplicity and readability.").
			WithTitle("Python Overview").
			WithSource("python.md"),
		rag.NewDocument("Go is designed for simplicity, concurrency, and fast compilation.").
			WithTitle("Go Overview").
			WithSource("go.md"),
	}

	for _, doc := range docs {
		if err := pipeline.Index(ctx, doc); err != nil {
			log.Printf("Failed to index: %v", err)
		}
	}

	stats := pipeline.Stats()
	fmt.Printf("Indexed %d documents with %d chunks\n\n",
		stats.TotalDocuments, stats.TotalChunks)

	// Query
	query := "Which language is good for systems programming?"
	response, err := pipeline.Retrieve(ctx, query)
	if err != nil {
		log.Printf("Query failed: %v", err)
		return
	}

	fmt.Printf("Query: %s\n", query)
	fmt.Printf("Found %d results:\n\n", len(response.Results))

	for i, result := range response.Results {
		fmt.Printf("%d. [%s] Score: %.3f\n",
			i+1, result.Metadata.Source, result.Score)
	}

	fmt.Println()
}

func exampleDocumentProcessing() {
	fmt.Println("📄 Example 3: Document Processing\n")

	// Create document
	doc := rag.NewDocument(`
# Introduction to Machine Learning

Machine learning is a subset of artificial intelligence that enables 
systems to learn from data without explicit programming.

## Types of Machine Learning

1. Supervised Learning
2. Unsupervised Learning
3. Reinforcement Learning

### Supervised Learning

Supervised learning uses labeled data to train models.
Common algorithms include linear regression and decision trees.

### Unsupervised Learning

Unsupervised learning finds patterns in unlabeled data.
Clustering and dimensionality reduction are common techniques.
	`).WithTitle("ML Introduction").WithSource("ml.md")

	fmt.Printf("Document: %s\n", doc.Metadata.Title)
	fmt.Printf("Content length: %d chars\n", len(doc.Content))
	fmt.Printf("Word count: %d\n\n", len(splitWords(doc.Content)))

	// Chunk document
	chunks := doc.Chunk(200, 50)
	fmt.Printf("Split into %d chunks:\n", len(chunks))

	for i, chunk := range chunks {
		fmt.Printf("  Chunk %d: %d chars\n", i, len(chunk.Content))
	}

	fmt.Println()

	// Use different splitters
	fmt.Println("Text Splitters:")
	fmt.Printf("  - RecursiveCharacter: chunk_size=1000, overlap=200\n")
	fmt.Printf("  - Semantic: sentence-based chunking\n")
	fmt.Printf("  - Code: language-aware splitting (Go, Python, JS)\n")

	fmt.Println()
}

// Helpers
func getEnv(key string) string {
	// In real code, use os.Getenv
	return ""
}

func splitWords(s string) []string {
	var words []string
	wordStart := -1

	for i, r := range s {
		if isSpace(r) {
			if wordStart != -1 {
				words = append(words, s[wordStart:i])
				wordStart = -1
			}
		} else {
			if wordStart == -1 {
				wordStart = i
			}
		}
	}

	if wordStart != -1 {
		words = append(words, s[wordStart:])
	}

	return words
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}
