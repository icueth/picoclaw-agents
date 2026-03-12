// Package rag provides Retrieval-Augmented Generation functionality.
package rag

import (
	"context"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"

	"picoclaw/agent/pkg/config"
	"picoclaw/agent/pkg/db"
)

// mockHTTPClient is a mock implementation of HTTPClient for testing.
type mockHTTPClient struct {
	response []byte
	err      error
}

func (m *mockHTTPClient) Post(url string, headers map[string]string, body []byte) ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.response, nil
}

// setupTestDB creates a temporary test database.
func setupTestDB(t *testing.T) (*db.DB, func()) {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	database, err := db.New(dbPath)
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	if err := database.Init(); err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}

	cleanup := func() {
		database.Close()
		os.Remove(dbPath)
	}

	return database, cleanup
}

// TestLocalEmbeddingGenerator tests the local embedding generator.
func TestLocalEmbeddingGenerator(t *testing.T) {
	cfg := &config.RAGConfig{
		Dimension: 384,
	}

	gen := NewLocalEmbeddingGenerator(cfg)

	t.Run("Dimension", func(t *testing.T) {
		if gen.Dimension() != 384 {
			t.Errorf("expected dimension 384, got %d", gen.Dimension())
		}
	})

	t.Run("Generate", func(t *testing.T) {
		embedding, err := gen.Generate("hello world")
		if err != nil {
			t.Fatalf("failed to generate embedding: %v", err)
		}

		if len(embedding) != 384 {
			t.Errorf("expected embedding length 384, got %d", len(embedding))
		}

		// Check that embedding is normalized (L2 norm ≈ 1)
		var sum float64
		for _, v := range embedding {
			sum += float64(v * v)
		}
		norm := math.Sqrt(sum)
		if math.Abs(norm-1.0) > 0.01 {
			t.Errorf("expected normalized vector (norm ≈ 1), got %f", norm)
		}
	})

	t.Run("Generate_Empty", func(t *testing.T) {
		embedding, err := gen.Generate("")
		if err != nil {
			t.Fatalf("failed to generate embedding for empty string: %v", err)
		}

		if len(embedding) != 384 {
			t.Errorf("expected embedding length 384, got %d", len(embedding))
		}

		// Empty string should produce zero vector
		allZero := true
		for _, v := range embedding {
			if v != 0 {
				allZero = false
				break
			}
		}
		if !allZero {
			t.Error("expected zero vector for empty string")
		}
	})

	t.Run("GenerateBatch", func(t *testing.T) {
		texts := []string{"hello", "world", "test"}
		embeddings, err := gen.GenerateBatch(texts)
		if err != nil {
			t.Fatalf("failed to generate batch embeddings: %v", err)
		}

		if len(embeddings) != len(texts) {
			t.Errorf("expected %d embeddings, got %d", len(texts), len(embeddings))
		}

		for i, emb := range embeddings {
			if len(emb) != 384 {
				t.Errorf("embedding %d: expected length 384, got %d", i, len(emb))
			}
		}
	})

	t.Run("Similarity", func(t *testing.T) {
		emb1, _ := gen.Generate("machine learning")
		emb2, _ := gen.Generate("artificial intelligence")
		emb3, _ := gen.Generate("pizza recipe")

		// Similar texts should have higher similarity
		sim1 := CosineSimilarity(emb1, emb2)
		sim2 := CosineSimilarity(emb1, emb3)

		if sim1 <= sim2 {
			t.Errorf("expected similar texts to have higher similarity: sim1=%f, sim2=%f", sim1, sim2)
		}
	})
}

// TestChunker tests the text chunker.
func TestChunker(t *testing.T) {
	t.Run("NewChunker", func(t *testing.T) {
		c := NewChunker(512, 128)
		if c.ChunkSize() != 512 {
			t.Errorf("expected chunk size 512, got %d", c.ChunkSize())
		}
		if c.Overlap() != 128 {
			t.Errorf("expected overlap 128, got %d", c.Overlap())
		}
	})

	t.Run("NewChunker_Defaults", func(t *testing.T) {
		c := NewChunker(0, -1)
		if c.ChunkSize() != 512 {
			t.Errorf("expected default chunk size 512, got %d", c.ChunkSize())
		}
		if c.Overlap() < 0 {
			t.Errorf("expected non-negative overlap, got %d", c.Overlap())
		}
	})

	t.Run("Chunk_ShortText", func(t *testing.T) {
		c := NewChunker(100, 20)
		text := "This is a short text."
		chunks := c.Chunk(text)

		if len(chunks) != 1 {
			t.Errorf("expected 1 chunk for short text, got %d", len(chunks))
		}
		if chunks[0] != text {
			t.Errorf("expected chunk to match original text, got %s", chunks[0])
		}
	})

	t.Run("Chunk_Empty", func(t *testing.T) {
		c := NewChunker(100, 20)
		chunks := c.Chunk("")

		if chunks != nil {
			t.Errorf("expected nil for empty text, got %v", chunks)
		}
	})

	t.Run("Chunk_LongText", func(t *testing.T) {
		c := NewChunker(50, 10)
		// Create a text longer than chunk size
		text := "This is a very long text that needs to be chunked into multiple pieces. " +
			"It contains multiple sentences and should be split appropriately. " +
			"Each chunk should have some overlap with the previous one."

		chunks := c.Chunk(text)

		if len(chunks) < 2 {
			t.Errorf("expected multiple chunks for long text, got %d", len(chunks))
		}

		// Verify all chunks are non-empty
		for i, chunk := range chunks {
			if chunk == "" {
				t.Errorf("chunk %d is empty", i)
			}
		}
	})

	t.Run("ChunkBySentences", func(t *testing.T) {
		c := NewChunker(100, 20)
		text := "First sentence here. Second sentence here. Third sentence here. Fourth sentence here."

		chunks := c.ChunkBySentences(text)

		if len(chunks) == 0 {
			t.Error("expected chunks from sentence splitting")
		}

		// Verify chunks try to preserve sentences
		for _, chunk := range chunks {
			if chunk == "" {
				t.Error("got empty chunk")
			}
		}
	})
}

// TestCosineSimilarity tests the cosine similarity function.
func TestCosineSimilarity(t *testing.T) {
	t.Run("IdenticalVectors", func(t *testing.T) {
		a := []float32{1, 0, 0}
		b := []float32{1, 0, 0}
		sim := CosineSimilarity(a, b)

		if math.Abs(float64(sim)-1.0) > 0.0001 {
			t.Errorf("expected similarity 1.0 for identical vectors, got %f", sim)
		}
	})

	t.Run("OppositeVectors", func(t *testing.T) {
		a := []float32{1, 0, 0}
		b := []float32{-1, 0, 0}
		sim := CosineSimilarity(a, b)

		if math.Abs(float64(sim)-(-1.0)) > 0.0001 {
			t.Errorf("expected similarity -1.0 for opposite vectors, got %f", sim)
		}
	})

	t.Run("OrthogonalVectors", func(t *testing.T) {
		a := []float32{1, 0, 0}
		b := []float32{0, 1, 0}
		sim := CosineSimilarity(a, b)

		if math.Abs(float64(sim)) > 0.0001 {
			t.Errorf("expected similarity 0.0 for orthogonal vectors, got %f", sim)
		}
	})

	t.Run("DifferentLengths", func(t *testing.T) {
		a := []float32{1, 0}
		b := []float32{1, 0, 0}
		sim := CosineSimilarity(a, b)

		if sim != 0 {
			t.Errorf("expected similarity 0 for different length vectors, got %f", sim)
		}
	})

	t.Run("EmptyVectors", func(t *testing.T) {
		a := []float32{}
		b := []float32{}
		sim := CosineSimilarity(a, b)

		if sim != 0 {
			t.Errorf("expected similarity 0 for empty vectors, got %f", sim)
		}
	})
}

// TestFindTopK tests the FindTopK function.
func TestFindTopK(t *testing.T) {
	query := []float32{1, 0, 0}
	candidates := [][]float32{
		{1, 0, 0},    // Most similar
		{0, 1, 0},    // Orthogonal
		{-1, 0, 0},   // Opposite
		{0.9, 0.1, 0}, // Very similar
		{0, 0, 1},    // Orthogonal
	}

	t.Run("FindTop3", func(t *testing.T) {
		results := FindTopK(query, candidates, 3)

		if len(results) != 3 {
			t.Errorf("expected 3 results, got %d", len(results))
		}

		// Check that results are sorted by similarity (descending)
		for i := 1; i < len(results); i++ {
			if results[i].Score > results[i-1].Score {
				t.Error("results not sorted by score descending")
			}
		}

		// First result should be the identical vector
		if results[0].Index != 0 {
			t.Errorf("expected first result to be index 0, got %d", results[0].Index)
		}
	})

	t.Run("FindTopKWithThreshold", func(t *testing.T) {
		results := FindTopKWithThreshold(query, candidates, 5, 0.5)

		// Only vectors with similarity >= 0.5 should be returned
		for _, r := range results {
			if r.Score < 0.5 {
				t.Errorf("result below threshold: score=%f", r.Score)
			}
		}
	})

	t.Run("KGreaterThanCandidates", func(t *testing.T) {
		results := FindTopK(query, candidates, 10)

		if len(results) != len(candidates) {
			t.Errorf("expected %d results, got %d", len(candidates), len(results))
		}
	})
}

// TestVectorStore tests the VectorStore.
func TestVectorStore(t *testing.T) {
	t.Run("AddAndSearch", func(t *testing.T) {
		vs := NewVectorStore()

		// Add some vectors
		vs.Add([]float32{1, 0, 0}, "item1")
		vs.Add([]float32{0, 1, 0}, "item2")
		vs.Add([]float32{0, 0, 1}, "item3")

		if vs.Size() != 3 {
			t.Errorf("expected size 3, got %d", vs.Size())
		}

		// Search
		results := vs.Search([]float32{1, 0, 0}, 2)

		if len(results) != 2 {
			t.Errorf("expected 2 results, got %d", len(results))
		}

		if results[0].Item != "item1" {
			t.Errorf("expected first result to be item1, got %v", results[0].Item)
		}
	})

	t.Run("Clear", func(t *testing.T) {
		vs := NewVectorStore()
		vs.Add([]float32{1, 0, 0}, "item1")
		vs.Clear()

		if vs.Size() != 0 {
			t.Errorf("expected size 0 after clear, got %d", vs.Size())
		}
	})
}

// TestManager tests the RAG manager.
func TestManager(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := &config.RAGConfig{
		Enabled:             true,
		EmbeddingModel:      "local",
		Dimension:           384,
		ChunkSize:           100,
		Overlap:             20,
		MaxResults:          5,
		SimilarityThreshold: 0.0,
	}

	manager, err := NewManager(database, cfg, nil)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	defer manager.Close()

	t.Run("AddDocument", func(t *testing.T) {
		content := "This is a test document for RAG. It contains important information."
		metadata := DocumentMetadata{
			Agent:   "test-agent",
			Role:    "tester",
			Project: "rag-test",
		}

		ids, err := manager.AddDocument(content, metadata)
		if err != nil {
			t.Fatalf("failed to add document: %v", err)
		}

		if len(ids) == 0 {
			t.Error("expected at least one document ID")
		}

		// Verify document count
		count, err := manager.Count()
		if err != nil {
			t.Fatalf("failed to count documents: %v", err)
		}
		if count == 0 {
			t.Error("expected documents in database")
		}
	})

	t.Run("Search", func(t *testing.T) {
		// Add some documents first
		docs := []string{
			"Machine learning is a subset of artificial intelligence.",
			"Deep learning uses neural networks with many layers.",
			"Pizza is a popular Italian dish with various toppings.",
		}

		for _, doc := range docs {
			_, err := manager.AddDocument(doc, DocumentMetadata{Source: "test"})
			if err != nil {
				t.Fatalf("failed to add document: %v", err)
			}
		}

		// Search for AI-related content
		results, err := manager.Search("artificial intelligence", 3)
		if err != nil {
			t.Fatalf("failed to search: %v", err)
		}

		if len(results) == 0 {
			t.Error("expected search results")
		}
	})

	t.Run("GetContextForQuery", func(t *testing.T) {
		ctx, err := manager.GetContextForQuery("neural networks")
		if err != nil {
			t.Fatalf("failed to get context: %v", err)
		}

		if ctx.Query != "neural networks" {
			t.Errorf("expected query 'neural networks', got %s", ctx.Query)
		}
	})

	t.Run("GetContextString", func(t *testing.T) {
		contextStr, err := manager.GetContextString("machine learning")
		if err != nil {
			t.Fatalf("failed to get context string: %v", err)
		}

		if contextStr == "" {
			t.Error("expected non-empty context string")
		}
	})

	t.Run("GetDocument", func(t *testing.T) {
		// Add a document and retrieve it
		ids, err := manager.AddDocument("Test document content", DocumentMetadata{})
		if err != nil {
			t.Fatalf("failed to add document: %v", err)
		}

		doc, err := manager.GetDocument(ids[0])
		if err != nil {
			t.Fatalf("failed to get document: %v", err)
		}

		if doc.Content != "Test document content" {
			t.Errorf("expected content 'Test document content', got %s", doc.Content)
		}
	})

	t.Run("DeleteDocument", func(t *testing.T) {
		// Add and then delete a document
		ids, err := manager.AddDocument("Document to delete", DocumentMetadata{})
		if err != nil {
			t.Fatalf("failed to add document: %v", err)
		}

		err = manager.DeleteDocument(ids[0])
		if err != nil {
			t.Fatalf("failed to delete document: %v", err)
		}

		// Verify deletion
		_, err = manager.GetDocument(ids[0])
		if err == nil {
			t.Error("expected error when getting deleted document")
		}
	})

	t.Run("DeleteOldDocuments", func(t *testing.T) {
		// Add a document
		_, err := manager.AddDocument("Old document", DocumentMetadata{})
		if err != nil {
			t.Fatalf("failed to add document: %v", err)
		}

		// Delete documents older than now
		deleted, err := manager.DeleteOldDocuments(time.Now().Add(time.Hour))
		if err != nil {
			t.Fatalf("failed to delete old documents: %v", err)
		}

		// Should have deleted at least one document
		if deleted == 0 {
			t.Error("expected to delete at least one document")
		}
	})

	t.Run("UpdateDocument", func(t *testing.T) {
		// Add a document
		ids, err := manager.AddDocument("Original content", DocumentMetadata{})
		if err != nil {
			t.Fatalf("failed to add document: %v", err)
		}

		// Update it
		err = manager.UpdateDocument(ids[0], "Updated content", DocumentMetadata{})
		if err != nil {
			t.Fatalf("failed to update document: %v", err)
		}

		// Verify update
		doc, err := manager.GetDocument(ids[0])
		if err != nil {
			t.Fatalf("failed to get updated document: %v", err)
		}

		if doc.Content != "Updated content" {
			t.Errorf("expected 'Updated content', got %s", doc.Content)
		}
	})
}

// TestManagerErrors tests error handling in the manager.
func TestManagerErrors(t *testing.T) {
	t.Run("NewManager_NilDB", func(t *testing.T) {
		cfg := &config.RAGConfig{Enabled: true}
		_, err := NewManager(nil, cfg, nil)
		if err == nil {
			t.Error("expected error for nil database")
		}
	})

	t.Run("NewManager_NilConfig", func(t *testing.T) {
		database, cleanup := setupTestDB(t)
		defer cleanup()

		_, err := NewManager(database, nil, nil)
		if err == nil {
			t.Error("expected error for nil config")
		}
	})

	t.Run("AddDocument_EmptyContent", func(t *testing.T) {
		database, cleanup := setupTestDB(t)
		defer cleanup()

		cfg := &config.RAGConfig{
			Enabled:        true,
			EmbeddingModel: "local",
		}

		manager, _ := NewManager(database, cfg, nil)
		_, err := manager.AddDocument("", DocumentMetadata{})
		if err == nil {
			t.Error("expected error for empty content")
		}
	})

	t.Run("Search_EmptyQuery", func(t *testing.T) {
		database, cleanup := setupTestDB(t)
		defer cleanup()

		cfg := &config.RAGConfig{
			Enabled:        true,
			EmbeddingModel: "local",
		}

		manager, _ := NewManager(database, cfg, nil)
		_, err := manager.Search("", 5)
		if err == nil {
			t.Error("expected error for empty query")
		}
	})
}

// TestEmbeddingSerialization tests embedding serialization.
func TestEmbeddingSerialization(t *testing.T) {
	original := []float32{0.1, 0.2, 0.3, 0.4, 0.5}

	serialized := SerializeEmbedding(original)
	if len(serialized) != len(original)*4 {
		t.Errorf("expected %d bytes, got %d", len(original)*4, len(serialized))
	}

	deserialized, err := DeserializeEmbedding(serialized, len(original))
	if err != nil {
		t.Fatalf("failed to deserialize: %v", err)
	}

	if len(deserialized) != len(original) {
		t.Errorf("expected %d values, got %d", len(original), len(deserialized))
	}

	for i := range original {
		if math.Abs(float64(original[i]-deserialized[i])) > 0.0001 {
			t.Errorf("value mismatch at index %d: expected %f, got %f", i, original[i], deserialized[i])
		}
	}
}

// TestEmbeddingSerializationErrors tests error handling in serialization.
func TestEmbeddingSerializationErrors(t *testing.T) {
	t.Run("Deserialize_WrongSize", func(t *testing.T) {
		data := []byte{0, 0, 0, 0} // 4 bytes = 1 float32
		_, err := DeserializeEmbedding(data, 2) // Expecting 2 floats = 8 bytes
		if err == nil {
			t.Error("expected error for wrong size")
		}
	})
}

// TestOpenAIEmbeddingGenerator tests the OpenAI embedding generator.
func TestOpenAIEmbeddingGenerator(t *testing.T) {
	mockResponse := `{
		"object": "list",
		"data": [
			{
				"object": "embedding",
				"embedding": [0.1, 0.2, 0.3, 0.4],
				"index": 0
			}
		],
		"model": "text-embedding-3-small",
		"usage": {
			"prompt_tokens": 10,
			"total_tokens": 10
		}
	}`

	mockClient := &mockHTTPClient{
		response: []byte(mockResponse),
	}

	cfg := &config.RAGConfig{
		Enabled:        true,
		EmbeddingModel: "openai",
		APIKey:         "test-key",
		APIBase:        "https://api.openai.com/v1",
		Dimension:      4,
	}

	gen := NewOpenAIEmbeddingGenerator(cfg, mockClient)

	t.Run("Dimension", func(t *testing.T) {
		if gen.Dimension() != 4 {
			t.Errorf("expected dimension 4, got %d", gen.Dimension())
		}
	})

	t.Run("Generate", func(t *testing.T) {
		embedding, err := gen.Generate("test text")
		if err != nil {
			t.Fatalf("failed to generate: %v", err)
		}

		if len(embedding) != 4 {
			t.Errorf("expected 4 values, got %d", len(embedding))
		}
	})

	t.Run("Generate_NoClient", func(t *testing.T) {
		genNoClient := NewOpenAIEmbeddingGenerator(cfg, nil)
		_, err := genNoClient.Generate("test")
		if err == nil {
			t.Error("expected error when client is nil")
		}
	})

	t.Run("Generate_NoAPIKey", func(t *testing.T) {
		cfgNoKey := &config.RAGConfig{
			Enabled:        true,
			EmbeddingModel: "openai",
			APIKey:         "",
		}
		genNoKey := NewOpenAIEmbeddingGenerator(cfgNoKey, mockClient)
		_, err := genNoKey.Generate("test")
		if err == nil {
			t.Error("expected error when API key is missing")
		}
	})
}

// TestNewEmbeddingGenerator tests the embedding generator factory.
func TestNewEmbeddingGenerator(t *testing.T) {
	t.Run("Local", func(t *testing.T) {
		cfg := &config.RAGConfig{
			Enabled:        true,
			EmbeddingModel: "local",
		}

		gen, err := NewEmbeddingGenerator(cfg, nil)
		if err != nil {
			t.Fatalf("failed to create generator: %v", err)
		}

		if _, ok := gen.(*LocalEmbeddingGenerator); !ok {
			t.Error("expected LocalEmbeddingGenerator")
		}
	})

	t.Run("OpenAI", func(t *testing.T) {
		cfg := &config.RAGConfig{
			Enabled:        true,
			EmbeddingModel: "openai",
			APIKey:         "test-key",
		}

		gen, err := NewEmbeddingGenerator(cfg, &mockHTTPClient{})
		if err != nil {
			t.Fatalf("failed to create generator: %v", err)
		}

		if _, ok := gen.(*OpenAIEmbeddingGenerator); !ok {
			t.Error("expected OpenAIEmbeddingGenerator")
		}
	})

	t.Run("OpenAI_NoKey", func(t *testing.T) {
		cfg := &config.RAGConfig{
			Enabled:        true,
			EmbeddingModel: "openai",
			APIKey:         "",
		}

		_, err := NewEmbeddingGenerator(cfg, nil)
		if err == nil {
			t.Error("expected error when OpenAI key is missing")
		}
	})

	t.Run("Disabled", func(t *testing.T) {
		cfg := &config.RAGConfig{
			Enabled: false,
		}

		_, err := NewEmbeddingGenerator(cfg, nil)
		if err == nil {
			t.Error("expected error when RAG is disabled")
		}
	})

	t.Run("Default", func(t *testing.T) {
		cfg := &config.RAGConfig{
			Enabled:        true,
			EmbeddingModel: "",
		}

		gen, err := NewEmbeddingGenerator(cfg, nil)
		if err != nil {
			t.Fatalf("failed to create generator: %v", err)
		}

		if _, ok := gen.(*LocalEmbeddingGenerator); !ok {
			t.Error("expected LocalEmbeddingGenerator for empty model")
		}
	})
}

// TestManagerReindexAll tests the ReindexAll functionality.
func TestManagerReindexAll(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	cfg := &config.RAGConfig{
		Enabled:        true,
		EmbeddingModel: "local",
		Dimension:      384,
	}

	manager, err := NewManager(database, cfg, nil)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Add some documents
	for i := 0; i < 3; i++ {
		_, err := manager.AddDocument("Document content", DocumentMetadata{})
		if err != nil {
			t.Fatalf("failed to add document: %v", err)
		}
	}

	ctx := context.Background()
	err = manager.ReindexAll(ctx)
	if err != nil {
		t.Fatalf("failed to reindex: %v", err)
	}

	// Verify documents still exist
	count, err := manager.Count()
	if err != nil {
		t.Fatalf("failed to count: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 documents after reindex, got %d", count)
	}
}

// BenchmarkLocalEmbedding benchmarks the local embedding generator.
func BenchmarkLocalEmbedding(b *testing.B) {
	cfg := &config.RAGConfig{Dimension: 384}
	gen := NewLocalEmbeddingGenerator(cfg)
	text := "This is a sample text for benchmarking the embedding generator performance."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.Generate(text)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkChunker benchmarks the text chunker.
func BenchmarkChunker(b *testing.B) {
	c := NewChunker(512, 128)
	text := "This is a sentence. " +
		"Here is another sentence with more words. " +
		"The quick brown fox jumps over the lazy dog. "

	// Make it longer
	longText := ""
	for i := 0; i < 100; i++ {
		longText += text
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = c.Chunk(longText)
	}
}

// BenchmarkCosineSimilarity benchmarks cosine similarity calculation.
func BenchmarkCosineSimilarity(b *testing.B) {
	a := make([]float32, 384)
	c := make([]float32, 384)

	// Fill with some values
	for i := range a {
		a[i] = float32(i) / 384.0
		c[i] = float32(384-i) / 384.0
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CosineSimilarity(a, c)
	}
}
