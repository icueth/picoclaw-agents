// Package office provides company-style workflow management for Picoclaw.
package office

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// SpriteLoader handles loading and serving sprites via HTTP
type SpriteLoader struct {
	manager    *SpriteManager
	config     *SpriteConfig
	basePath   string
	httpClient *http.Client
}

// NewSpriteLoader creates a new sprite loader
func NewSpriteLoader(basePath string) *SpriteLoader {
	return &SpriteLoader{
		manager:    NewSpriteManager(basePath),
		basePath:   basePath,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// NewSpriteLoaderWithConfig creates a sprite loader with configuration
func NewSpriteLoaderWithConfig(config *SpriteConfig) (*SpriteLoader, error) {
	loader := &SpriteLoader{
		manager:    NewSpriteManagerWithConfig(config),
		config:     config,
		basePath:   config.BasePath,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	return loader, nil
}

// LoadConfig loads sprite configuration from file
func (sl *SpriteLoader) LoadConfig(path string) error {
	config, err := LoadConfig(path)
	if err != nil {
		return err
	}

	sl.config = config
	sl.manager = NewSpriteManagerWithConfig(config)

	return nil
}

// GetSpriteHandler returns an HTTP handler for serving sprites
func (sl *SpriteLoader) GetSpriteHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract sprite ID from URL path
		// Expected format: /api/sprites/{sprite_id}
		// or: /api/sprites/characters/agent_name
		path := strings.TrimPrefix(r.URL.Path, "/api/sprites/")
		if path == "" {
			http.Error(w, "Sprite ID required", http.StatusBadRequest)
			return
		}

		spriteID := strings.TrimSuffix(path, ".png")

		// Get sprite
		sprite, err := sl.manager.GetSprite(spriteID)
		if err != nil {
			log.Printf("Error loading sprite %s: %v", spriteID, err)
			http.Error(w, "Sprite not found", http.StatusNotFound)
			return
		}

		// Set content type
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		w.Header().Set("Last-Modified", sprite.LoadedAt.Format(http.TimeFormat))

		// Write image data
		if _, err := w.Write(sprite.Data); err != nil {
			log.Printf("Error writing sprite data: %v", err)
		}
	}
}

// GetSpriteMetadataHandler returns an HTTP handler for serving sprite metadata
func (sl *SpriteLoader) GetSpriteMetadataHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract sprite ID from URL path
		path := strings.TrimPrefix(r.URL.Path, "/api/sprites/")
		path = strings.TrimSuffix(path, "/meta")

		if path == "" {
			// Return list of all sprites
			sprites, err := sl.manager.ListSprites()
			if err != nil {
				http.Error(w, "Failed to list sprites", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"sprites": sprites,
				"count":   len(sprites),
			})
			return
		}

		// Get specific sprite metadata
		sprite, err := sl.manager.GetSprite(path)
		if err != nil {
			http.Error(w, "Sprite not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sprite.Metadata)
	}
}

// GetAgentSpriteHandler returns an HTTP handler for serving agent-specific sprites
func (sl *SpriteLoader) GetAgentSpriteHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract agent ID from URL path
		// Expected format: /api/agents/{agent_id}/sprite
		path := strings.TrimPrefix(r.URL.Path, "/api/agents/")
		path = strings.TrimSuffix(path, "/sprite")

		if path == "" {
			http.Error(w, "Agent ID required", http.StatusBadRequest)
			return
		}

		agentID := path

		// Get sprite for agent
		sprite, err := sl.manager.GetSpriteForAgent(agentID)
		if err != nil {
			log.Printf("Error loading sprite for agent %s: %v", agentID, err)
			http.Error(w, "Sprite not found", http.StatusNotFound)
			return
		}

		// Set content type
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "public, max-age=3600")

		// Write image data
		if _, err := w.Write(sprite.Data); err != nil {
			log.Printf("Error writing sprite data: %v", err)
		}
	}
}

// GetCategorySpritesHandler returns an HTTP handler for listing sprites by category
func (sl *SpriteLoader) GetCategorySpritesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract category from URL path
		// Expected format: /api/sprites/category/{category}
		path := strings.TrimPrefix(r.URL.Path, "/api/sprites/category/")

		if path == "" {
			// Return list of categories
			categories := []map[string]string{
				{"id": "characters", "name": "Characters", "path": "characters"},
				{"id": "rooms", "name": "Rooms", "path": "rooms"},
				{"id": "furniture", "name": "Furniture", "path": "furniture"},
				{"id": "ui", "name": "UI Elements", "path": "ui"},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"categories": categories,
			})
			return
		}

		// Get sprites in category
		spriteType := SpriteType(path)
		sprites, err := sl.manager.GetSpriteByType(spriteType)
		if err != nil {
			http.Error(w, "Failed to list sprites", http.StatusInternalServerError)
			return
		}

		// Convert to metadata list
		metadata := make([]SpriteMetadata, len(sprites))
		for i, sprite := range sprites {
			metadata[i] = sprite.Metadata
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"category": path,
			"sprites":  metadata,
			"count":    len(metadata),
		})
	}
}

// GetCacheStatsHandler returns cache statistics
func (sl *SpriteLoader) GetCacheStatsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		stats := sl.manager.GetCacheStats()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	}
}

// ClearCacheHandler handles cache clearing requests
func (sl *SpriteLoader) ClearCacheHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		sl.manager.ClearCache()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "Cache cleared successfully",
		})
	}
}

// AssignSpriteHandler handles sprite assignment to agents
func (sl *SpriteLoader) AssignSpriteHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var request struct {
			AgentID  string `json:"agent_id"`
			SpriteID string `json:"sprite_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if request.AgentID == "" || request.SpriteID == "" {
			http.Error(w, "agent_id and sprite_id are required", http.StatusBadRequest)
			return
		}

		if err := sl.manager.AssignSpriteToAgent(request.AgentID, request.SpriteID); err != nil {
			log.Printf("Error assigning sprite: %v", err)
			http.Error(w, "Failed to assign sprite", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":    "success",
			"agent_id":  request.AgentID,
			"sprite_id": request.SpriteID,
		})
	}
}

// UploadSpriteHandler handles sprite uploads
func (sl *SpriteLoader) UploadSpriteHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse multipart form
		if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB max
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		// Get file from form
		file, header, err := r.FormFile("sprite")
		if err != nil {
			http.Error(w, "Failed to get file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Validate file type
		contentType := header.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "image/") {
			http.Error(w, "Invalid file type", http.StatusBadRequest)
			return
		}

		// Get sprite ID from form
		spriteID := r.FormValue("sprite_id")
		if spriteID == "" {
			http.Error(w, "sprite_id is required", http.StatusBadRequest)
			return
		}

		// Read file data
		data, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Failed to read file", http.StatusInternalServerError)
			return
		}

		// Save sprite
		if err := sl.saveSpriteToDisk(spriteID, data); err != nil {
			log.Printf("Error saving sprite: %v", err)
			http.Error(w, "Failed to save sprite", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":    "success",
			"sprite_id": spriteID,
			"message":   "Sprite uploaded successfully",
		})
	}
}

// saveSpriteToDisk saves sprite data to the filesystem
func (sl *SpriteLoader) saveSpriteToDisk(spriteID string, data []byte) error {
	// Create directory if needed
	spritePath := filepath.Join(sl.basePath, spriteID+".png")
	dir := filepath.Dir(spritePath)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Save file
	if err := os.WriteFile(spritePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// DownloadSprite downloads a sprite from a remote URL
func (sl *SpriteLoader) DownloadSprite(spriteID, url string) error {
	resp, err := sl.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download sprite: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download sprite: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read sprite data: %w", err)
	}

	return sl.saveSpriteToDisk(spriteID, data)
}

// RegisterRoutes registers all sprite-related HTTP routes
func (sl *SpriteLoader) RegisterRoutes(mux *http.ServeMux) {
	// Sprite serving endpoints
	mux.HandleFunc("/api/sprites/", sl.GetSpriteHandler())
	mux.HandleFunc("/api/sprites/meta", sl.GetSpriteMetadataHandler())
	mux.HandleFunc("/api/sprites/category/", sl.GetCategorySpritesHandler())

	// Agent sprite endpoints
	mux.HandleFunc("/api/agents/", sl.GetAgentSpriteHandler())
	mux.HandleFunc("/api/agents/sprite/assign", sl.AssignSpriteHandler())

	// Cache management
	mux.HandleFunc("/api/sprites/cache/stats", sl.GetCacheStatsHandler())
	mux.HandleFunc("/api/sprites/cache/clear", sl.ClearCacheHandler())

	// Upload endpoint
	mux.HandleFunc("/api/sprites/upload", sl.UploadSpriteHandler())
}

// ServeSprites starts an HTTP server to serve sprites
func (sl *SpriteLoader) ServeSprites(addr string) error {
	mux := http.NewServeMux()
	sl.RegisterRoutes(mux)

	// Add CORS middleware
	handler := corsMiddleware(mux)

	log.Printf("Starting sprite server on %s", addr)
	return http.ListenAndServe(addr, handler)
}

// corsMiddleware adds CORS headers to responses
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GetManager returns the underlying sprite manager
func (sl *SpriteLoader) GetManager() *SpriteManager {
	return sl.manager
}

// GetConfig returns the current sprite configuration
func (sl *SpriteLoader) GetConfig() *SpriteConfig {
	return sl.config
}

// SpriteBatchLoader handles batch loading of sprites
type SpriteBatchLoader struct {
	loader    *SpriteLoader
	batchSize int
}

// NewSpriteBatchLoader creates a new batch loader
func NewSpriteBatchLoader(loader *SpriteLoader, batchSize int) *SpriteBatchLoader {
	if batchSize <= 0 {
		batchSize = 10
	}

	return &SpriteBatchLoader{
		loader:    loader,
		batchSize: batchSize,
	}
}

// LoadBatch loads a batch of sprites asynchronously
func (sbl *SpriteBatchLoader) LoadBatch(spriteIDs []string) <-chan BatchResult {
	results := make(chan BatchResult, len(spriteIDs))

	go func() {
		defer close(results)

		semaphore := make(chan struct{}, sbl.batchSize)

		for _, id := range spriteIDs {
			semaphore <- struct{}{} // Acquire

			go func(spriteID string) {
				defer func() { <-semaphore }() // Release

				sprite, err := sbl.loader.manager.GetSprite(spriteID)
				results <- BatchResult{
					SpriteID: spriteID,
					Sprite:   sprite,
					Error:    err,
				}
			}(id)
		}

		// Wait for all to complete
		for i := 0; i < cap(semaphore); i++ {
			semaphore <- struct{}{}
		}
	}()

	return results
}

// BatchResult represents the result of a batch load operation
type BatchResult struct {
	SpriteID string
	Sprite   *Sprite
	Error    error
}

// SpritePreloader handles preloading sprites for better performance
type SpritePreloader struct {
	loader      *SpriteLoader
	toPreload   []string
	preloaded   map[string]bool
	mu          sync.RWMutex
}

// NewSpritePreloader creates a new sprite preloader
func NewSpritePreloader(loader *SpriteLoader) *SpritePreloader {
	return &SpritePreloader{
		loader:    loader,
		toPreload: make([]string, 0),
		preloaded: make(map[string]bool),
	}
}

// Add adds sprites to the preload queue
func (sp *SpritePreloader) Add(spriteIDs ...string) {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	for _, id := range spriteIDs {
		if !sp.preloaded[id] {
			sp.toPreload = append(sp.toPreload, id)
		}
	}
}

// Preload loads all queued sprites
func (sp *SpritePreloader) Preload() error {
	sp.mu.Lock()
	toLoad := make([]string, len(sp.toPreload))
	copy(toLoad, sp.toPreload)
	sp.toPreload = sp.toPreload[:0] // Clear queue
	sp.mu.Unlock()

	for _, id := range toLoad {
		_, err := sp.loader.manager.GetSprite(id)
		if err != nil {
			return fmt.Errorf("failed to preload sprite %s: %w", id, err)
		}

		sp.mu.Lock()
		sp.preloaded[id] = true
		sp.mu.Unlock()
	}

	return nil
}

// IsPreloaded checks if a sprite has been preloaded
func (sp *SpritePreloader) IsPreloaded(spriteID string) bool {
	sp.mu.RLock()
	defer sp.mu.RUnlock()

	return sp.preloaded[spriteID]
}

// GetPreloadStatus returns preload statistics
func (sp *SpritePreloader) GetPreloadStatus() map[string]interface{} {
	sp.mu.RLock()
	defer sp.mu.RUnlock()

	return map[string]interface{}{
		"preloaded_count": len(sp.preloaded),
		"queued_count":    len(sp.toPreload),
		"preloaded":       sp.preloaded,
	}
}

// SpriteValidator validates sprite files
type SpriteValidator struct {
	maxFileSize int64
	validTypes  []string
}

// NewSpriteValidator creates a new sprite validator
func NewSpriteValidator() *SpriteValidator {
	return &SpriteValidator{
		maxFileSize: 10 * 1024 * 1024, // 10 MB
		validTypes:  []string{"image/png", "image/jpeg"},
	}
}

// ValidateFile validates a sprite file
func (sv *SpriteValidator) ValidateFile(data []byte, filename string) error {
	// Check file size
	if int64(len(data)) > sv.maxFileSize {
		return fmt.Errorf("file too large: %d bytes (max %d)", len(data), sv.maxFileSize)
	}

	// Detect content type
	contentType := http.DetectContentType(data)

	// Check if valid type
	valid := false
	for _, t := range sv.validTypes {
		if contentType == t {
			valid = true
			break
		}
	}

	if !valid {
		return fmt.Errorf("invalid file type: %s", contentType)
	}

	// Validate image dimensions (basic check)
	if _, err := png.Decode(bytes.NewReader(data)); err != nil {
		return fmt.Errorf("invalid image data: %w", err)
	}

	return nil
}

// SetMaxFileSize sets the maximum allowed file size
func (sv *SpriteValidator) SetMaxFileSize(size int64) {
	sv.maxFileSize = size
}

// AddValidType adds a valid MIME type
func (sv *SpriteValidator) AddValidType(mimeType string) {
	sv.validTypes = append(sv.validTypes, mimeType)
}

// SpriteCacheWarmer warms up the cache with frequently used sprites
type SpriteCacheWarmer struct {
	loader      *SpriteLoader
	frequentlyUsed []string
	warmInterval   time.Duration
	stopChan       chan struct{}
}

// NewSpriteCacheWarmer creates a new cache warmer
func NewSpriteCacheWarmer(loader *SpriteLoader) *SpriteCacheWarmer {
	return &SpriteCacheWarmer{
		loader:         loader,
		frequentlyUsed: make([]string, 0),
		warmInterval:   5 * time.Minute,
		stopChan:       make(chan struct{}),
	}
}

// SetFrequentlyUsed sets the list of frequently used sprites
func (scw *SpriteCacheWarmer) SetFrequentlyUsed(spriteIDs []string) {
	scw.frequentlyUsed = spriteIDs
}

// SetWarmInterval sets the interval between cache warming cycles
func (scw *SpriteCacheWarmer) SetWarmInterval(interval time.Duration) {
	scw.warmInterval = interval
}

// Start begins the cache warming routine
func (scw *SpriteCacheWarmer) Start() {
	go func() {
		ticker := time.NewTicker(scw.warmInterval)
		defer ticker.Stop()

		// Warm immediately on start
		scw.warm()

		for {
			select {
			case <-ticker.C:
				scw.warm()
			case <-scw.stopChan:
				return
			}
		}
	}()
}

// Stop stops the cache warming routine
func (scw *SpriteCacheWarmer) Stop() {
	close(scw.stopChan)
}

// warm loads all frequently used sprites into cache
func (scw *SpriteCacheWarmer) warm() {
	for _, id := range scw.frequentlyUsed {
		_, err := scw.loader.manager.GetSprite(id)
		if err != nil {
			log.Printf("Cache warmer: failed to load sprite %s: %v", id, err)
		}
	}
}

// WarmOnce performs a single cache warming cycle
func (scw *SpriteCacheWarmer) WarmOnce() {
	scw.warm()
}
