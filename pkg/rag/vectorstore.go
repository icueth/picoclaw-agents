// Package rag provides Retrieval-Augmented Generation functionality.
package rag

// SearchResult represents a single search result from the vector store.
type SearchResultItem struct {
	Item  string
	Score float32
}

// VectorStore provides in-memory vector storage and search.
type VectorStore struct {
	vectors [][]float32
	items   []string
}

// NewVectorStore creates a new in-memory vector store.
func NewVectorStore() *VectorStore {
	return &VectorStore{
		vectors: make([][]float32, 0),
		items:   make([]string, 0),
	}
}

// Add adds a vector with its associated item to the store.
func (vs *VectorStore) Add(vector []float32, item string) {
	vs.vectors = append(vs.vectors, vector)
	vs.items = append(vs.items, item)
}

// Size returns the number of vectors in the store.
func (vs *VectorStore) Size() int {
	return len(vs.vectors)
}

// Clear removes all vectors from the store.
func (vs *VectorStore) Clear() {
	vs.vectors = vs.vectors[:0]
	vs.items = vs.items[:0]
}

// scoredItem holds an index and score for sorting.
type scoredItem struct {
	index int
	score float32
}

// Search finds the k most similar vectors to the query.
func (vs *VectorStore) Search(query []float32, k int) []SearchResultItem {
	if len(vs.vectors) == 0 || k <= 0 {
		return nil
	}

	// Calculate similarities
	scored := make([]scoredItem, len(vs.vectors))
	for i, vec := range vs.vectors {
		scored[i] = scoredItem{
			index: i,
			score: CosineSimilarity(query, vec),
		}
	}

	// Sort by score (descending) - simple bubble sort for small k
	for i := 0; i < len(scored)-1; i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].score > scored[i].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	// Return top k
	if k > len(scored) {
		k = len(scored)
	}

	results := make([]SearchResultItem, k)
	for i := 0; i < k; i++ {
		results[i] = SearchResultItem{
			Item:  vs.items[scored[i].index],
			Score: scored[i].score,
		}
	}

	return results
}
