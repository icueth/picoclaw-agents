// Package rag provides Retrieval-Augmented Generation functionality.
package rag

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"picoclaw/agent/pkg/config"
	"picoclaw/agent/pkg/db"
)

// Manager handles RAG operations including document storage, retrieval, and search.
type Manager struct {
	db             *db.DB
	embedder       EmbeddingGenerator
	fallback       EmbeddingGenerator // Fallback embedder if primary fails
	keywordSearch  *KeywordSearcher   // FTS5 keyword search
	chunker        *Chunker
	config         *config.RAGConfig
	hybridConfig   HybridConfig
	useFallback    bool
}

// DocumentMetadata contains metadata for a RAG document.
type DocumentMetadata struct {
	Agent     string            `json:"agent,omitempty"`
	Role      string            `json:"role,omitempty"`
	Project   string            `json:"project,omitempty"`
	Source    string            `json:"source,omitempty"`
	Timestamp time.Time         `json:"timestamp,omitempty"`
	Extra     map[string]string `json:"extra,omitempty"`
}

// SearchResult represents a retrieved document with similarity score.
type SearchResult struct {
	ID       string           `json:"id"`
	Content  string           `json:"content"`
	Score    float32          `json:"score"`
	Metadata DocumentMetadata `json:"metadata"`
}

// ContextResult represents relevant context for a query.
type ContextResult struct {
	Documents []SearchResult `json:"documents"`
	Query     string         `json:"query"`
}

// NewManager creates a new RAG manager with appropriate embedder based on config.
func NewManager(database *db.DB, cfg *config.RAGConfig, client HTTPClient) (*Manager, error) {
	if database == nil {
		return nil, fmt.Errorf("database is required")
	}
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}

	// Create primary embedder based on configuration
	embedder, err := createEmbedder(cfg, client)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding generator: %w", err)
	}

	// Create fallback embedder (local) for resilience
	fallbackCfg := *cfg
	fallbackCfg.EmbeddingModel = "local"
	fallback, _ := createEmbedder(&fallbackCfg, client)

	chunkSize := cfg.ChunkSize
	if chunkSize == 0 {
		chunkSize = 512
	}

	overlap := cfg.Overlap
	if overlap == 0 {
		overlap = 128
	}

	mgr := &Manager{
		db:       database,
		embedder: embedder,
		fallback: fallback,
		chunker:  NewChunker(chunkSize, overlap),
		config:   cfg,
		hybridConfig: HybridConfig{
			VectorWeight:  float32(cfg.VectorWeight),
			KeywordWeight: float32(cfg.KeywordWeight),
		},
	}

	// Initialize keyword searcher (FTS5)
	mgr.keywordSearch = NewKeywordSearcher(database.Conn())

	return mgr, nil
}

// createEmbedder creates the appropriate embedding generator based on config.
func createEmbedder(cfg *config.RAGConfig, client HTTPClient) (EmbeddingGenerator, error) {
	switch cfg.EmbeddingModel {
	case "openai":
		if cfg.APIKey == "" {
			return nil, fmt.Errorf("OpenAI API key required for openai embedding model")
		}
		return NewOpenAIEmbeddingGenerator(cfg, client), nil
	case "http", "embeddinggemma", "embeddinggemma-300m-qat":
		// Use HTTP embedding service
		dimension := cfg.Dimension
		if dimension == 0 {
			dimension = 384 // Default for MiniLM
		}
		return NewHTTPEmbeddingGenerator(cfg.APIBase, cfg.EmbeddingModel, dimension), nil
	case "local", "":
		return NewLocalEmbeddingGenerator(cfg), nil
	default:
		// For custom or unknown models, use local as fallback
		return NewLocalEmbeddingGenerator(cfg), nil
	}
}

// AddDocument adds a document to the RAG system, chunking it and storing embeddings.
func (m *Manager) AddDocument(content string, metadata DocumentMetadata) ([]string, error) {
	if content == "" {
		return nil, fmt.Errorf("content cannot be empty")
	}

	// Chunk the content
	chunks := m.chunker.Chunk(content)
	if len(chunks) == 0 {
		chunks = []string{content}
	}

	// Generate embeddings for all chunks (with fallback)
	embeddings, err := m.generateEmbeddingsWithFallback(chunks)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embeddings: %w", err)
	}

	if len(embeddings) != len(chunks) {
		return nil, fmt.Errorf("embedding count mismatch: expected %d, got %d", len(chunks), len(embeddings))
	}

	// Store each chunk as a separate document
	var docIDs []string
	for i, chunk := range chunks {
		// Convert metadata to db.Metadata
		dbMetadata := m.toDBMetadata(metadata, i, len(chunks))

		// Serialize embedding
		embeddingBytes := SerializeEmbedding(embeddings[i])

		// Store in database
		doc, err := m.db.CreateRAGDocument(chunk, embeddingBytes, dbMetadata)
		if err != nil {
			return nil, fmt.Errorf("failed to create RAG document for chunk %d: %w", i, err)
		}

		docIDs = append(docIDs, doc.ID)
	}

	return docIDs, nil
}

// generateEmbeddingsWithFallback generates embeddings with fallback to local embedder.
func (m *Manager) generateEmbeddingsWithFallback(texts []string) ([][]float32, error) {
	// Try primary embedder first
	embeddings, err := m.embedder.GenerateBatch(texts)
	if err == nil {
		m.useFallback = false
		return embeddings, nil
	}

	// Log the error and try fallback if available
	log.Printf("Primary embedder failed: %v, trying fallback", err)

	if m.fallback != nil {
		embeddings, err = m.fallback.GenerateBatch(texts)
		if err == nil {
			m.useFallback = true
			log.Printf("Using fallback embedder (local)")
			return embeddings, nil
		}
	}

	return nil, fmt.Errorf("all embedders failed: %w", err)
}

// AddDocumentWithID adds a document with a specific ID (for updates).
func (m *Manager) AddDocumentWithID(docID, content string, metadata DocumentMetadata) error {
	if content == "" {
		return fmt.Errorf("content cannot be empty")
	}
	if docID == "" {
		return fmt.Errorf("document ID cannot be empty")
	}

	// Generate embedding
	embedding, err := m.embedder.Generate(content)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Convert metadata
	dbMetadata := m.toDBMetadata(metadata, 0, 1)

	// Serialize embedding
	embeddingBytes := SerializeEmbedding(embedding)

	// Delete existing document if it exists
	_ = m.db.DeleteRAGDocument(docID)

	// Insert with specific ID using raw SQL
	metadataJSON, _ := json.Marshal(dbMetadata)
	_, err = m.db.Conn().Exec(
		`INSERT INTO rag_documents (id, content, embedding, metadata) VALUES (?, ?, ?, ?)`,
		docID, content, embeddingBytes, string(metadataJSON),
	)
	if err != nil {
		return fmt.Errorf("failed to insert document: %w", err)
	}

	return nil
}

// Search finds documents similar to the query using hybrid search (vector + keyword).
func (m *Manager) Search(query string, maxResults int) ([]SearchResult, error) {
	if query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	if maxResults <= 0 {
		maxResults = m.config.MaxResults
		if maxResults == 0 {
			maxResults = 5
		}
	}

	// Check if we're using noop provider (keyword-only mode)
	if m.embedder.IsNoop() {
		// Keyword-only search
		return m.keywordSearchOnly(query, maxResults)
	}

	// Hybrid search: vector + keyword
	return m.hybridSearch(query, maxResults)
}

// keywordSearchOnly performs keyword-only search (when embedding_provider = none).
func (m *Manager) keywordSearchOnly(query string, maxResults int) ([]SearchResult, error) {
	keywordResults, err := m.keywordSearch.Search(query, maxResults*2)
	if err != nil {
		return nil, fmt.Errorf("keyword search failed: %w", err)
	}

	results := make([]SearchResult, 0, len(keywordResults))
	for _, kr := range keywordResults {
		results = append(results, SearchResult{
			ID:      kr.ID,
			Content: kr.Content,
			Score:   kr.Score,
		})
		if len(results) >= maxResults {
			break
		}
	}

	return results, nil
}

// hybridSearch performs hybrid search combining vector similarity and keyword matching.
func (m *Manager) hybridSearch(query string, maxResults int) ([]SearchResult, error) {
	// Get vector search results
	vectorResults, err := m.vectorSearch(query, maxResults*2)
	if err != nil {
		// Fall back to keyword-only if vector search fails
		vectorResults = nil
	}

	// Get keyword search results
	keywordResults, err := m.keywordSearch.Search(query, maxResults*2)
	if err != nil {
		keywordResults = nil
	}

	// If no keyword results, return vector results only
	if len(keywordResults) == 0 {
		if len(vectorResults) > maxResults {
			return vectorResults[:maxResults], nil
		}
		return vectorResults, nil
	}

	// If no vector results, return keyword results only
	if len(vectorResults) == 0 {
		results := make([]SearchResult, 0, len(keywordResults))
		for _, kr := range keywordResults {
			results = append(results, SearchResult{
				ID:      kr.ID,
				Content: kr.Content,
				Score:   kr.Score,
			})
			if len(results) >= maxResults {
				break
			}
		}
		return results, nil
	}

	// Merge results using hybrid approach
	return MergeHybridResults(vectorResults, keywordResults, m.hybridConfig, maxResults), nil
}

// vectorSearch performs vector similarity search.
func (m *Manager) vectorSearch(query string, maxResults int) ([]SearchResult, error) {
	// Generate query embedding (with fallback)
	queryEmbedding, err := m.generateQueryEmbedding(query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Retrieve IDs and embeddings from database (memory-efficient)
	// Default to last 2000 documents for performance if no limit specified elsewhere
	docs, err := m.db.ListRAGEmbeddings(2000) 
	if err != nil {
		return nil, fmt.Errorf("failed to list embeddings: %w", err)
	}

	if len(docs) == 0 {
		return []SearchResult{}, nil
	}

	// Determine which embedder dimension to use
	dimension := m.embedder.Dimension()
	if m.useFallback && m.fallback != nil {
		dimension = m.fallback.Dimension()
	}

	// Deserialize embeddings and calculate similarity
	candidates := make([][]float32, len(docs))
	for i, doc := range docs {
		embedding, err := DeserializeEmbedding(doc.Embedding, dimension)
		if err != nil {
			// Skip documents with invalid embeddings
			continue
		}
		candidates[i] = embedding
	}

	// Find top-k similar documents
	threshold := float32(m.config.SimilarityThreshold)
	topResults := FindTopK(queryEmbedding, candidates, maxResults)

	// Filter by threshold if specified
	if threshold > 0 {
		filtered := make([]VectorIndex, 0, len(topResults))
		for _, r := range topResults {
			if r.Score >= threshold {
				filtered = append(filtered, r)
			}
		}
		topResults = filtered
	}

	// Build search results (fetch full content only for top results)
	results := make([]SearchResult, 0, len(topResults))
	for _, r := range topResults {
		if r.Index < len(docs) {
			docID := docs[r.Index].ID
			
			// Fetch full document including content and metadata
			fullDoc, err := m.db.GetRAGDocument(docID)
			if err != nil {
				continue
			}

			results = append(results, SearchResult{
				ID:      fullDoc.ID,
				Content: fullDoc.Content,
				Score:   r.Score,
				Metadata: DocumentMetadata{
					Agent:     fullDoc.Metadata.Agent,
					Role:      fullDoc.Metadata.Role,
					Project:   fullDoc.Metadata.Project,
					Source:    fullDoc.Metadata.Source,
					Timestamp: fullDoc.Metadata.Timestamp,
					Extra:     fullDoc.Metadata.Extra,
				},
			})
		}
	}

	return results, nil
}

// generateQueryEmbedding generates embedding for query with fallback.
func (m *Manager) generateQueryEmbedding(query string) ([]float32, error) {
	// Try primary embedder first
	embedding, err := m.embedder.Generate(query)
	if err == nil {
		m.useFallback = false
		return embedding, nil
	}

	// Log the error and try fallback if available
	log.Printf("Primary embedder failed for query: %v, trying fallback", err)

	if m.fallback != nil {
		embedding, err = m.fallback.Generate(query)
		if err == nil {
			m.useFallback = true
			log.Printf("Using fallback embedder (local) for query")
			return embedding, nil
		}
	}

	return nil, fmt.Errorf("all embedders failed: %w", err)
}

// GetContextForQuery retrieves relevant context for a query.
// This is the main method for RAG retrieval.
func (m *Manager) GetContextForQuery(query string) (*ContextResult, error) {
	results, err := m.Search(query, m.config.MaxResults)
	if err != nil {
		return nil, err
	}

	return &ContextResult{
		Documents: results,
		Query:     query,
	}, nil
}

// GetContextString returns the context as a formatted string for inclusion in prompts.
func (m *Manager) GetContextString(query string) (string, error) {
	ctx, err := m.GetContextForQuery(query)
	if err != nil {
		return "", err
	}

	if len(ctx.Documents) == 0 {
		return "", nil
	}

	result := "## Relevant Context\n\n"
	for i, doc := range ctx.Documents {
		result += fmt.Sprintf("### Document %d (relevance: %.2f)\n%s\n\n", i+1, doc.Score, doc.Content)
	}

	return result, nil
}

// DeleteOldDocuments removes documents created before the given time.
func (m *Manager) DeleteOldDocuments(before time.Time) (int, error) {
	result, err := m.db.Conn().Exec(
		`DELETE FROM rag_documents WHERE created_at < ?`,
		before,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old documents: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return int(rowsAffected), nil
}

// DeleteDocument removes a specific document by ID.
func (m *Manager) DeleteDocument(docID string) error {
	return m.db.DeleteRAGDocument(docID)
}

// GetDocument retrieves a specific document by ID.
func (m *Manager) GetDocument(docID string) (*db.RAGDocument, error) {
	return m.db.GetRAGDocument(docID)
}

// ListDocuments returns all documents, optionally limited.
func (m *Manager) ListDocuments(limit int) ([]*db.RAGDocument, error) {
	return m.db.ListRAGDocuments(limit)
}

// UpdateDocument updates an existing document.
func (m *Manager) UpdateDocument(docID, content string, metadata DocumentMetadata) error {
	return m.AddDocumentWithID(docID, content, metadata)
}

// Count returns the total number of documents in the RAG system.
func (m *Manager) Count() (int, error) {
	var count int
	err := m.db.Conn().QueryRow(`SELECT COUNT(*) FROM rag_documents`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count documents: %w", err)
	}
	return count, nil
}

// SearchByMetadata searches for documents matching specific metadata criteria.
func (m *Manager) SearchByMetadata(agent, role, project string, limit int) ([]*db.RAGDocument, error) {
	query := `SELECT id, content, embedding, metadata, created_at FROM rag_documents WHERE 1=1`
	var args []interface{}

	// Build dynamic query based on provided filters
	// Note: This is a simplified version; in production, you'd want proper JSON querying
	if agent != "" {
		query += ` AND metadata LIKE ?`
		args = append(args, `%"agent":"`+agent+`"%`)
	}
	if role != "" {
		query += ` AND metadata LIKE ?`
		args = append(args, `%"role":"`+role+`"%`)
	}
	if project != "" {
		query += ` AND metadata LIKE ?`
		args = append(args, `%"project":"`+project+`"%`)
	}

	query += ` ORDER BY created_at DESC`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := m.db.Conn().Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search by metadata: %w", err)
	}
	defer rows.Close()

	return scanRAGDocumentsFromRows(rows)
}

// scanRAGDocumentsFromRows scans SQL rows into RAGDocument structs.
func scanRAGDocumentsFromRows(rows *sql.Rows) ([]*db.RAGDocument, error) {
	var docs []*db.RAGDocument
	for rows.Next() {
		var doc db.RAGDocument
		var metadataJSON string

		if err := rows.Scan(&doc.ID, &doc.Content, &doc.Embedding, &metadataJSON, &doc.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan RAG document: %w", err)
		}

		if err := json.Unmarshal([]byte(metadataJSON), &doc.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		docs = append(docs, &doc)
	}
	return docs, rows.Err()
}

// toDBMetadata converts DocumentMetadata to db.Metadata.
func (m *Manager) toDBMetadata(meta DocumentMetadata, chunkIndex, totalChunks int) db.Metadata {
	dbMeta := db.Metadata{
		Agent:     meta.Agent,
		Role:      meta.Role,
		Project:   meta.Project,
		Timestamp: meta.Timestamp,
	}

	if dbMeta.Timestamp.IsZero() {
		dbMeta.Timestamp = time.Now()
	}

	return dbMeta
}

// ReindexAll regenerates embeddings for all documents.
// This is useful when changing embedding models or dimensions.
func (m *Manager) ReindexAll(ctx context.Context) error {
	docs, err := m.db.ListRAGDocuments(0)
	if err != nil {
		return fmt.Errorf("failed to list documents: %w", err)
	}

	for _, doc := range docs {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Regenerate embedding
		embedding, err := m.embedder.Generate(doc.Content)
		if err != nil {
			continue // Skip failed documents
		}

		// Update embedding in database
		embeddingBytes := SerializeEmbedding(embedding)
		_, err = m.db.Conn().Exec(
			`UPDATE rag_documents SET embedding = ? WHERE id = ?`,
			embeddingBytes, doc.ID,
		)
		if err != nil {
			continue // Skip failed updates
		}
	}

	return nil
}

// Close closes the RAG manager and releases resources.
func (m *Manager) Close() error {
	// Nothing to close currently, but good practice to have this method
	return nil
}
