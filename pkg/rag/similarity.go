// Package rag provides Retrieval-Augmented Generation functionality.
package rag

import (
	"math"
	"sort"
)

// CosineSimilarity calculates the cosine similarity between two vectors.
// Returns a value between -1 and 1, where 1 means identical direction.
func CosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}

	if len(a) == 0 {
		return 0
	}

	var dotProduct float64
	var normA float64
	var normB float64

	for i := range a {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return float32(dotProduct / (math.Sqrt(normA) * math.Sqrt(normB)))
}

// EuclideanDistance calculates the Euclidean distance between two vectors.
// Returns a non-negative value, where 0 means identical vectors.
func EuclideanDistance(a, b []float32) float32 {
	if len(a) != len(b) {
		return math.MaxFloat32
	}

	if len(a) == 0 {
		return 0
	}

	var sum float64
	for i := range a {
		diff := float64(a[i] - b[i])
		sum += diff * diff
	}

	return float32(math.Sqrt(sum))
}

// DotProduct calculates the dot product of two vectors.
func DotProduct(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}

	var sum float64
	for i := range a {
		sum += float64(a[i]) * float64(b[i])
	}

	return float32(sum)
}

// Normalize normalizes a vector to unit length.
func Normalize(v []float32) []float32 {
	if len(v) == 0 {
		return v
	}

	var sum float64
	for _, x := range v {
		sum += float64(x) * float64(x)
	}

	norm := math.Sqrt(sum)
	if norm == 0 {
		return v
	}

	result := make([]float32, len(v))
	for i, x := range v {
		result[i] = float32(float64(x) / norm)
	}

	return result
}

// VectorIndex holds a vector with its index for sorting.
type VectorIndex struct {
	Index    int
	Vector   []float32
	Score    float32
}

// FindTopK finds the top k most similar vectors to the query.
func FindTopK(query []float32, candidates [][]float32, k int) []VectorIndex {
	if len(candidates) == 0 || k <= 0 {
		return nil
	}

	// Calculate similarity for all candidates
	scored := make([]VectorIndex, len(candidates))
	for i, candidate := range candidates {
		scored[i] = VectorIndex{
			Index:  i,
			Vector: candidate,
			Score:  CosineSimilarity(query, candidate),
		}
	}

	// Sort by score (descending)
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	// Return top k
	if k > len(scored) {
		k = len(scored)
	}
	return scored[:k]
}

// BatchCosineSimilarity calculates cosine similarity between query and multiple candidates.
func BatchCosineSimilarity(query []float32, candidates [][]float32) []float32 {
	scores := make([]float32, len(candidates))
	for i, candidate := range candidates {
		scores[i] = CosineSimilarity(query, candidate)
	}
	return scores
}

// MaxSimilarity finds the maximum similarity between query and any candidate.
func MaxSimilarity(query []float32, candidates [][]float32) float32 {
	if len(candidates) == 0 {
		return 0
	}

	maxScore := float32(-1)
	for _, candidate := range candidates {
		score := CosineSimilarity(query, candidate)
		if score > maxScore {
			maxScore = score
		}
	}

	return maxScore
}

// AverageSimilarity calculates the average similarity between query and all candidates.
func AverageSimilarity(query []float32, candidates [][]float32) float32 {
	if len(candidates) == 0 {
		return 0
	}

	var sum float64
	for _, candidate := range candidates {
		sum += float64(CosineSimilarity(query, candidate))
	}

	return float32(sum / float64(len(candidates)))
}

// FindTopKWithThreshold finds the top k most similar vectors above a threshold.
func FindTopKWithThreshold(query []float32, candidates [][]float32, k int, threshold float32) []VectorIndex {
	results := FindTopK(query, candidates, k)

	// Filter by threshold
	var filtered []VectorIndex
	for _, r := range results {
		if r.Score >= threshold {
			filtered = append(filtered, r)
		}
	}

	return filtered
}
