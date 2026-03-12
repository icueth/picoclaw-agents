package office

import (
	"encoding/json"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewSpriteManager(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSpriteManager(tmpDir)

	if sm == nil {
		t.Fatal("NewSpriteManager returned nil")
	}

	if sm.basePath != tmpDir {
		t.Errorf("Expected basePath to be %s, got %s", tmpDir, sm.basePath)
	}

	if sm.maxCacheSize != 100 {
		t.Errorf("Expected maxCacheSize to be 100, got %d", sm.maxCacheSize)
	}

	if sm.defaultSprite == nil {
		t.Error("Expected defaultSprite to be set")
	}
}

func TestSpriteManager_CreatePlaceholderSprite(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSpriteManager(tmpDir)

	sprite, err := sm.CreatePlaceholderSprite("test", 32, 32, color.RGBA{255, 0, 0, 255})
	if err != nil {
		t.Fatalf("CreatePlaceholderSprite failed: %v", err)
	}

	if sprite.Metadata.ID != "test" {
		t.Errorf("Expected ID to be 'test', got %s", sprite.Metadata.ID)
	}

	if sprite.Metadata.Width != 32 {
		t.Errorf("Expected width to be 32, got %d", sprite.Metadata.Width)
	}

	if sprite.Metadata.Height != 32 {
		t.Errorf("Expected height to be 32, got %d", sprite.Metadata.Height)
	}

	// Check file was created
	spritePath := filepath.Join(tmpDir, "test.png")
	if _, err := os.Stat(spritePath); os.IsNotExist(err) {
		t.Error("Expected sprite file to be created")
	}
}

func TestSpriteManager_GetSprite(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSpriteManager(tmpDir)

	// Create a test sprite
	_, err := sm.CreatePlaceholderSprite("test_sprite", 32, 32, color.RGBA{0, 255, 0, 255})
	if err != nil {
		t.Fatalf("Failed to create test sprite: %v", err)
	}

	// Clear cache to force reload
	sm.ClearCache()

	// Get the sprite
	sprite, err := sm.GetSprite("test_sprite")
	if err != nil {
		t.Fatalf("GetSprite failed: %v", err)
	}

	if sprite.Metadata.ID != "test_sprite" {
		t.Errorf("Expected ID to be 'test_sprite', got %s", sprite.Metadata.ID)
	}

	// Test cache hit
	sprite2, err := sm.GetSprite("test_sprite")
	if err != nil {
		t.Fatalf("GetSprite (cache hit) failed: %v", err)
	}

	if sprite != sprite2 {
		t.Error("Expected cached sprite to be the same object")
	}
}

func TestSpriteManager_GetSprite_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSpriteManager(tmpDir)

	// Try to get non-existent sprite
	sprite, err := sm.GetSprite("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent sprite")
	}

	// Should return default sprite
	if sprite == nil {
		t.Error("Expected default sprite on error")
	}
}

func TestSpriteManager_Cache(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSpriteManager(tmpDir)

	// Create test sprites
	for i := 0; i < 5; i++ {
		name := fmt.Sprintf("sprite_%d", i)
		_, err := sm.CreatePlaceholderSprite(name, 32, 32, color.RGBA{byte(i * 50), 0, 0, 255})
		if err != nil {
			t.Fatalf("Failed to create sprite %s: %v", name, err)
		}
	}

	// Load all sprites
	for i := 0; i < 5; i++ {
		name := fmt.Sprintf("sprite_%d", i)
		_, err := sm.GetSprite(name)
		if err != nil {
			t.Fatalf("Failed to get sprite %s: %v", name, err)
		}
	}

	// Check cache stats
	stats := sm.GetCacheStats()
	if stats["cached_sprites"] != 5 {
		t.Errorf("Expected 5 cached sprites, got %v", stats["cached_sprites"])
	}

	// Clear cache
	sm.ClearCache()

	stats = sm.GetCacheStats()
	if stats["cached_sprites"] != 0 {
		t.Errorf("Expected 0 cached sprites after clear, got %v", stats["cached_sprites"])
	}
}

func TestSpriteManager_AssignSpriteToAgent(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSpriteManager(tmpDir)

	// Create a test sprite
	_, err := sm.CreatePlaceholderSprite("agent_avatar", 32, 32, color.RGBA{0, 0, 255, 255})
	if err != nil {
		t.Fatalf("Failed to create test sprite: %v", err)
	}

	// Assign sprite to agent
	err = sm.AssignSpriteToAgent("agent_001", "agent_avatar")
	if err != nil {
		t.Fatalf("AssignSpriteToAgent failed: %v", err)
	}

	// Get sprite for agent
	sprite, err := sm.GetSpriteForAgent("agent_001")
	if err != nil {
		t.Fatalf("GetSpriteForAgent failed: %v", err)
	}

	if sprite.Metadata.ID != "agent_avatar" {
		t.Errorf("Expected sprite ID 'agent_avatar', got %s", sprite.Metadata.ID)
	}
}

func TestSpriteManager_ListSprites(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSpriteManager(tmpDir)

	// Create test sprites in subdirectory
	subDir := filepath.Join(tmpDir, "characters")
	os.MkdirAll(subDir, 0755)

	_, err := sm.CreatePlaceholderSprite("characters/hero", 32, 32, color.RGBA{255, 255, 0, 255})
	if err != nil {
		t.Fatalf("Failed to create test sprite: %v", err)
	}

	_, err = sm.CreatePlaceholderSprite("characters/villain", 32, 32, color.RGBA{255, 0, 255, 255})
	if err != nil {
		t.Fatalf("Failed to create test sprite: %v", err)
	}

	sprites, err := sm.ListSprites()
	if err != nil {
		t.Fatalf("ListSprites failed: %v", err)
	}

	if len(sprites) != 2 {
		t.Errorf("Expected 2 sprites, got %d", len(sprites))
	}
}

func TestSpriteManager_Animation(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSpriteManager(tmpDir)

	// Create a sprite with animation metadata
	sprite, err := sm.CreatePlaceholderSprite("animated", 32, 32, color.RGBA{0, 255, 255, 255})
	if err != nil {
		t.Fatalf("Failed to create test sprite: %v", err)
	}

	// Initialize Animations map and add animation definition
	sprite.Metadata.Animations = make(map[string]AnimationDefinition)
	sprite.Metadata.Animations["idle"] = AnimationDefinition{
		Name: "idle",
		Frames: []AnimationFrame{
			{SpriteID: "frame1", Duration: 250 * time.Millisecond},
			{SpriteID: "frame2", Duration: 250 * time.Millisecond},
		},
		Loop:      true,
		FrameRate: 4,
	}

	// Save the updated sprite with animations
	if err := sm.SaveSprite(sprite); err != nil {
		t.Fatalf("Failed to save sprite with animations: %v", err)
	}

	// Start animation
	err = sm.StartAnimation("animated", "idle", true)
	if err != nil {
		t.Fatalf("StartAnimation failed: %v", err)
	}

	// Check if animation is playing
	if !sm.IsAnimationPlaying("animated", "idle") {
		t.Error("Expected animation to be playing")
	}

	// Get animation frame
	frame, err := sm.GetAnimationFrame("animated", "idle")
	if err != nil {
		t.Fatalf("GetAnimationFrame failed: %v", err)
	}

	if frame == nil {
		t.Error("Expected animation frame")
	}

	// Stop animation
	sm.StopAnimation("animated", "idle")

	if sm.IsAnimationPlaying("animated", "idle") {
		t.Error("Expected animation to be stopped")
	}
}

func TestLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "sprites.json")

	config := SpriteConfig{
		BasePath:     tmpDir,
		MaxCacheSize: 50,
		CacheTimeout: 15,
		SpriteMappings: map[string]string{
			"agent1": "characters/agent1",
			"agent2": "characters/agent2",
		},
	}

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	loadedConfig, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if loadedConfig.MaxCacheSize != 50 {
		t.Errorf("Expected MaxCacheSize to be 50, got %d", loadedConfig.MaxCacheSize)
	}

	if loadedConfig.SpriteMappings["agent1"] != "characters/agent1" {
		t.Error("Expected sprite mapping to be loaded")
	}
}

func TestColorRGB_ToColor(t *testing.T) {
	c := ColorRGB{R: 255, G: 128, B: 64}
	col := c.ToColor()

	r, g, b, a := col.RGBA()
	// RGBA returns values in range [0, 65535]
	if r/257 != 255 {
		t.Errorf("Expected R to be 255, got %d", r/257)
	}
	if g/257 != 128 {
		t.Errorf("Expected G to be 128, got %d", g/257)
	}
	if b/257 != 64 {
		t.Errorf("Expected B to be 64, got %d", b/257)
	}
	if a/257 != 255 {
		t.Errorf("Expected A to be 255, got %d", a/257)
	}
}

func TestSpriteManager_CleanupCache(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSpriteManager(tmpDir)
	sm.cacheTimeout = 100 * time.Millisecond // Short timeout for testing

	// Create and load a sprite
	_, err := sm.CreatePlaceholderSprite("temp", 32, 32, color.RGBA{100, 100, 100, 255})
	if err != nil {
		t.Fatalf("Failed to create test sprite: %v", err)
	}

	sm.ClearCache()

	_, err = sm.GetSprite("temp")
	if err != nil {
		t.Fatalf("Failed to get sprite: %v", err)
	}

	// Wait for timeout
	time.Sleep(200 * time.Millisecond)

	// Cleanup should remove expired entry
	sm.CleanupCache()

	stats := sm.GetCacheStats()
	if stats["cached_sprites"] != 0 {
		t.Errorf("Expected 0 cached sprites after cleanup, got %v", stats["cached_sprites"])
	}
}

func TestSpriteManager_SaveSprite(t *testing.T) {
	tmpDir := t.TempDir()
	sm := NewSpriteManager(tmpDir)

	// Create a sprite
	sprite := sm.createPlaceholderSprite("test_save", 64, 64, color.RGBA{200, 100, 50, 255})

	// Save it
	err := sm.SaveSprite(sprite)
	if err != nil {
		t.Fatalf("SaveSprite failed: %v", err)
	}

	// Check both PNG and JSON files exist
	pngPath := filepath.Join(tmpDir, "test_save.png")
	jsonPath := filepath.Join(tmpDir, "test_save.json")

	if _, err := os.Stat(pngPath); os.IsNotExist(err) {
		t.Error("Expected PNG file to exist")
	}

	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		t.Error("Expected JSON metadata file to exist")
	}
}
