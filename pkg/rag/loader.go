package rag

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Loader is the interface for document loaders
type Loader interface {
	Load(ctx context.Context, source string) (*Document, error)
	LoadBatch(ctx context.Context, sources []string) ([]*Document, error)
	Supports(source string) bool
}

// TextLoader loads raw text
type TextLoader struct{}

// NewTextLoader creates a new text loader
func NewTextLoader() *TextLoader {
	return &TextLoader{}
}

// Load loads text content
func (l *TextLoader) Load(ctx context.Context, source string) (*Document, error) {
	return NewDocument(source), nil
}

// LoadBatch loads multiple texts
func (l *TextLoader) LoadBatch(ctx context.Context, sources []string) ([]*Document, error) {
	docs := make([]*Document, len(sources))
	for i, source := range sources {
		doc, err := l.Load(ctx, source)
		if err != nil {
			return nil, err
		}
		docs[i] = doc
	}
	return docs, nil
}

// Supports always returns true for text loader
func (l *TextLoader) Supports(source string) bool {
	return true
}

// FileLoader loads documents from files
type FileLoader struct {
	supportedExts map[string]bool
}

// NewFileLoader creates a new file loader
func NewFileLoader() *FileLoader {
	return &FileLoader{
		supportedExts: map[string]bool{
			".txt":      true,
			".md":       true,
			".markdown": true,
			".rs":       true,
			".py":       true,
			".js":       true,
			".ts":       true,
			".go":       true,
			".java":     true,
			".json":     true,
			".yaml":     true,
			".yml":      true,
			".toml":     true,
		},
	}
}

// Load loads a file
func (l *FileLoader) Load(ctx context.Context, source string) (*Document, error) {
	file, err := os.Open(source)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	ext := filepath.Ext(source)
	doc := NewDocument(string(content)).
		WithSource(source).
		WithMetadata("extension", ext)

	// Set title from filename
	if filename := filepath.Base(source); filename != "" {
		doc = doc.WithTitle(strings.TrimSuffix(filename, ext))
	}

	return doc, nil
}

// LoadBatch loads multiple files
func (l *FileLoader) LoadBatch(ctx context.Context, sources []string) ([]*Document, error) {
	docs := make([]*Document, 0, len(sources))
	for _, source := range sources {
		doc, err := l.Load(ctx, source)
		if err != nil {
			return nil, fmt.Errorf("failed to load %s: %w", source, err)
		}
		docs = append(docs, doc)
	}
	return docs, nil
}

// Supports checks if file extension is supported
func (l *FileLoader) Supports(source string) bool {
	ext := strings.ToLower(filepath.Ext(source))
	return l.supportedExts[ext]
}

// URLLoader loads documents from URLs
type URLLoader struct {
	client *http.Client
}

// NewURLLoader creates a new URL loader
func NewURLLoader() *URLLoader {
	return &URLLoader{
		client: &http.Client{},
	}
}

// Load fetches content from URL
func (l *URLLoader) Load(ctx context.Context, source string) (*Document, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", source, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := l.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Simple HTML to text conversion
	text := l.htmlToText(string(content))

	return NewDocument(text).
		WithSource(source).
		WithMetadata("type", "web"), nil
}

// LoadBatch loads multiple URLs
func (l *URLLoader) LoadBatch(ctx context.Context, sources []string) ([]*Document, error) {
	docs := make([]*Document, 0, len(sources))
	for _, source := range sources {
		doc, err := l.Load(ctx, source)
		if err != nil {
			return nil, fmt.Errorf("failed to load %s: %w", source, err)
		}
		docs = append(docs, doc)
	}
	return docs, nil
}

// Supports checks if source is a URL
func (l *URLLoader) Supports(source string) bool {
	return strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://")
}

// htmlToText converts HTML to plain text
func (l *URLLoader) htmlToText(html string) string {
	// Remove HTML tags
	lines := bufio.NewScanner(strings.NewReader(html))
	var text strings.Builder
	for lines.Scan() {
		line := lines.Text()
		// Simple tag removal
		if !strings.HasPrefix(strings.TrimSpace(line), "<") {
			text.WriteString(line)
			text.WriteString("\n")
		}
	}
	return text.String()
}

// DirectoryLoader loads all files from a directory
type DirectoryLoader struct {
	fileLoader *FileLoader
	recursive  bool
}

// NewDirectoryLoader creates a new directory loader
func NewDirectoryLoader(recursive bool) *DirectoryLoader {
	return &DirectoryLoader{
		fileLoader: NewFileLoader(),
		recursive:  recursive,
	}
}

// Load loads all files from directory
func (l *DirectoryLoader) Load(ctx context.Context, source string) (*Document, error) {
	files, err := l.listFiles(source)
	if err != nil {
		return nil, err
	}

	docs, err := l.fileLoader.LoadBatch(ctx, files)
	if err != nil {
		return nil, err
	}

	// Combine all documents
	var combined strings.Builder
	for _, doc := range docs {
		source := doc.Metadata.Source
		if source == "" {
			source = "unknown"
		}
		combined.WriteString("--- ")
		combined.WriteString(source)
		combined.WriteString(" ---\n")
		combined.WriteString(doc.Content)
		combined.WriteString("\n\n")
	}

	return NewDocument(combined.String()).
		WithSource(source).
		WithMetadata("type", "directory").
		WithMetadata("file_count", fmt.Sprintf("%d", len(docs))), nil
}

// LoadBatch loads multiple directories
func (l *DirectoryLoader) LoadBatch(ctx context.Context, sources []string) ([]*Document, error) {
	docs := make([]*Document, 0, len(sources))
	for _, source := range sources {
		doc, err := l.Load(ctx, source)
		if err != nil {
			return nil, err
		}
		docs = append(docs, doc)
	}
	return docs, nil
}

// Supports checks if source is a directory
func (l *DirectoryLoader) Supports(source string) bool {
	info, err := os.Stat(source)
	return err == nil && info.IsDir()
}

// listFiles lists all supported files in directory
func (l *DirectoryLoader) listFiles(dir string) ([]string, error) {
	var files []string

	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && l.fileLoader.Supports(path) {
			files = append(files, path)
		}
		if info.IsDir() && !l.recursive && path != dir {
			return filepath.SkipDir
		}
		return nil
	}

	if err := filepath.Walk(dir, walkFn); err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return files, nil
}
