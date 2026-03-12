// Package rag provides Retrieval-Augmented Generation functionality.
package rag

import (
	"strings"
	"unicode"
)

// Chunker splits text into overlapping chunks.
type Chunker struct {
	chunkSize int
	overlap   int
}

// NewChunker creates a new chunker with the specified size and overlap.
func NewChunker(chunkSize, overlap int) *Chunker {
	if chunkSize <= 0 {
		chunkSize = 512
	}
	if overlap < 0 {
		overlap = 0
	}
	if overlap >= chunkSize {
		overlap = chunkSize / 4
	}
	return &Chunker{
		chunkSize: chunkSize,
		overlap:   overlap,
	}
}

// Chunk splits text into overlapping chunks.
func (c *Chunker) Chunk(text string) []string {
	if text == "" {
		return nil
	}

	// If text is smaller than chunk size, return as single chunk
	if len(text) <= c.chunkSize {
		return []string{text}
	}

	var chunks []string
	start := 0

	for start < len(text) {
		end := start + c.chunkSize
		if end > len(text) {
			end = len(text)
		}

		// Try to break at a sentence boundary
		if end < len(text) {
			// Look for sentence ending punctuation followed by space or newline
			for i := end - 1; i > start+c.overlap; i-- {
				if (text[i] == '.' || text[i] == '!' || text[i] == '?') &&
					(i+1 < len(text) && (text[i+1] == ' ' || text[i+1] == '\n')) {
					end = i + 1
					break
				}
			}

			// If no sentence boundary found, try word boundary
			if end == start+c.chunkSize {
				for i := end - 1; i > start+c.overlap; i-- {
					if unicode.IsSpace(rune(text[i])) {
						end = i
						break
					}
				}
			}
		}

		chunk := strings.TrimSpace(text[start:end])
		if chunk != "" {
			chunks = append(chunks, chunk)
		}

		// Move start position, accounting for overlap
		start = end - c.overlap
		if start <= 0 || start >= len(text) {
			break
		}
	}

	return chunks
}

// ChunkSize returns the chunk size.
func (c *Chunker) ChunkSize() int {
	return c.chunkSize
}

// Overlap returns the overlap size.
func (c *Chunker) Overlap() int {
	return c.overlap
}

// ChunkBySentences splits text into chunks at sentence boundaries.
// This is a convenience method that creates sentence-aligned chunks.
func (c *Chunker) ChunkBySentences(text string) []string {
	if text == "" {
		return nil
	}

	// Split into sentences
	sentences := splitSentences(text)
	if len(sentences) == 0 {
		return []string{text}
	}

	var chunks []string
	var currentChunk strings.Builder

	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if sentence == "" {
			continue
		}

		// If adding this sentence would exceed chunk size, save current chunk
		if currentChunk.Len() > 0 && currentChunk.Len()+len(sentence)+1 > c.chunkSize {
			chunks = append(chunks, strings.TrimSpace(currentChunk.String()))
			// Start new chunk with overlap (keep last part of previous chunk)
			overlapStart := max(0, currentChunk.Len()-c.overlap)
			overlapText := currentChunk.String()[overlapStart:]
			currentChunk.Reset()
			currentChunk.WriteString(overlapText)
		}

		if currentChunk.Len() > 0 {
			currentChunk.WriteString(" ")
		}
		currentChunk.WriteString(sentence)
	}

	// Don't forget the last chunk
	if currentChunk.Len() > 0 {
		chunks = append(chunks, strings.TrimSpace(currentChunk.String()))
	}

	return chunks
}

// splitSentences splits text into sentences.
func splitSentences(text string) []string {
	var sentences []string
	var current strings.Builder

	runes := []rune(text)
	for i, r := range runes {
		current.WriteRune(r)

		// Check for sentence ending
		if r == '.' || r == '!' || r == '?' {
			// Check if next char is space or end of text
			if i+1 >= len(runes) || unicode.IsSpace(runes[i+1]) {
				sentences = append(sentences, current.String())
				current.Reset()
			}
		}
	}

	// Add any remaining text
	if current.Len() > 0 {
		sentences = append(sentences, current.String())
	}

	return sentences
}

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
