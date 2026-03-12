// Package embeddingservice manages the Python embedding service lifecycle
//
// DEPRECATED: This package is deprecated. PicoClaw now uses keyword-only search
// by default (embedding_provider = "none"). Embedding service is only needed
// if you explicitly set embedding_provider to "http" in your config.
//
// The new hybrid search system uses SQLite FTS5 for keyword search without
// requiring any external services or models.
package embeddingservice

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"picoclaw/agent/pkg/config"
	"picoclaw/agent/pkg/logger"
)

// Manager handles the lifecycle of the Python embedding service
//
// DEPRECATED: Embedding service is no longer required for default operation.
// The manager now returns immediately without starting any services when
// embedding_provider is "none" (the default).
type Manager struct {
	cfg       *config.Config
	cmd       *exec.Cmd
	apiBase   string
	modelPath string
	modelName string
	dimension int
	isRunning bool
	stopChan  chan struct{}
}

// NewManager creates a new embedding service manager
//
// Note: With the new keyword-only default (embedding_provider = "none"),
// this manager is no longer started automatically. It only starts if you
// explicitly configure embedding_provider = "http" in your config.
func NewManager(cfg *config.Config) *Manager {
	apiBase := cfg.Memory.RAG.APIBase
	if apiBase == "" {
		apiBase = "http://localhost:18190"
	}

	// Default to sentence-transformers multilingual model
	modelName := "sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2"
	dimension := 384

	// Use model dimension from config if available
	if cfg.Memory.RAG.Dimension > 0 {
		dimension = cfg.Memory.RAG.Dimension
	}

	return &Manager{
		cfg:       cfg,
		apiBase:   apiBase,
		modelPath: cfg.Memory.RAG.ModelPath,
		modelName: modelName,
		dimension: dimension,
		stopChan:  make(chan struct{}),
	}
}

// IsConfigured returns true if embedding service should be auto-started
//
// With the new default (embedding_provider = "none"), this returns false
// unless you explicitly configure embedding_provider = "http"
func (m *Manager) IsConfigured() bool {
	// Only auto-start if explicitly using HTTP embedding model
	// Default is now "none" (keyword-only search)
	if !m.cfg.Memory.RAG.Enabled {
		return false
	}
	// Only start for explicit "http" configuration
	if m.cfg.Memory.RAG.EmbeddingModel != "http" {
		logger.Debug("Embedding service not configured - using keyword-only search (embedding_provider = \"none\")")
		return false
	}
	return true
}

// IsHealthy checks if the embedding service is responding
func (m *Manager) IsHealthy() bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(m.apiBase + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// Start starts the embedding service if not already running
//
// This method now logs a deprecation warning and only starts the service
// if explicitly configured with embedding_provider = "http"
func (m *Manager) Start() error {
	if !m.IsConfigured() {
		logger.Debug("Embedding service not started - using keyword-only search (FTS5)")
		return nil
	}

	logger.Warn("DEPRECATED: You are using embedding_provider = \"http\". " +
		"This is no longer required. Consider switching to \"none\" for zero-config operation.")

	// Check if already running
	if m.IsHealthy() {
		logger.Info("Embedding service already running at " + m.apiBase)
		m.isRunning = true
		return nil
	}

	// Find virtual environment in multiple locations
	venvDir := m.findVenvDir()
	if venvDir == "" {
		return fmt.Errorf("virtual environment not found. Run 'make setup-embedding' or install embedding service manually")
	}

	// Determine Python path
	var pythonPath string
	if runtime.GOOS == "windows" {
		pythonPath = filepath.Join(venvDir, "Scripts", "python.exe")
	} else {
		pythonPath = filepath.Join(venvDir, "bin", "python")
	}

	if _, err := os.Stat(pythonPath); os.IsNotExist(err) {
		return fmt.Errorf("Python not found at %s", pythonPath)
	}

	// Prepare environment variables
	env := os.Environ()

	// Set the embedding model name (for sentence-transformers)
	if m.modelName != "" {
		env = append(env, "EMBEDDING_MODEL="+m.modelName)
	}

	// Set dimension
	if m.dimension > 0 {
		env = append(env, fmt.Sprintf("EMBEDDING_DIMENSION=%d", m.dimension))
	}

	// Set models directory for caching
	modelsDir := m.modelPath
	if modelsDir == "" {
		// Default to ~/.picoclaw/models
		home, _ := os.UserHomeDir()
		modelsDir = filepath.Join(home, ".picoclaw", "models")
	}
	env = append(env, "MODELS_DIR="+modelsDir)

	// Start the service on port 18190
	m.cmd = exec.Command(pythonPath, "-m", "uvicorn", "main:app", "--host", "0.0.0.0", "--port", "18190")
	m.cmd.Dir = filepath.Dir(venvDir) // Use parent of .venv as working directory
	m.cmd.Env = env

	// Redirect output to logger
	m.cmd.Stdout = &logWriter{prefix: "[embedding] "}
	m.cmd.Stderr = &logWriter{prefix: "[embedding] "}

	logger.Info("Starting embedding service...")
	if err := m.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start embedding service: %w", err)
	}

	// Wait for service to be ready
	logger.Info("Waiting for embedding service to be ready...")
	if err := m.waitForReady(120 * time.Second); err != nil {
		m.cmd.Process.Kill()
		return fmt.Errorf("embedding service failed to start: %w", err)
	}

	m.isRunning = true
	logger.Info("Embedding service started successfully at " + m.apiBase)

	// Start monitoring goroutine
	go m.monitor()

	return nil
}

// Stop stops the embedding service
func (m *Manager) Stop() error {
	if !m.isRunning || m.cmd == nil {
		return nil
	}

	close(m.stopChan)

	// Try graceful shutdown first
	if m.cmd.Process != nil {
		logger.Info("Stopping embedding service...")
		m.cmd.Process.Signal(os.Interrupt)

		// Wait for process to exit
		done := make(chan error, 1)
		go func() {
			done <- m.cmd.Wait()
		}()

		select {
		case <-done:
			logger.Info("Embedding service stopped gracefully")
		case <-time.After(5 * time.Second):
			// Force kill if not stopped
			logger.Warn("Embedding service did not stop gracefully, forcing...")
			m.cmd.Process.Kill()
		}
	}

	m.isRunning = false
	return nil
}

// waitForReady waits for the service to become healthy
func (m *Manager) waitForReady(timeout time.Duration) error {
	start := time.Now()
	for time.Since(start) < timeout {
		if m.IsHealthy() {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for service to be ready")
}

// monitor monitors the service and restarts if needed
func (m *Manager) monitor() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopChan:
			return
		case <-ticker.C:
			if !m.IsHealthy() {
				logger.Warn("Embedding service health check failed, attempting restart...")
				if err := m.Start(); err != nil {
					logger.Error("Failed to restart embedding service: " + err.Error())
				}
			}
		}
	}
}

// findProjectRoot finds the embedding service root directory
// It checks PICOCLAW_HOME first, then falls back to ~/.picoclaw
// Returns the parent directory where services/embedding will be created
func (m *Manager) findProjectRoot() (string, error) {
	// Try PICOCLAW_HOME first
	if home := os.Getenv("PICOCLAW_HOME"); home != "" {
		if err := os.MkdirAll(home, 0755); err == nil {
			return home, nil
		}
	}

	// Try ~/.picoclaw
	userHome, err := os.UserHomeDir()
	if err == nil {
		picoclawDir := filepath.Join(userHome, ".picoclaw")
		if err := os.MkdirAll(picoclawDir, 0755); err == nil {
			return picoclawDir, nil
		}
	}

	// Fallback to current directory
	cwd, err := os.Getwd()
	if err == nil {
		return cwd, nil
	}

	return "", fmt.Errorf("cannot find suitable directory for embedding service")
}

// findVenvDir finds the virtual environment directory with valid Python executable
func (m *Manager) findVenvDir() string {
	possiblePaths := []string{
		// PICOCLAW_HOME
		filepath.Join(os.Getenv("PICOCLAW_HOME"), "services", "embedding", ".venv"),
		// ~/.picoclaw
		func() string {
			home, _ := os.UserHomeDir()
			return filepath.Join(home, ".picoclaw", "services", "embedding", ".venv")
		}(),
		// Project root (development)
		func() string {
			execPath, _ := os.Executable()
			dir := filepath.Dir(execPath)
			for {
				venv := filepath.Join(dir, "services", "embedding", ".venv")
				if _, err := os.Stat(venv); err == nil {
					return venv
				}
				parent := filepath.Dir(dir)
				if parent == dir {
					break
				}
				dir = parent
			}
			return ""
		}(),
	}

	for _, path := range possiblePaths {
		if path == "" {
			continue
		}
		// Check if venv exists and has Python executable
		var pythonPath string
		if runtime.GOOS == "windows" {
			pythonPath = filepath.Join(path, "Scripts", "python.exe")
		} else {
			pythonPath = filepath.Join(path, "bin", "python")
		}
		if _, err := os.Stat(pythonPath); err == nil {
			// Ensure we return absolute path
			absPath, _ := filepath.Abs(path)
			return absPath
		}
	}

	return ""
}

// logWriter implements io.Writer to log output
type logWriter struct {
	prefix string
}

func (w *logWriter) Write(p []byte) (n int, err error) {
	logger.Debug(w.prefix + string(p))
	return len(p), nil
}
