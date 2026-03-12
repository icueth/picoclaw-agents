// Package office provides company-style workflow management for Picoclaw.
package office

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// SpriteType represents the category of a sprite
type SpriteType string

const (
	// SpriteTypeCharacter represents agent avatars
	SpriteTypeCharacter SpriteType = "character"
	// SpriteTypeRoom represents room backgrounds
	SpriteTypeRoom SpriteType = "room"
	// SpriteTypeFurniture represents office furniture
	SpriteTypeFurniture SpriteType = "furniture"
	// SpriteTypeUI represents UI elements
	SpriteTypeUI SpriteType = "ui"
)

// AnimationFrame represents a single frame in an animation
type AnimationFrame struct {
	SpriteID  string        `json:"sprite_id"`
	Duration  time.Duration `json:"duration_ms"`
	OffsetX   int           `json:"offset_x"`
	OffsetY   int           `json:"offset_y"`
}

// AnimationDefinition defines an animation sequence
type AnimationDefinition struct {
	Name        string           `json:"name"`
	Frames      []AnimationFrame `json:"frames"`
	Loop        bool             `json:"loop"`
	FrameRate   int              `json:"frame_rate"` // Frames per second
	TotalFrames int              `json:"total_frames"`
}

// SpriteMetadata contains metadata about a sprite
type SpriteMetadata struct {
	ID          string                       `json:"id"`
	Name        string                       `json:"name"`
	Type        SpriteType                   `json:"type"`
	Width       int                          `json:"width"`
	Height      int                          `json:"height"`
	Category    string                       `json:"category"`
	Tags        []string                     `json:"tags"`
	Animations  map[string]AnimationDefinition `json:"animations"`
	DefaultAnim string                       `json:"default_animation"`
	ColorKey    *ColorRGB                    `json:"color_key,omitempty"` // Transparent color
}

// ColorRGB represents an RGB color
type ColorRGB struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
}

// ToColor converts ColorRGB to color.Color
func (c *ColorRGB) ToColor() color.Color {
	return color.RGBA{R: c.R, G: c.G, B: c.B, A: 255}
}

// Sprite represents a loaded sprite with its image data and metadata
type Sprite struct {
	Metadata SpriteMetadata `json:"metadata"`
	Image    image.Image    `json:"-"` // Image data (not serialized)
	Data     []byte         `json:"-"` // Raw PNG data
	LoadedAt time.Time      `json:"loaded_at"`
}

// SpriteCacheEntry represents a cached sprite with access tracking
type SpriteCacheEntry struct {
	Sprite     *Sprite
	LastAccess time.Time
	AccessCount int
}

// SpriteManager manages sprite loading, caching, and animation
type SpriteManager struct {
	sprites       map[string]*SpriteCacheEntry
	animations    map[string]*ActiveAnimation
	config        *SpriteConfig
	basePath      string
	mu            sync.RWMutex
	maxCacheSize  int
	cacheTimeout  time.Duration
	defaultSprite *Sprite
}

// ActiveAnimation represents a currently playing animation
type ActiveAnimation struct {
	SpriteID      string
	AnimationName string
	CurrentFrame  int
	StartTime     time.Time
	LastFrameTime time.Time
	IsPlaying     bool
	Loop          bool
}

// SpriteConfig contains configuration for the sprite system
type SpriteConfig struct {
	BasePath      string                     `json:"base_path"`
	MaxCacheSize  int                        `json:"max_cache_size"`
	CacheTimeout  int                        `json:"cache_timeout_minutes"`
	SpriteMappings map[string]string          `json:"sprite_mappings"` // agent_id -> sprite_id
	Categories    map[string]SpriteCategory  `json:"categories"`
}

// SpriteCategory defines a category of sprites
type SpriteCategory struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Path        string   `json:"path"`
	DefaultTags []string `json:"default_tags"`
}

// NewSpriteManager creates a new sprite manager
func NewSpriteManager(basePath string) *SpriteManager {
	sm := &SpriteManager{
		sprites:      make(map[string]*SpriteCacheEntry),
		animations:   make(map[string]*ActiveAnimation),
		basePath:     basePath,
		maxCacheSize: 100,
		cacheTimeout: 30 * time.Minute,
	}

	// Create default placeholder sprite
	sm.defaultSprite = sm.createPlaceholderSprite("default", 32, 32, color.RGBA{128, 128, 128, 255})

	return sm
}

// NewSpriteManagerWithConfig creates a sprite manager with configuration
func NewSpriteManagerWithConfig(config *SpriteConfig) *SpriteManager {
	sm := NewSpriteManager(config.BasePath)
	sm.config = config

	if config.MaxCacheSize > 0 {
		sm.maxCacheSize = config.MaxCacheSize
	}

	if config.CacheTimeout > 0 {
		sm.cacheTimeout = time.Duration(config.CacheTimeout) * time.Minute
	}

	return sm
}

// LoadConfig loads sprite configuration from a JSON file
func LoadConfig(path string) (*SpriteConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read sprite config: %w", err)
	}

	var config SpriteConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse sprite config: %w", err)
	}

	// Set defaults
	if config.MaxCacheSize == 0 {
		config.MaxCacheSize = 100
	}

	if config.CacheTimeout == 0 {
		config.CacheTimeout = 30
	}

	return &config, nil
}

// GetSprite retrieves a sprite by ID, loading it if necessary
func (sm *SpriteManager) GetSprite(spriteID string) (*Sprite, error) {
	// Check cache first
	sm.mu.RLock()
	if entry, ok := sm.sprites[spriteID]; ok {
		entry.LastAccess = time.Now()
		entry.AccessCount++
		sm.mu.RUnlock()
		return entry.Sprite, nil
	}
	sm.mu.RUnlock()

	// Load from disk
	sprite, err := sm.loadSpriteFromDisk(spriteID)
	if err != nil {
		// Return default sprite on error
		return sm.defaultSprite, fmt.Errorf("failed to load sprite %s: %w", spriteID, err)
	}

	// Add to cache
	sm.addToCache(spriteID, sprite)

	return sprite, nil
}

// GetSpriteForAgent retrieves the sprite assigned to an agent
func (sm *SpriteManager) GetSpriteForAgent(agentID string) (*Sprite, error) {
	// Check if there's a mapping for this agent
	if sm.config != nil && sm.config.SpriteMappings != nil {
		if spriteID, ok := sm.config.SpriteMappings[agentID]; ok {
			return sm.GetSprite(spriteID)
		}
	}

	// Try to load agent-specific sprite
	spriteID := fmt.Sprintf("characters/%s", agentID)
	sprite, err := sm.GetSprite(spriteID)
	if err == nil && sprite != sm.defaultSprite {
		return sprite, nil
	}

	// Return default character sprite
	return sm.GetSprite("characters/default")
}

// GetSpriteByType retrieves sprites by type
func (sm *SpriteManager) GetSpriteByType(spriteType SpriteType) ([]*Sprite, error) {
	var sprites []*Sprite

	// Scan the directory for sprites of this type
	typePath := filepath.Join(sm.basePath, string(spriteType))

	entries, err := os.ReadDir(typePath)
	if err != nil {
		return sprites, fmt.Errorf("failed to read sprite directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Check if it's a PNG file
		if filepath.Ext(entry.Name()) == ".png" {
			spriteID := string(spriteType) + "/" + entry.Name()[:len(entry.Name())-4]
			sprite, err := sm.GetSprite(spriteID)
			if err == nil {
				sprites = append(sprites, sprite)
			}
		}
	}

	return sprites, nil
}

// loadSpriteFromDisk loads a sprite from the filesystem
func (sm *SpriteManager) loadSpriteFromDisk(spriteID string) (*Sprite, error) {
	// Construct file path
	filePath := filepath.Join(sm.basePath, spriteID+".png")

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("sprite file not found: %s", filePath)
	}

	// Read image data
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read sprite file: %w", err)
	}

	// Decode image
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode sprite image: %w", err)
	}

	// Load metadata if available
	metadata := sm.loadMetadata(spriteID)
	if metadata.ID == "" {
		// Create default metadata
		bounds := img.Bounds()
		metadata = SpriteMetadata{
			ID:       spriteID,
			Name:     filepath.Base(spriteID),
			Width:    bounds.Dx(),
			Height:   bounds.Dy(),
			Category: filepath.Dir(spriteID),
		}
	}

	sprite := &Sprite{
		Metadata: metadata,
		Image:    img,
		Data:     data,
		LoadedAt: time.Now(),
	}

	return sprite, nil
}

// loadMetadata loads sprite metadata from JSON file
func (sm *SpriteManager) loadMetadata(spriteID string) SpriteMetadata {
	metaPath := filepath.Join(sm.basePath, spriteID+".json")

	data, err := os.ReadFile(metaPath)
	if err != nil {
		return SpriteMetadata{}
	}

	var metadata SpriteMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return SpriteMetadata{}
	}

	return metadata
}

// addToCache adds a sprite to the cache
func (sm *SpriteManager) addToCache(spriteID string, sprite *Sprite) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check if cache is full
	if len(sm.sprites) >= sm.maxCacheSize {
		sm.evictOldest()
	}

	sm.sprites[spriteID] = &SpriteCacheEntry{
		Sprite:      sprite,
		LastAccess:  time.Now(),
		AccessCount: 1,
	}
}

// evictOldest removes the least recently used sprite from cache
func (sm *SpriteManager) evictOldest() {
	var oldestID string
	var oldestTime time.Time

	for id, entry := range sm.sprites {
		if oldestTime.IsZero() || entry.LastAccess.Before(oldestTime) {
			oldestTime = entry.LastAccess
			oldestID = id
		}
	}

	if oldestID != "" {
		delete(sm.sprites, oldestID)
	}
}

// ClearCache clears all cached sprites
func (sm *SpriteManager) ClearCache() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.sprites = make(map[string]*SpriteCacheEntry)
}

// GetCacheStats returns cache statistics
func (sm *SpriteManager) GetCacheStats() map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	totalAccesses := 0
	for _, entry := range sm.sprites {
		totalAccesses += entry.AccessCount
	}

	return map[string]interface{}{
		"cached_sprites":   len(sm.sprites),
		"max_cache_size":   sm.maxCacheSize,
		"total_accesses":   totalAccesses,
		"cache_timeout":    sm.cacheTimeout.String(),
	}
}

// StartAnimation starts an animation for a sprite
func (sm *SpriteManager) StartAnimation(spriteID, animationName string, loop bool) error {
	sprite, err := sm.GetSprite(spriteID)
	if err != nil {
		return err
	}

	animation, ok := sprite.Metadata.Animations[animationName]
	if !ok {
		return fmt.Errorf("animation %s not found for sprite %s", animationName, spriteID)
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	key := fmt.Sprintf("%s:%s", spriteID, animationName)
	sm.animations[key] = &ActiveAnimation{
		SpriteID:      spriteID,
		AnimationName: animationName,
		CurrentFrame:  0,
		StartTime:     time.Now(),
		LastFrameTime: time.Now(),
		IsPlaying:     true,
		Loop:          loop || animation.Loop,
	}

	return nil
}

// StopAnimation stops an animation
func (sm *SpriteManager) StopAnimation(spriteID, animationName string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	key := fmt.Sprintf("%s:%s", spriteID, animationName)
	delete(sm.animations, key)
}

// GetAnimationFrame gets the current frame for an active animation
func (sm *SpriteManager) GetAnimationFrame(spriteID, animationName string) (*AnimationFrame, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	key := fmt.Sprintf("%s:%s", spriteID, animationName)
	anim, ok := sm.animations[key]
	if !ok {
		return nil, fmt.Errorf("animation not active: %s", key)
	}

	sprite, err := sm.GetSprite(spriteID)
	if err != nil {
		return nil, err
	}

	animationDef, ok := sprite.Metadata.Animations[animationName]
	if !ok {
		return nil, fmt.Errorf("animation definition not found: %s", animationName)
	}

	if len(animationDef.Frames) == 0 {
		return nil, fmt.Errorf("animation has no frames: %s", animationName)
	}

	// Calculate current frame based on time
	elapsed := time.Since(anim.LastFrameTime)
	frameDuration := time.Second / time.Duration(animationDef.FrameRate)

	if elapsed >= frameDuration {
		anim.CurrentFrame++
		anim.LastFrameTime = time.Now()

		// Handle loop or end
		if anim.CurrentFrame >= len(animationDef.Frames) {
			if anim.Loop {
				anim.CurrentFrame = 0
			} else {
				anim.CurrentFrame = len(animationDef.Frames) - 1
				anim.IsPlaying = false
			}
		}
	}

	return &animationDef.Frames[anim.CurrentFrame], nil
}

// IsAnimationPlaying checks if an animation is currently playing
func (sm *SpriteManager) IsAnimationPlaying(spriteID, animationName string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	key := fmt.Sprintf("%s:%s", spriteID, animationName)
	anim, ok := sm.animations[key]
	return ok && anim.IsPlaying
}

// GetActiveAnimations returns all currently active animations
func (sm *SpriteManager) GetActiveAnimations() []ActiveAnimation {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	animations := make([]ActiveAnimation, 0, len(sm.animations))
	for _, anim := range sm.animations {
		animations = append(animations, *anim)
	}

	return animations
}

// createPlaceholderSprite creates a simple placeholder sprite
func (sm *SpriteManager) createPlaceholderSprite(id string, width, height int, c color.Color) *Sprite {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: c}, image.Point{}, draw.Src)

	// Add a border
	borderColor := color.RGBA{255, 255, 255, 255}
	for x := 0; x < width; x++ {
		img.Set(x, 0, borderColor)
		img.Set(x, height-1, borderColor)
	}
	for y := 0; y < height; y++ {
		img.Set(0, y, borderColor)
		img.Set(width-1, y, borderColor)
	}

	// Encode to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		// Return nil on error - this should not happen
		return nil
	}

	return &Sprite{
		Metadata: SpriteMetadata{
			ID:       id,
			Name:     id,
			Type:     SpriteTypeUI,
			Width:    width,
			Height:   height,
			Category: "placeholder",
		},
		Image:    img,
		Data:     buf.Bytes(),
		LoadedAt: time.Now(),
	}
}

// CreatePlaceholderSprite creates a colored placeholder sprite
func (sm *SpriteManager) CreatePlaceholderSprite(id string, width, height int, c color.Color) (*Sprite, error) {
	sprite := sm.createPlaceholderSprite(id, width, height, c)
	if sprite == nil {
		return nil, fmt.Errorf("failed to create placeholder sprite")
	}

	// Save to disk for future use
	if err := sm.SaveSprite(sprite); err != nil {
		return nil, fmt.Errorf("failed to save placeholder sprite: %w", err)
	}

	return sprite, nil
}

// SaveSprite saves a sprite to disk
func (sm *SpriteManager) SaveSprite(sprite *Sprite) error {
	// Create directory if needed
	spritePath := filepath.Join(sm.basePath, sprite.Metadata.ID+".png")
	dir := filepath.Dir(spritePath)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create sprite directory: %w", err)
	}

	// Save image
	if err := os.WriteFile(spritePath, sprite.Data, 0644); err != nil {
		return fmt.Errorf("failed to write sprite file: %w", err)
	}

	// Save metadata
	metaPath := filepath.Join(sm.basePath, sprite.Metadata.ID+".json")
	metaData, err := json.MarshalIndent(sprite.Metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal sprite metadata: %w", err)
	}

	if err := os.WriteFile(metaPath, metaData, 0644); err != nil {
		return fmt.Errorf("failed to write sprite metadata: %w", err)
	}

	return nil
}

// ListSprites returns a list of all available sprites
func (sm *SpriteManager) ListSprites() ([]SpriteMetadata, error) {
	var sprites []SpriteMetadata

	// Walk the sprite directory
	err := filepath.Walk(sm.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || filepath.Ext(path) != ".png" {
			return nil
		}

		// Get relative path as sprite ID
		relPath, err := filepath.Rel(sm.basePath, path)
		if err != nil {
			return err
		}

		spriteID := relPath[:len(relPath)-4] // Remove .png

		// Try to load metadata
		metadata := sm.loadMetadata(spriteID)
		if metadata.ID == "" {
			// Create basic metadata
			metadata.ID = spriteID
			metadata.Name = filepath.Base(spriteID)
			metadata.Category = filepath.Dir(spriteID)
		}

		sprites = append(sprites, metadata)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list sprites: %w", err)
	}

	return sprites, nil
}

// AssignSpriteToAgent assigns a sprite to an agent
func (sm *SpriteManager) AssignSpriteToAgent(agentID, spriteID string) error {
	if sm.config == nil {
		sm.config = &SpriteConfig{
			BasePath:       sm.basePath,
			SpriteMappings: make(map[string]string),
		}
	}

	if sm.config.SpriteMappings == nil {
		sm.config.SpriteMappings = make(map[string]string)
	}

	sm.config.SpriteMappings[agentID] = spriteID

	// Save config
	if err := sm.SaveConfig(); err != nil {
		return fmt.Errorf("failed to save sprite assignment: %w", err)
	}

	return nil
}

// SaveConfig saves the current sprite configuration
func (sm *SpriteManager) SaveConfig() error {
	if sm.config == nil {
		return fmt.Errorf("no configuration to save")
	}

	configPath := filepath.Join(sm.basePath, "sprites.json")
	data, err := json.MarshalIndent(sm.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// CleanupCache removes expired entries from cache
func (sm *SpriteManager) CleanupCache() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	for id, entry := range sm.sprites {
		if now.Sub(entry.LastAccess) > sm.cacheTimeout {
			delete(sm.sprites, id)
		}
	}
}

// StartCleanupRoutine starts a background routine to clean up expired cache entries
func (sm *SpriteManager) StartCleanupRoutine(interval time.Duration) chan struct{} {
	stop := make(chan struct{})

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				sm.CleanupCache()
			case <-stop:
				return
			}
		}
	}()

	return stop
}
