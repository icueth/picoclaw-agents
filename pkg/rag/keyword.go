// Package rag provides Retrieval-Augmented Generation functionality.
package rag

import (
	"database/sql"
	"fmt"
	"strings"
)

// KeywordSearchResult represents a result from FTS5 keyword search.
type KeywordSearchResult struct {
	ID       string
	Content  string
	Metadata string
	Score    float32 // BM25 score
}

// KeywordSearcher provides FTS5-based keyword search.
type KeywordSearcher struct {
	db *sql.DB
}

// NewKeywordSearcher creates a new keyword searcher.
func NewKeywordSearcher(db *sql.DB) *KeywordSearcher {
	return &KeywordSearcher{db: db}
}

// Search performs BM25 keyword search using FTS5.
func (k *KeywordSearcher) Search(query string, limit int) ([]KeywordSearchResult, error) {
	if query == "" || limit <= 0 {
		return nil, nil
	}

	// Escape FTS5 special characters and prepare query
	ftsQuery := prepareFTS5Query(query)

	// Use FTS5 bm25() for ranking
	sqlQuery := `
		SELECT d.id, d.content, d.metadata, bm25(rag_documents_fts) as score
		FROM rag_documents_fts
		JOIN rag_documents d ON d.rowid = rag_documents_fts.rowid
		WHERE rag_documents_fts MATCH ?
		ORDER BY score ASC
		LIMIT ?
	`

	rows, err := k.db.Query(sqlQuery, ftsQuery, limit)
	if err != nil {
		return nil, fmt.Errorf("keyword search failed: %w", err)
	}
	defer rows.Close()

	var results []KeywordSearchResult
	for rows.Next() {
		var r KeywordSearchResult
		var score float64
		if err := rows.Scan(&r.ID, &r.Content, &r.Metadata, &score); err != nil {
			continue
		}
		// Convert BM25 score (lower is better) to our score format (higher is better, 0-1 range)
		// BM25 scores are typically negative, closer to 0 is better
		r.Score = bm25ToScore(score)
		results = append(results, r)
	}

	return results, rows.Err()
}

// prepareFTS5Query escapes special characters and prepares query for FTS5.
func prepareFTS5Query(query string) string {
	// FTS5 special characters: " * ( ) - ^
	// Replace with spaces and split into tokens
	replacer := strings.NewReplacer(
		"\"", " ",
		"*", " ",
		"(", " ",
		")", " ",
		"-", " ",
		"^", " ",
	)
	cleaned := replacer.Replace(query)
	
	// Split into words and add * for prefix matching
	words := strings.Fields(cleaned)
	if len(words) == 0 {
		return cleaned
	}
	
	// Add prefix matching to each word
	for i, word := range words {
		words[i] = word + "*"
	}
	
	return strings.Join(words, " ")
}

// bm25ToScore converts BM25 score to 0-1 range (higher is better).
// BM25 returns negative values where closer to 0 is better.
func bm25ToScore(bm25 float64) float32 {
	if bm25 >= 0 {
		return 1.0
	}
	// Convert negative BM25 to positive score
	// Typical BM25 range is -10 to 0
	score := 1.0 + (bm25 / 10.0)
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}
	return float32(score)
}

// IsFTSAvailable checks if FTS5 is available in the database.
func IsFTSAvailable(db *sql.DB) bool {
	var name string
	err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='rag_documents_fts'").Scan(&name)
	return err == nil && name == "rag_documents_fts"
}

// Reindex rebuilds the FTS5 index.
func (k *KeywordSearcher) Reindex() error {
	// Delete all from FTS5
	_, err := k.db.Exec("DELETE FROM rag_documents_fts")
	if err != nil {
		return fmt.Errorf("failed to clear FTS5 index: %w", err)
	}

	// Rebuild from rag_documents
	_, err = k.db.Exec(`
		INSERT INTO rag_documents_fts(rowid, content, metadata)
		SELECT rowid, content, metadata FROM rag_documents
	`)
	if err != nil {
		return fmt.Errorf("failed to rebuild FTS5 index: %w", err)
	}

	return nil
}
