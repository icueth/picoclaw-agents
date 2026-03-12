#!/bin/bash
#
# Setup script for Python embedding service in user home (~/.picoclaw)
# This script is called by 'make install' to set up embedding service for end users
# Supports: macOS, Linux, Windows (Git Bash, WSL, MSYS2)
#

set -e

# Colors for output (disable on Windows if not supported)
if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" || "$OSTYPE" == "win32" ]]; then
    RED=''
    GREEN=''
    YELLOW=''
    BLUE=''
    NC=''
else
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    BLUE='\033[0;34m'
    NC='\033[0m' # No Color
fi

# Default configuration
# Using sentence-transformers model (downloads automatically on first use)
DEFAULT_MODEL="sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2"
MODEL="${EMBEDDING_MODEL:-$DEFAULT_MODEL}"
# Download model during install for faster first startup
SKIP_MODEL="${SKIP_MODEL:-false}"

# Cross-platform home directory detection
detect_home_dir() {
    if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" || "$OSTYPE" == "win32" ]]; then
        # Windows
        if [[ -n "$USERPROFILE" ]]; then
            echo "$USERPROFILE"
        elif [[ -n "$HOME" ]]; then
            echo "$HOME"
        else
            echo "C:/Users/$(whoami)"
        fi
    else
        # macOS, Linux, Unix
        if [[ -n "$HOME" ]]; then
            echo "$HOME"
        else
            echo "$(eval echo ~$(whoami))"
        fi
    fi
}

USER_HOME="$(detect_home_dir)"

# Use PICOCLAW_HOME or default to ~/.picoclaw
PICOCLAW_HOME="${PICOCLAW_HOME:-$USER_HOME/.picoclaw}"

# Get script directory and project root (for source files)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Cross-platform path setup
if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" || "$OSTYPE" == "win32" ]]; then
    # Windows paths
    VENV_DIR="$PICOCLAW_HOME/services/embedding/.venv"
    VENV_BIN_DIR="$VENV_DIR/Scripts"
    VENV_PYTHON="$VENV_BIN_DIR/python.exe"
    VENV_PIP="$VENV_BIN_DIR/pip.exe"
else
    # Unix paths
    VENV_DIR="$PICOCLAW_HOME/services/embedding/.venv"
    VENV_BIN_DIR="$VENV_DIR/bin"
    VENV_PYTHON="$VENV_BIN_DIR/python"
    VENV_PIP="$VENV_BIN_DIR/pip"
fi

MODELS_DIR="$PICOCLAW_HOME/models"
SERVICE_DIR="$PICOCLAW_HOME/services/embedding"
REQUIREMENTS="$PROJECT_ROOT/services/embedding/requirements.txt"

echo -e "${BLUE}🔧 PicoClaw Embedding Service Setup (User Install)${NC}"
echo "===================================="
echo ""
echo -e "${BLUE}Detected OS:${NC} $OSTYPE"
echo -e "${BLUE}User Home:${NC} $USER_HOME"
echo -e "${BLUE}Install Location:${NC} $PICOCLAW_HOME"
echo ""

# Function to print status
print_status() {
    echo -e "${BLUE}➜${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# Check Python installation
check_python() {
    print_status "Checking Python installation..."

    if command -v python3 &> /dev/null; then
        PYTHON_CMD="python3"
    elif command -v python &> /dev/null; then
        PYTHON_CMD="python"
    else
        print_error "Python is not installed"
        echo ""
        echo "Please install Python 3.9 or higher:"
        echo "  macOS:   brew install python"
        echo "  Ubuntu:  sudo apt install python3 python3-venv python3-pip"
        echo "  Windows: https://python.org/downloads"
        exit 1
    fi

    # Check Python version
    PYTHON_VERSION=$($PYTHON_CMD --version 2>&1 | awk '{print $2}')
    print_success "Found Python $PYTHON_VERSION"

    # Check if version is 3.9+
    REQUIRED_VERSION="3.9"
    if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$PYTHON_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
        print_error "Python 3.9 or higher is required (found $PYTHON_VERSION)"
        exit 1
    fi
}

# Copy service files to user directory
setup_service_files() {
    print_status "Setting up service files in $SERVICE_DIR..."
    
    # Create directories
    mkdir -p "$SERVICE_DIR"
    mkdir -p "$MODELS_DIR"
    
    # Copy embedding service files
    if [ -d "$PROJECT_ROOT/services/embedding" ]; then
        # Copy main.py and any other Python files
        cp "$PROJECT_ROOT/services/embedding/"*.py "$SERVICE_DIR/" 2>/dev/null || true
        # Copy requirements.txt
        cp "$PROJECT_ROOT/services/embedding/requirements.txt" "$SERVICE_DIR/" 2>/dev/null || true
        print_success "Copied service files to $SERVICE_DIR"
    else
        print_error "Service files not found at $PROJECT_ROOT/services/embedding"
        exit 1
    fi
}

# Setup virtual environment
setup_venv() {
    print_status "Setting up virtual environment at $VENV_DIR..."

    if [ -d "$VENV_DIR" ]; then
        print_warning "Virtual environment already exists"
        if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" || "$OSTYPE" == "win32" ]]; then
            read -p "Recreate? (y/N): " -n 1 -r
        else
            read -p "Recreate? (y/N): " -n 1 -r
        fi
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            rm -rf "$VENV_DIR"
        else
            print_success "Using existing virtual environment"
            return
        fi
    fi

    $PYTHON_CMD -m venv "$VENV_DIR"
    print_success "Created virtual environment at $VENV_DIR"
}

# Install dependencies
install_deps() {
    print_status "Installing dependencies..."

    # Upgrade pip
    if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" || "$OSTYPE" == "win32" ]]; then
        "$VENV_PYTHON" -m pip install --upgrade pip -q
    else
        "$VENV_PIP" install --upgrade pip -q
    fi

    # Install requirements
    SERVICE_REQUIREMENTS="$SERVICE_DIR/requirements.txt"
    if [ -f "$SERVICE_REQUIREMENTS" ]; then
        if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" || "$OSTYPE" == "win32" ]]; then
            "$VENV_PYTHON" -m pip install -r "$SERVICE_REQUIREMENTS" -q
        else
            "$VENV_PIP" install -r "$SERVICE_REQUIREMENTS" -q
        fi
        print_success "Installed dependencies from requirements.txt"
    else
        print_error "requirements.txt not found at $SERVICE_REQUIREMENTS"
        exit 1
    fi

    # Install huggingface_hub for model download
    if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" || "$OSTYPE" == "win32" ]]; then
        "$VENV_PYTHON" -m pip install huggingface_hub -q
    else
        "$VENV_PIP" install huggingface_hub -q
    fi
    print_success "Installed huggingface_hub"
}

# Download model
download_model() {
    if [ "$SKIP_MODEL" = "true" ]; then
        print_warning "Skipping model download (--skip-model flag set)"
        return
    fi

    print_status "Downloading embedding model: $MODEL"

    # Create models directory
    mkdir -p "$MODELS_DIR"

    # Download using Python script
    print_status "Downloading from Hugging Face Hub..."

    DOWNLOAD_SCRIPT="
from huggingface_hub import hf_hub_download
import sys
import os

model_name = '$MODEL'
models_dir = '$MODELS_DIR'

# Parse repo_id and filename
if '/' in model_name:
    parts = model_name.rsplit('/', 1)
    if len(parts) == 2 and '.' in parts[1]:
        repo_id = parts[0]
        filename = parts[1]
    else:
        # Default repo for embeddinggemma
        repo_id = 'ChristianAzinn/embeddinggemma-300m-qat'
        filename = model_name
else:
    repo_id = 'ChristianAzinn/embeddinggemma-300m-qat'
    filename = model_name

print(f'Downloading {filename} from {repo_id}...')
try:
    local_path = hf_hub_download(
        repo_id=repo_id,
        filename=filename,
        local_dir=models_dir,
        local_dir_use_symlinks=False
    )
    print(f'Successfully downloaded to: {local_path}')
except Exception as e:
    print(f'Error: {e}', file=sys.stderr)
    sys.exit(1)
"

    if "$VENV_PYTHON" -c "$DOWNLOAD_SCRIPT"; then
        print_success "Model downloaded to $MODELS_DIR/$MODEL"
    else
        print_warning "Failed to download model (will retry on first use)"
        echo ""
        echo "You can download it manually from:"
        echo "  https://huggingface.co/ChristianAzinn/embeddinggemma-300m-qat"
        echo ""
        echo "Or skip this step with: SKIP_MODEL=true $0"
    fi
}

# Create wrapper scripts
create_wrappers() {
    print_status "Creating wrapper scripts..."
    
    WRAPPER_DIR="$PICOCLAW_HOME/bin"
    mkdir -p "$WRAPPER_DIR"
    
    # Create start-embedding wrapper
    if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" || "$OSTYPE" == "win32" ]]; then
        # Windows batch wrapper
        cat > "$WRAPPER_DIR/start-embedding.bat" << EOF
@echo off
set EMBEDDING_MODEL=$MODEL
set MODELS_DIR=$MODELS_DIR
set EMBEDDING_PORT=18190
set EMBEDDING_HOST=0.0.0.0
"$VENV_PYTHON" "$SERVICE_DIR/main.py"
EOF
    else
        # Unix shell wrapper
        cat > "$WRAPPER_DIR/start-embedding" << EOF
#!/bin/bash
# Auto-generated embedding service wrapper
export EMBEDDING_MODEL="${EMBEDDING_MODEL:-$DEFAULT_MODEL}"
export MODELS_DIR="$MODELS_DIR"
export EMBEDDING_PORT="\${EMBEDDING_PORT:-18190}"
export EMBEDDING_HOST="\${EMBEDDING_HOST:-0.0.0.0}"
cd "$SERVICE_DIR"
exec "$VENV_PYTHON" -m uvicorn main:app --host "\$EMBEDDING_HOST" --port "\$EMBEDDING_PORT" --workers 1
EOF
        chmod +x "$WRAPPER_DIR/start-embedding"
    fi
    
    print_success "Created wrapper scripts in $WRAPPER_DIR"
}

# Print usage
print_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --skip-model          Skip model download"
    echo "  --model MODEL_NAME    Specify model name (default: $DEFAULT_MODEL)"
    echo "  -h, --help            Show this help message"
    echo ""
    echo "Environment Variables:"
    echo "  PICOCLAW_HOME         Installation directory (default: ~/.picoclaw)"
    echo "  EMBEDDING_MODEL       Model name to download"
    echo "  SKIP_MODEL            Set to 'true' to skip model download"
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --skip-model)
            SKIP_MODEL="true"
            shift
            ;;
        --model)
            MODEL="$2"
            shift 2
            ;;
        -h|--help)
            print_usage
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            print_usage
            exit 1
            ;;
    esac
done

# Main setup流程
main() {
    check_python
    setup_service_files
    setup_venv
    install_deps
    download_model
    create_wrappers

    echo ""
    echo -e "${GREEN}====================================${NC}"
    echo -e "${GREEN}🎉 Setup Complete!${NC}"
    echo -e "${GREEN}====================================${NC}"
    echo ""
    echo "Installation locations:"
    echo "  Virtual Env: $VENV_DIR"
    echo "  Models:      $MODELS_DIR"
    echo "  Service:     $SERVICE_DIR"
    echo "  Wrappers:    $PICOCLAW_HOME/bin/"
    echo ""
    echo "Environment variables set:"
    echo "  export PICOCLAW_HOME='$PICOCLAW_HOME'"
    echo ""
    echo "To start the embedding service manually:"
    echo "  $PICOCLAW_HOME/bin/start-embedding"
    echo ""
    echo "The embedding service will auto-start when you run:"
    echo -e "  ${YELLOW}picoclaw gateway${NC}"
    echo ""
}

# Run main
main
