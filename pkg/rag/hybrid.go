// Package rag provides Retrieval-Augmented Generation functionality.
package rag

import (
	"sort"
)

// HybridResult combines vector and keyword search results.
type HybridResult struct {
	SearchResult
	VectorScore  float32
	KeywordScore float32
}

// HybridConfig configures hybrid search behavior.
type HybridConfig struct {
	VectorWeight  float32 // Weight for vector similarity (default 0.7)
	KeywordWeight float32 // Weight for keyword/BM25 score (default 0.3)
}

// DefaultHybridConfig returns default hybrid configuration.
func DefaultHybridConfig() HybridConfig {
	return HybridConfig{
		VectorWeight:  0.7,
		KeywordWeight: 0.3,
	}
}

// MergeHybridResults merges vector and keyword search results.
// Uses weighted reciprocal rank fusion when scores are not directly comparable.
func MergeHybridResults(
	vectorResults []SearchResult,
	keywordResults []KeywordSearchResult,
	config HybridConfig,
	maxResults int,
) []SearchResult {
	if config.VectorWeight == 0 && config.KeywordWeight == 0 {
		// Default weights if not set
		config = DefaultHybridConfig()
	}

	// Create a map to collect all results by ID
	resultMap := make(map[string]*HybridResult)

	// Add vector results
	for _, vr := range vectorResults {
		resultMap[vr.ID] = &HybridResult{
			SearchResult: SearchResult{
				ID:       vr.ID,
				Content:  vr.Content,
				Metadata: vr.Metadata,
				Score:    0, // Will be calculated
			},
			VectorScore: vr.Score,
		}
	}

	// Add or merge keyword results
	for _, kr := range keywordResults {
		if existing, ok := resultMap[kr.ID]; ok {
			existing.KeywordScore = kr.Score
			existing.Metadata = DocumentMetadata{} // Could parse from kr.Metadata if needed
		} else {
			resultMap[kr.ID] = &HybridResult{
				SearchResult: SearchResult{
					ID:      kr.ID,
					Content: kr.Content,
					Score:   0,
				},
				KeywordScore: kr.Score,
			}
		}
	}

	// Calculate combined scores
	var merged []HybridResult
	for _, hr := range resultMap {
		// Weighted combination
		// If only one score is available, use it fully
		if hr.VectorScore == 0 && hr.KeywordScore > 0 {
			hr.Score = hr.KeywordScore
		} else if hr.KeywordScore == 0 && hr.VectorScore > 0 {
			hr.Score = hr.VectorScore
		} else {
			// Both scores available - weighted average
			totalWeight := config.VectorWeight + config.KeywordWeight
			hr.Score = (hr.VectorScore*config.VectorWeight + hr.KeywordScore*config.KeywordWeight) / totalWeight
		}
		merged = append(merged, *hr)
	}

	// Sort by combined score (descending)
	sort.Slice(merged, func(i, j int) bool {
		return merged[i].Score > merged[j].Score
	})

	// Return top maxResults
	if maxResults > len(merged) {
		maxResults = len(merged)
	}

	results := make([]SearchResult, maxResults)
	for i := 0; i < maxResults; i++ {
		results[i] = merged[i].SearchResult
	}

	return results
}

// ReciprocalRankFusion merges results using Reciprocal Rank Fusion (RRF).
// RRF is more effective when combining results from different ranking systems.
func ReciprocalRankFusion(
	vectorResults []SearchResult,
	keywordResults []KeywordSearchResult,
	k int, // RRF constant (typically 60)
	maxResults int,
) []SearchResult {
	if k <= 0 {
		k = 60 // Default RRF constant
	}

	// Map to collect RRF scores
	rrfScores := make(map[string]float32)
	contentMap := make(map[string]string)
	metadataMap := make(map[string]DocumentMetadata)

	// Score vector results by rank
	for rank, vr := range vectorResults {
		id := vr.ID
		rrfScores[id] += 1.0 / float32(k+rank+1)
		contentMap[id] = vr.Content
		metadataMap[id] = vr.Metadata
	}

	// Score keyword results by rank
	for rank, kr := range keywordResults {
		id := kr.ID
		rrfScores[id] += 1.0 / float32(k+rank+1)
		if _, ok := contentMap[id]; !ok {
			contentMap[id] = kr.Content
		}
	}

	// Convert to slice
	type scoredResult struct {
		id    string
		score float32
	}
	var scored []scoredResult
	for id, score := range rrfScores {
		scored = append(scored, scoredResult{id, score})
	}

	// Sort by RRF score (descending)
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// Return top maxResults
	if maxResults > len(scored) {
		maxResults = len(scored)
	}

	results := make([]SearchResult, maxResults)
	for i := 0; i < maxResults; i++ {
		sr := scored[i]
		results[i] = SearchResult{
			ID:       sr.id,
			Content:  contentMap[sr.id],
			Score:    sr.score,
			Metadata: metadataMap[sr.id],
		}
	}

	return results
}
