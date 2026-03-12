package agent

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"picoclaw/agent/pkg/logger"
)

// ModelManager handles downloading and managing embedding models
type ModelManager struct {
	modelsDir string
	client    *http.Client
}

// ModelInfo represents information about a downloadable model
type ModelInfo struct {
	Name        string            `json:"name"`
	URL         string            `json:"url"`
	Size        int64             `json:"size"`        // size in bytes
	SHA256      string            `json:"sha256"`      // checksum for verification
	Dimension   int               `json:"dimension"`   // embedding dimension
	MaxSeqLen   int               `json:"max_seq_len"` // max sequence length
	Languages   []string          `json:"languages"`   // supported languages
	Description string            `json:"description"`
}

// Predefined models available for download
var AvailableModels = map[string]ModelInfo{
	"paraphrase-multilingual-MiniLM-L12-v2": {
		Name:        "paraphrase-multilingual-MiniLM-L12-v2",
		URL:         "https://huggingface.co/sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2/resolve/main/onnx/model.onnx",
		Size:        120000000, // ~120MB (approximate)
		Dimension:   384,
		MaxSeqLen:   128,
		Languages:   []string{"en", "zh", "th", "ja", "ko", "de", "fr", "es", "it", "pt", "ar", "hi"},
		Description: "Multilingual sentence embeddings optimized for semantic similarity",
	},
}

// NewModelManager creates a new model manager
func NewModelManager(modelsDir string) *ModelManager {
	if modelsDir == "" {
		homeDir, _ := os.UserHomeDir()
		modelsDir = filepath.Join(homeDir, ".picoclaw", "models")
	}

	return &ModelManager{
		modelsDir: modelsDir,
		client: &http.Client{
			Timeout: 10 * time.Minute, // Models can be large
		},
	}
}

// EnsureModel ensures the specified model is available locally
// Downloads if necessary
func (mm *ModelManager) EnsureModel(modelName string) (string, error) {
	modelInfo, ok := AvailableModels[modelName]
	if !ok {
		return "", fmt.Errorf("unknown model: %s", modelName)
	}

	modelPath := mm.GetModelPath(modelName)

	// Check if model already exists
	if _, err := os.Stat(modelPath); err == nil {
		logger.InfoCF("model_manager", "Model already exists", map[string]any{
			"model": modelName,
			"path":  modelPath,
		})
		return modelPath, nil
	}

	// Download the model
	logger.InfoCF("model_manager", "Downloading model", map[string]any{
		"model": modelName,
		"url":   modelInfo.URL,
		"size":  formatBytes(modelInfo.Size),
	})

	if err := mm.downloadModel(modelInfo, modelPath); err != nil {
		return "", fmt.Errorf("failed to download model: %w", err)
	}

	logger.InfoCF("model_manager", "Model downloaded successfully", map[string]any{
		"model": modelName,
		"path":  modelPath,
	})

	return modelPath, nil
}

// GetModelPath returns the local path for a model
func (mm *ModelManager) GetModelPath(modelName string) string {
	return filepath.Join(mm.modelsDir, modelName+".onnx")
}

// ListAvailableModels returns list of available models that can be downloaded
func (mm *ModelManager) ListAvailableModels() []ModelInfo {
	models := make([]ModelInfo, 0, len(AvailableModels))
	for _, info := range AvailableModels {
		models = append(models, info)
	}
	return models
}

// ListLocalModels returns list of models that are already downloaded
func (mm *ModelManager) ListLocalModels() ([]string, error) {
	entries, err := os.ReadDir(mm.modelsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	models := make([]string, 0)
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".onnx" {
			modelName := entry.Name()[:len(entry.Name())-5] // remove .onnx
			models = append(models, modelName)
		}
	}

	return models, nil
}

// RemoveModel removes a downloaded model
func (mm *ModelManager) RemoveModel(modelName string) error {
	modelPath := mm.GetModelPath(modelName)
	if err := os.Remove(modelPath); err != nil {
		return fmt.Errorf("failed to remove model: %w", err)
	}
	logger.InfoCF("model_manager", "Model removed", map[string]any{
		"model": modelName,
	})
	return nil
}

// downloadModel downloads a model from URL to the specified path
func (mm *ModelManager) downloadModel(info ModelInfo, destPath string) error {
	// Create models directory if it doesn't exist
	if err := os.MkdirAll(mm.modelsDir, 0755); err != nil {
		return fmt.Errorf("failed to create models directory: %w", err)
	}

	// Create temporary file for download
	tmpPath := destPath + ".tmp"
	file, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpPath) // Clean up on error

	// Download with progress tracking
	resp, err := mm.client.Get(info.URL)
	if err != nil {
		file.Close()
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		file.Close()
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	// Copy with progress reporting
	written, err := io.Copy(file, resp.Body)
	file.Close()
	if err != nil {
		return fmt.Errorf("failed to write model file: %w", err)
	}

	logger.InfoCF("model_manager", "Download complete", map[string]any{
		"bytes_written": written,
		"size_mb":       float64(written) / 1024 / 1024,
	})

	// Rename temp file to final destination
	if err := os.Rename(tmpPath, destPath); err != nil {
		return fmt.Errorf("failed to finalize model file: %w", err)
	}

	return nil
}

// formatBytes formats byte count to human readable string
func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// GetModelInfo returns information about a specific model
func GetModelInfo(modelName string) (*ModelInfo, error) {
	info, ok := AvailableModels[modelName]
	if !ok {
		return nil, fmt.Errorf("model not found: %s", modelName)
	}
	return &info, nil
}

// ValidateModelPath checks if a model file exists and is valid
func ValidateModelPath(path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("model file does not exist: %s", path)
		}
		return fmt.Errorf("cannot access model file: %w", err)
	}

	if stat.IsDir() {
		return fmt.Errorf("model path is a directory: %s", path)
	}

	if stat.Size() < 1024*1024 { // Less than 1MB is suspicious
		return fmt.Errorf("model file seems too small (%s): %s", formatBytes(stat.Size()), path)
	}

	return nil
}
