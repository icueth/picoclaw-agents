package setup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

// getPicoclawHome returns the PICOCLAW_HOME directory
func getPicoclawHome() string {
	if home := os.Getenv("PICOCLAW_HOME"); home != "" {
		return home
	}
	userHome, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory
		cwd, _ := os.Getwd()
		return filepath.Join(cwd, ".picoclaw")
	}
	return filepath.Join(userHome, ".picoclaw")
}

// getProjectRoot tries to find the project root (for development mode)
func getProjectRoot() (string, error) {
	// Try to find project root by looking for go.mod
	execPath, err := os.Executable()
	if err != nil {
		// Fallback to current directory
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		return cwd, nil
	}

	// If running from build directory, go up to find project root
	dir := filepath.Dir(execPath)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// Fallback to current directory
	return os.Getwd()
}

// getServiceSourceDir returns the source directory for embedding service files
// DEPRECATED: Embedding service is no longer required
func getServiceSourceDir() string {
	projectRoot, err := getProjectRoot()
	if err != nil {
		return ""
	}

	serviceDir := filepath.Join(projectRoot, "services", "embedding")
	if _, err := os.Stat(serviceDir); err == nil {
		// Development mode - service files exist in project
		return serviceDir
	}

	return ""
}

func NewSetupCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Setup picoclaw components",
		Long: `Setup command for picoclaw.

DEPRECATED: Most setup is now automatic with zero-config keyword search.
The embedding service is no longer required for default operation.

PicoClaw now uses SQLite FTS5 for keyword search by default, requiring
no external services or configuration.`,
	}

	cmd.AddCommand(
		newSetupEmbeddingCommand(),
		newDownloadModelCommand(),
	)

	return cmd
}

func newSetupEmbeddingCommand() *cobra.Command {
	var skipModel bool
	var modelName string

	cmd := &cobra.Command{
		Use:   "embedding",
		Short: "Setup Python embedding service (DEPRECATED)",
		Long: `DEPRECATED: Setup Python embedding service for RAG functionality.

This command is DEPRECATED. PicoClaw now uses keyword-only search by default
(embedding_provider = "none"), which requires no external services.

You only need this if you explicitly want to use HTTP embeddings for hybrid search.

Examples:
  picoclaw setup embedding              # Full setup with default model (DEPRECATED)
  picoclaw setup embedding --skip-model # Setup without downloading model`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("⚠️  WARNING: Embedding service is DEPRECATED")
			fmt.Println()
			fmt.Println("PicoClaw now uses keyword-only search by default (embedding_provider = \"none\")")
			fmt.Println("No external services are required for memory/RAG functionality.")
			fmt.Println()
			fmt.Println("Only continue if you explicitly need HTTP embeddings for hybrid search.")
			fmt.Println()
			
			// Ask for confirmation
			fmt.Print("Continue with deprecated embedding setup? [y/N]: ")
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				fmt.Println("Cancelled. Using default keyword search.")
				return
			}
			
			runSetupEmbedding(skipModel, modelName)
		},
	}

	cmd.Flags().BoolVar(&skipModel, "skip-model", false, "Skip model download")
	cmd.Flags().StringVar(&modelName, "model", "embeddinggemma-300m-qat-Q8_0.gguf", "Model name to download")

	return cmd
}

func newDownloadModelCommand() *cobra.Command {
	var modelName string

	cmd := &cobra.Command{
		Use:   "download-model",
		Short: "Download embedding model only (DEPRECATED)",
		Long: `DEPRECATED: Download the specified embedding model.

This is only needed if using embedding_provider = "http".
Default operation (embedding_provider = "none") requires no models.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("⚠️  WARNING: Model download is DEPRECATED")
			fmt.Println()
			fmt.Println("Default keyword search (embedding_provider = \"none\") requires no models.")
			fmt.Println()
			
			// Ask for confirmation
			fmt.Print("Continue with deprecated model download? [y/N]: ")
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				fmt.Println("Cancelled.")
				return
			}
			
			runDownloadModel(modelName)
		},
	}

	cmd.Flags().StringVar(&modelName, "model", "embeddinggemma-300m-qat-Q8_0.gguf", "Model name to download")

	return cmd
}

func runSetupEmbedding(skipModel bool, modelName string) {
	fmt.Println("🔧 Setting up Python embedding service...")
	fmt.Println()

	picoclawHome := getPicoclawHome()
	fmt.Printf("📁 PICOCLAW_HOME: %s\n", picoclawHome)
	fmt.Println()

	// Check Python version
	fmt.Println("📋 Checking Python installation...")
	pythonCmd, err := checkPython()
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		fmt.Println()
		fmt.Println("Please install Python 3.9 or higher:")
		fmt.Println("  macOS:   brew install python")
		fmt.Println("  Ubuntu:  sudo apt install python3 python3-venv python3-pip")
		fmt.Println("  Windows: https://python.org/downloads")
		os.Exit(1)
	}
	fmt.Println("✅ Python is available")
	fmt.Println()

	// Setup directories
	embeddingDir := filepath.Join(picoclawHome, "services", "embedding")
	modelsDir := filepath.Join(picoclawHome, "models")
	venvDir := filepath.Join(embeddingDir, ".venv")

	// Create directories
	if err := os.MkdirAll(embeddingDir, 0755); err != nil {
		fmt.Printf("❌ Error creating embedding directory: %v\n", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(modelsDir, 0755); err != nil {
		fmt.Printf("❌ Error creating models directory: %v\n", err)
		os.Exit(1)
	}

	// Copy service files from source if available (development mode)
	sourceDir := getServiceSourceDir()
	if sourceDir != "" {
		fmt.Println("📂 Copying service files...")
		if err := copyServiceFiles(sourceDir, embeddingDir); err != nil {
			fmt.Printf("⚠️  Warning: Could not copy all service files: %v\n", err)
		} else {
			fmt.Println("✅ Service files copied")
		}
		fmt.Println()
	}

	// Check if main.py exists
	mainPyPath := filepath.Join(embeddingDir, "main.py")
	if _, err := os.Stat(mainPyPath); err != nil {
		fmt.Printf("❌ Embedding service files not found at %s\n", embeddingDir)
		fmt.Println()
		fmt.Println("Please ensure you're running from the project directory")
		fmt.Println("or install service files manually to:")
		fmt.Printf("  %s\n", embeddingDir)
		os.Exit(1)
	}

	// Setup virtual environment
	fmt.Println("📦 Setting up virtual environment...")
	if err := setupVirtualEnv(pythonCmd, venvDir); err != nil {
		fmt.Printf("❌ Error setting up virtual environment: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Virtual environment ready")
	fmt.Println()

	// Install dependencies
	fmt.Println("📥 Installing dependencies...")
	if err := installDependencies(venvDir, embeddingDir); err != nil {
		fmt.Printf("❌ Error installing dependencies: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Dependencies installed")
	fmt.Println()

	// Download model if not skipped
	if !skipModel {
		fmt.Printf("📥 Downloading model: %s\n", modelName)
		if err := downloadModel(venvDir, modelsDir, modelName); err != nil {
			fmt.Printf("⚠️  Warning: Could not download model: %v\n", err)
			fmt.Println("   You can download it later with: picoclaw setup download-model")
		} else {
			fmt.Println("✅ Model downloaded")
		}
		fmt.Println()
	}

	fmt.Println("🎉 Setup complete!")
	fmt.Println()
	fmt.Println("⚠️  Note: Embedding service is DEPRECATED")
	fmt.Println("Consider using keyword-only search (embedding_provider = \"none\") for zero-config operation.")
	fmt.Println()
	fmt.Println("Installation locations:")
	fmt.Printf("  Service: %s\n", embeddingDir)
	fmt.Printf("  Models:  %s\n", modelsDir)
	fmt.Printf("  Venv:    %s\n", venvDir)
	fmt.Println()
	fmt.Println("To use HTTP embeddings, configure in config.json:")
	fmt.Println(`  "embedding_provider": "http"`)
	fmt.Println()
}

func runDownloadModel(modelName string) {
	fmt.Printf("📥 Downloading model: %s\n", modelName)
	fmt.Println()

	picoclawHome := getPicoclawHome()
	modelsDir := filepath.Join(picoclawHome, "models")
	venvDir := filepath.Join(picoclawHome, "services", "embedding", ".venv")

	if err := os.MkdirAll(modelsDir, 0755); err != nil {
		fmt.Printf("❌ Error creating models directory: %v\n", err)
		os.Exit(1)
	}

	if err := downloadModel(venvDir, modelsDir, modelName); err != nil {
		fmt.Printf("❌ Error downloading model: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("✅ Model downloaded successfully!")
	fmt.Printf("📁 Location: %s/%s\n", modelsDir, modelName)
	fmt.Println()
	fmt.Println("⚠️  Note: This model is only needed for HTTP embedding mode.")
	fmt.Println("Default keyword search (embedding_provider = \"none\") requires no models.")
}

func checkPython() (string, error) {
	// Try python3 first
	if _, err := exec.LookPath("python3"); err == nil {
		cmd := exec.Command("python3", "--version")
		output, err := cmd.Output()
		if err == nil {
			fmt.Printf("   Found: %s", string(output))
			return "python3", nil
		}
	}

	// Try python
	if _, err := exec.LookPath("python"); err == nil {
		cmd := exec.Command("python", "--version")
		output, err := cmd.Output()
		if err == nil {
			fmt.Printf("   Found: %s", string(output))
			return "python", nil
		}
	}

	return "", fmt.Errorf("Python not found")
}

func copyServiceFiles(sourceDir, targetDir string) error {
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		sourcePath := filepath.Join(sourceDir, entry.Name())
		targetPath := filepath.Join(targetDir, entry.Name())

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.IsDir() {
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return err
			}
			if err := copyServiceFiles(sourcePath, targetPath); err != nil {
				return err
			}
		} else {
			data, err := os.ReadFile(sourcePath)
			if err != nil {
				return err
			}
			if err := os.WriteFile(targetPath, data, 0644); err != nil {
				return err
			}
		}
	}

	return nil
}

func setupVirtualEnv(pythonCmd, venvPath string) error {
	// Check if venv already exists
	if _, err := os.Stat(venvPath); err == nil {
		fmt.Println("   Virtual environment already exists, skipping creation")
		return nil
	}

	cmd := exec.Command(pythonCmd, "-m", "venv", venvPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create virtual environment: %v\n%s", err, output)
	}

	return nil
}

func installDependencies(venvDir, embeddingDir string) error {
	var pipPath string
	if runtime.GOOS == "windows" {
		pipPath = filepath.Join(venvDir, "Scripts", "pip.exe")
	} else {
		pipPath = filepath.Join(venvDir, "bin", "pip")
	}

	requirementsPath := filepath.Join(embeddingDir, "requirements.txt")

	// Check if requirements.txt exists
	if _, err := os.Stat(requirementsPath); err != nil {
		fmt.Printf("   Warning: requirements.txt not found at %s\n", requirementsPath)
		return nil
	}

	// Upgrade pip first
	cmd := exec.Command(pipPath, "install", "--upgrade", "pip")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to upgrade pip: %v\n%s", err, output)
	}

	// Install requirements
	cmd = exec.Command(pipPath, "install", "-r", requirementsPath)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install dependencies: %v\n%s", err, output)
	}

	// Install huggingface_hub
	cmd = exec.Command(pipPath, "install", "huggingface_hub")
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install huggingface_hub: %v\n%s", err, output)
	}

	return nil
}

func downloadModel(venvDir, modelsDir, modelName string) error {
	var pythonPath string
	if runtime.GOOS == "windows" {
		pythonPath = filepath.Join(venvDir, "Scripts", "python.exe")
	} else {
		pythonPath = filepath.Join(venvDir, "bin", "python")
	}

	modelPath := filepath.Join(modelsDir, modelName)

	// Check if model already exists
	if _, err := os.Stat(modelPath); err == nil {
		fmt.Printf("   Model already exists: %s\n", modelPath)
		return nil
	}

	// Check if huggingface_hub is installed
	cmd := exec.Command(pythonPath, "-c", "import huggingface_hub")
	if err := cmd.Run(); err != nil {
		// Install huggingface_hub
		fmt.Println("   Installing huggingface_hub...")
		var pipPath string
		if runtime.GOOS == "windows" {
			pipPath = filepath.Join(venvDir, "Scripts", "pip.exe")
		} else {
			pipPath = filepath.Join(venvDir, "bin", "pip")
		}
		cmd = exec.Command(pipPath, "install", "huggingface_hub")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to install huggingface_hub: %v\n%s", err, output)
		}
	}

	// Download model using Python script
	directScript := fmt.Sprintf(`
from huggingface_hub import hf_hub_download
import os

model_name = '%s'
models_dir = '%s'

# Parse repo_id and filename
# Format: username/repo_id/filename or repo_id/filename
parts = model_name.split('/')
if len(parts) >= 2:
    repo_id = '/'.join(parts[:-1])
    filename = parts[-1]
else:
    # Default repo for embedding models
    repo_id = "ChristianAzinn/embeddinggemma-300m-qat"
    filename = model_name

print(f"Downloading {filename} from {repo_id}...")
local_path = hf_hub_download(
    repo_id=repo_id,
    filename=filename,
    local_dir=models_dir,
    local_dir_use_symlinks=False
)
print(f"Downloaded to: {local_path}")
`, modelName, modelsDir)

	cmd = exec.Command(pythonPath, "-c", directScript)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to download model: %v", err)
	}

	return nil
}
