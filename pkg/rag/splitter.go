package rag

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// Splitter is the interface for text splitters
type Splitter interface {
	Split(text string) []string
	Name() string
}

// RecursiveCharacterSplitter splits text recursively
type RecursiveCharacterSplitter struct {
	chunkSize    int
	chunkOverlap int
	separators   []string
}

// NewRecursiveCharacterSplitter creates a new splitter
func NewRecursiveCharacterSplitter(chunkSize, chunkOverlap int) *RecursiveCharacterSplitter {
	return &RecursiveCharacterSplitter{
		chunkSize:    chunkSize,
		chunkOverlap: chunkOverlap,
		separators:   []string{"\n\n", "\n", ". ", " ", ""},
	}
}

// Split splits text into chunks
func (s *RecursiveCharacterSplitter) Split(text string) []string {
	if utf8.RuneCountInString(text) <= s.chunkSize {
		return []string{text}
	}

	// Find the best separator
	for _, sep := range s.separators {
		if !strings.Contains(text, sep) {
			continue
		}

		splits := strings.Split(text, sep)
		return s.mergeSplits(splits, sep)
	}

	// Fallback: character-level splitting
	return s.splitByChars(text)
}

// mergeSplits merges splits into chunks
func (s *RecursiveCharacterSplitter) mergeSplits(splits []string, sep string) []string {
	var chunks []string
	var currentChunk strings.Builder
	currentLen := 0

	for _, split := range splits {
		splitLen := utf8.RuneCountInString(split)
		sepLen := utf8.RuneCountInString(sep)

		if currentLen+splitLen+sepLen > s.chunkSize && currentChunk.Len() > 0 {
			chunks = append(chunks, currentChunk.String())

			// Handle overlap
			if s.chunkOverlap > 0 {
				overlapText := s.getOverlap(currentChunk.String())
				currentChunk.Reset()
				currentChunk.WriteString(overlapText)
				currentLen = utf8.RuneCountInString(overlapText)
			} else {
				currentChunk.Reset()
				currentLen = 0
			}
		}

		if currentChunk.Len() > 0 {
			currentChunk.WriteString(sep)
			currentLen += sepLen
		}
		currentChunk.WriteString(split)
		currentLen += splitLen
	}

	if currentChunk.Len() > 0 {
		chunks = append(chunks, currentChunk.String())
	}

	return chunks
}

// getOverlap gets overlap text from the end of a chunk
func (s *RecursiveCharacterSplitter) getOverlap(chunk string) string {
	runes := []rune(chunk)
	if len(runes) <= s.chunkOverlap {
		return chunk
	}

	overlap := runes[len(runes)-s.chunkOverlap:]
	// Try to start at word boundary
	for i, r := range overlap {
		if r == ' ' {
			return string(overlap[i+1:])
		}
	}
	return string(overlap)
}

// splitByChars splits by characters
func (s *RecursiveCharacterSplitter) splitByChars(text string) []string {
	runes := []rune(text)
	var chunks []string

	for i := 0; i < len(runes); i += s.chunkSize - s.chunkOverlap {
		end := i + s.chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[i:end]))
		if end >= len(runes) {
			break
		}
	}

	return chunks
}

// Name returns splitter name
func (s *RecursiveCharacterSplitter) Name() string {
	return "recursive_character"
}

// SemanticSplitter splits text by sentences
type SemanticSplitter struct {
	chunkSize    int
	chunkOverlap int
	sentenceRe   *regexp.Regexp
}

// NewSemanticSplitter creates a new semantic splitter
func NewSemanticSplitter(chunkSize, chunkOverlap int) *SemanticSplitter {
	return &SemanticSplitter{
		chunkSize:    chunkSize,
		chunkOverlap: chunkOverlap,
		sentenceRe:   regexp.MustCompile(`[.!?]+\s+`),
	}
}

// Split splits text into semantic chunks
func (s *SemanticSplitter) Split(text string) []string {
	sentences := s.splitSentences(text)
	var chunks []string
	var currentChunk strings.Builder
	currentLen := 0

	for _, sentence := range sentences {
		sentenceLen := utf8.RuneCountInString(sentence)

		if currentLen+sentenceLen+1 > s.chunkSize && currentChunk.Len() > 0 {
			chunks = append(chunks, currentChunk.String())

			if s.chunkOverlap > 0 {
				// Get overlap from last few sentences
				lastChunk := chunks[len(chunks)-1]
				overlapSentences := s.splitSentences(lastChunk)
				overlapLen := 0
				var overlapText strings.Builder

				for i := len(overlapSentences) - 1; i >= 0; i-- {
					if overlapLen >= s.chunkOverlap {
						break
					}
					overlapText.Reset()
					overlapText.WriteString(overlapSentences[i])
					overlapText.WriteString(" ")
					overlapText.WriteString(overlapText.String())
					overlapLen += utf8.RuneCountInString(overlapSentences[i])
				}

				currentChunk.Reset()
				currentChunk.WriteString(overlapText.String())
				currentLen = overlapLen
			} else {
				currentChunk.Reset()
				currentLen = 0
			}
		}

		if currentChunk.Len() > 0 {
			currentChunk.WriteString(" ")
			currentLen++
		}
		currentChunk.WriteString(sentence)
		currentLen += sentenceLen
	}

	if currentChunk.Len() > 0 {
		chunks = append(chunks, currentChunk.String())
	}

	return chunks
}

// splitSentences splits text into sentences
func (s *SemanticSplitter) splitSentences(text string) []string {
	indices := s.sentenceRe.FindAllStringIndex(text, -1)
	if len(indices) == 0 {
		return []string{text}
	}

	var sentences []string
	start := 0
	for _, idx := range indices {
		end := idx[1]
		sentences = append(sentences, strings.TrimSpace(text[start:end]))
		start = end
	}

	if start < len(text) {
		sentences = append(sentences, strings.TrimSpace(text[start:]))
	}

	return sentences
}

// Name returns splitter name
func (s *SemanticSplitter) Name() string {
	return "semantic"
}

// CodeSplitter splits code into chunks
type CodeSplitter struct {
	chunkSize    int
	chunkOverlap int
	language     string
	separators   []string
}

// NewCodeSplitter creates a new code splitter
func NewCodeSplitter(language string) *CodeSplitter {
	s := &CodeSplitter{
		chunkSize:    1500,
		chunkOverlap: 200,
		language:     language,
	}

	// Language-specific separators
	switch language {
	case "go":
		s.separators = []string{"\nfunc ", "\ntype ", "\nvar ", "\nconst ", "\n\n", "\n", " "}
	case "python":
		s.separators = []string{"\nclass ", "\ndef ", "\n\tdef ", "\n\n", "\n", " "}
	case "javascript", "typescript":
		s.separators = []string{"\nfunction ", "\nclass ", "\nexport ", "\nconst ", "\n\n", "\n", " "}
	default:
		s.separators = []string{"\n\n", "\n", " "}
	}

	return s
}

// Split splits code into chunks
func (s *CodeSplitter) Split(text string) []string {
	fallback := NewRecursiveCharacterSplitter(s.chunkSize, s.chunkOverlap)
	fallback.separators = s.separators
	return fallback.Split(text)
}

// Name returns splitter name
func (s *CodeSplitter) Name() string {
	return "code_" + s.language
}
