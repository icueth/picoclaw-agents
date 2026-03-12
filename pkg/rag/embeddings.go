// Package rag provides Retrieval-Augmented Generation functionality.
package rag

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"
	"unicode"

	"picoclaw/agent/pkg/config"
)

// EmbeddingGenerator is the interface for different embedding providers.
type EmbeddingGenerator interface {
	// Generate creates an embedding vector for a single text.
	Generate(text string) ([]float32, error)
	// GenerateBatch creates embedding vectors for multiple texts.
	GenerateBatch(texts []string) ([][]float32, error)
	// Dimension returns the dimension of the embedding vectors.
	Dimension() int
	// IsNoop returns true if this is a noop provider (no embeddings)
	IsNoop() bool
}

// NoopEmbeddingProvider is a no-op embedding provider that returns empty vectors.
// Used when embedding_provider = "none" for keyword-only search.
type NoopEmbeddingProvider struct{}

// NewNoopEmbeddingProvider creates a new no-op embedding provider.
func NewNoopEmbeddingProvider() *NoopEmbeddingProvider {
	return &NoopEmbeddingProvider{}
}

// Generate returns an empty embedding (no-op).
func (n *NoopEmbeddingProvider) Generate(text string) ([]float32, error) {
	return []float32{}, nil
}

// GenerateBatch returns empty embeddings for all texts.
func (n *NoopEmbeddingProvider) GenerateBatch(texts []string) ([][]float32, error) {
	return make([][]float32, len(texts)), nil
}

// Dimension returns 0 (no embeddings).
func (n *NoopEmbeddingProvider) Dimension() int {
	return 0
}

// IsNoop returns true for the no-op provider.
func (n *NoopEmbeddingProvider) IsNoop() bool {
	return true
}

// LocalEmbeddingGenerator uses a simple local approach for embeddings.
// For production, this should be replaced with a proper sentence transformer model.
type LocalEmbeddingGenerator struct {
	dimension int
	vocab     map[string]int
}

// NewLocalEmbeddingGenerator creates a new local embedding generator.
func NewLocalEmbeddingGenerator(cfg *config.RAGConfig) *LocalEmbeddingGenerator {
	dimension := cfg.Dimension
	if dimension == 0 {
		dimension = 384 // Default dimension
	}

	return &LocalEmbeddingGenerator{
		dimension: dimension,
		vocab:     make(map[string]int),
	}
}

// Generate creates a simple embedding based on character n-grams and word frequencies.
// This is a fallback implementation that doesn't require external models.
func (g *LocalEmbeddingGenerator) Generate(text string) ([]float32, error) {
	if text == "" {
		return make([]float32, g.dimension), nil
	}

	// Normalize text
	text = strings.ToLower(strings.TrimSpace(text))

	// Create a hash-based embedding using character n-grams
	embedding := make([]float32, g.dimension)

	// Add word-based features
	words := tokenize(text)
	wordFreq := make(map[string]int)
	for _, word := range words {
		wordFreq[word]++
	}

	// Add character n-gram features (2-grams and 3-grams)
	runes := []rune(text)
	for i := 0; i < len(runes)-1; i++ {
		// 2-grams
		bigram := string(runes[i : i+2])
		hash := hashString(bigram)
		idx := hash % g.dimension
		embedding[idx] += 1.0

		// 3-grams
		if i < len(runes)-2 {
			trigram := string(runes[i : i+3])
			hash := hashString(trigram)
			idx := hash % g.dimension
			embedding[idx] += 1.5
		}
	}

	// Add word-based features
	for word, freq := range wordFreq {
		hash := hashString(word)
		idx := hash % g.dimension
		// Weight by log frequency
		embedding[idx] += float32(math.Log1p(float64(freq))) * 2.0
	}

	// Normalize the embedding
	g.normalize(embedding)

	return embedding, nil
}

// GenerateBatch creates embeddings for multiple texts.
func (g *LocalEmbeddingGenerator) GenerateBatch(texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))
	for i, text := range texts {
		embedding, err := g.Generate(text)
		if err != nil {
			return nil, fmt.Errorf("failed to generate embedding for text %d: %w", i, err)
		}
		embeddings[i] = embedding
	}
	return embeddings, nil
}

// Dimension returns the embedding dimension.
func (g *LocalEmbeddingGenerator) Dimension() int {
	return g.dimension
}

// IsNoop returns false for local embedder.
func (g *LocalEmbeddingGenerator) IsNoop() bool {
	return false
}

// normalize normalizes the embedding vector to unit length.
func (g *LocalEmbeddingGenerator) normalize(vec []float32) {
	var sum float64
	for _, v := range vec {
		sum += float64(v * v)
	}
	norm := math.Sqrt(sum)
	if norm > 0 {
		for i := range vec {
			vec[i] = float32(float64(vec[i]) / norm)
		}
	}
}

// tokenize splits text into words.
func tokenize(text string) []string {
	var words []string
	var current strings.Builder

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current.WriteRune(r)
		} else if current.Len() > 0 {
			words = append(words, current.String())
			current.Reset()
		}
	}
	if current.Len() > 0 {
		words = append(words, current.String())
	}

	return words
}

// hashString creates a simple hash of a string.
func hashString(s string) int {
	h := 0
	for i, c := range s {
		h += int(c) * (i + 1)
	}
	if h < 0 {
		h = -h
	}
	return h
}

// OpenAIEmbeddingGenerator uses OpenAI's embedding API.
type OpenAIEmbeddingGenerator struct {
	apiKey     string
	apiBase    string
	model      string
	dimension  int
	httpClient HTTPClient
}

// HTTPClient is the interface for HTTP requests.
type HTTPClient interface {
	Post(url string, headers map[string]string, body []byte) ([]byte, error)
}

// NewOpenAIEmbeddingGenerator creates a new OpenAI embedding generator.
func NewOpenAIEmbeddingGenerator(cfg *config.RAGConfig, client HTTPClient) *OpenAIEmbeddingGenerator {
	model := cfg.EmbeddingModel
	if model == "" || model == "openai" {
		model = "text-embedding-3-small"
	}

	dimension := cfg.Dimension
	if dimension == 0 {
		if model == "text-embedding-3-small" {
			dimension = 1536
		} else if model == "text-embedding-ada-002" {
			dimension = 1536
		} else {
			dimension = 1536
		}
	}

	apiBase := cfg.APIBase
	if apiBase == "" {
		apiBase = "https://api.openai.com/v1"
	}

	return &OpenAIEmbeddingGenerator{
		apiKey:     cfg.APIKey,
		apiBase:    apiBase,
		model:      model,
		dimension:  dimension,
		httpClient: client,
	}
}

// openAIEmbeddingRequest is the request body for OpenAI embeddings API.
type openAIEmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

// openAIEmbeddingResponse is the response from OpenAI embeddings API.
type openAIEmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// Generate creates an embedding using OpenAI API.
func (g *OpenAIEmbeddingGenerator) Generate(text string) ([]float32, error) {
	embeddings, err := g.GenerateBatch([]string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}
	return embeddings[0], nil
}

// GenerateBatch creates embeddings for multiple texts using OpenAI API.
func (g *OpenAIEmbeddingGenerator) GenerateBatch(texts []string) ([][]float32, error) {
	if g.httpClient == nil {
		return nil, fmt.Errorf("HTTP client not configured")
	}
	if g.apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key not configured")
	}

	reqBody := openAIEmbeddingRequest{
		Model: g.model,
		Input: texts,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + g.apiKey,
	}

	respBody, err := g.httpClient.Post(g.apiBase+"/embeddings", headers, jsonBody)
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenAI API: %w", err)
	}

	var resp openAIEmbeddingResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	embeddings := make([][]float32, len(resp.Data))
	for _, data := range resp.Data {
		embeddings[data.Index] = data.Embedding
	}

	return embeddings, nil
}

// Dimension returns the embedding dimension.
func (g *OpenAIEmbeddingGenerator) Dimension() int {
	return g.dimension
}

// IsNoop returns false for OpenAI embedder.
func (g *OpenAIEmbeddingGenerator) IsNoop() bool {
	return false
}

// NewEmbeddingGenerator creates the appropriate embedding generator based on config.
func NewEmbeddingGenerator(cfg *config.RAGConfig, client HTTPClient) (EmbeddingGenerator, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("RAG is not enabled")
	}

	switch cfg.EmbeddingModel {
	case "openai":
		if cfg.APIKey == "" {
			return nil, fmt.Errorf("OpenAI API key required for openai embedding model")
		}
		return NewOpenAIEmbeddingGenerator(cfg, client), nil
	case "http", "embeddinggemma", "embeddinggemma-300m-qat":
		// Use HTTP embedding service
		return NewHTTPEmbeddingGenerator(cfg.APIBase, cfg.EmbeddingModel, cfg.Dimension), nil
	case "local":
		return NewLocalEmbeddingGenerator(cfg), nil
	case "none":
		// No-op provider for keyword-only search
		return NewNoopEmbeddingProvider(), nil
	case "":
		// Default to none (keyword-only) instead of local
		return NewNoopEmbeddingProvider(), nil
	default:
		// For custom or unknown models, use none as fallback
		return NewNoopEmbeddingProvider(), nil
	}
}

// HTTPEmbeddingGenerator calls a Python embedding service over HTTP.
type HTTPEmbeddingGenerator struct {
	baseURL   string
	model     string
	dimension int
	client    *http.Client
}

// httpEmbedRequest is the request body for the embedding service.
type httpEmbedRequest struct {
	Texts []string `json:"texts"`
	Model string   `json:"model"`
}

// httpEmbedResponse is the response from the embedding service.
type httpEmbedResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
	Dimension  int         `json:"dimension"`
	Model      string      `json:"model"`
}

// NewHTTPEmbeddingGenerator creates a new HTTP-based embedding generator.
func NewHTTPEmbeddingGenerator(baseURL, model string, dimension int) *HTTPEmbeddingGenerator {
	if baseURL == "" {
		baseURL = "http://localhost:8000"
	}
	if model == "" {
		model = "default"
	}
	if dimension == 0 {
		dimension = 384 // Default for MiniLM
	}

	return &HTTPEmbeddingGenerator{
		baseURL:   baseURL,
		model:     model,
		dimension: dimension,
		client:    &http.Client{Timeout: 30 * time.Second},
	}
}

// Generate creates an embedding by calling the HTTP service.
func (g *HTTPEmbeddingGenerator) Generate(text string) ([]float32, error) {
	embeddings, err := g.GenerateBatch([]string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embedding returned from HTTP service")
	}
	return embeddings[0], nil
}

// GenerateBatch creates embeddings for multiple texts via HTTP service.
func (g *HTTPEmbeddingGenerator) GenerateBatch(texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return [][]float32{}, nil
	}

	reqBody := httpEmbedRequest{
		Texts: texts,
		Model: g.model,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := g.client.Post(
		g.baseURL+"/embed",
		"application/json",
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to call embedding service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedding service returned status %d", resp.StatusCode)
	}

	var embedResp httpEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return embedResp.Embeddings, nil
}

// Dimension returns the embedding dimension.
func (g *HTTPEmbeddingGenerator) Dimension() int {
	return g.dimension
}

// IsNoop returns false for HTTP embedder.
func (g *HTTPEmbeddingGenerator) IsNoop() bool {
	return false
}

// Health checks if the embedding service is healthy.
func (g *HTTPEmbeddingGenerator) Health() (bool, error) {
	resp, err := g.client.Get(g.baseURL + "/health")
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

// SerializeEmbedding converts a float32 slice to bytes for storage.
func SerializeEmbedding(embedding []float32) []byte {
	buf := new(bytes.Buffer)
	for _, v := range embedding {
		binary.Write(buf, binary.LittleEndian, v)
	}
	return buf.Bytes()
}

// DeserializeEmbedding converts bytes back to a float32 slice.
func DeserializeEmbedding(data []byte, dimension int) ([]float32, error) {
	if len(data) != dimension*4 {
		return nil, fmt.Errorf("invalid embedding data length: expected %d, got %d", dimension*4, len(data))
	}

	embedding := make([]float32, dimension)
	buf := bytes.NewReader(data)
	for i := range embedding {
		if err := binary.Read(buf, binary.LittleEndian, &embedding[i]); err != nil {
			return nil, fmt.Errorf("failed to read embedding value: %w", err)
		}
	}
	return embedding, nil
}
