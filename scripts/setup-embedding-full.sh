#!/bin/bash
#
# Setup script for Python embedding service
# This script sets up the Python environment and downloads the embedding model
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default configuration
# Using sentence-transformers model (downloads automatically on first use)
DEFAULT_MODEL="sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2"
MODEL="${EMBEDDING_MODEL:-$DEFAULT_MODEL}"
# Download model during install for faster first startup
SKIP_MODEL="${SKIP_MODEL:-false}"

# Get script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Paths
VENV_DIR="$PROJECT_ROOT/services/embedding/.venv"
MODELS_DIR="$PROJECT_ROOT/models"
REQUIREMENTS="$PROJECT_ROOT/services/embedding/requirements.txt"

echo -e "${BLUE}🔧 PicoClaw Embedding Service Setup${NC}"
echo "===================================="
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

# Setup virtual environment
setup_venv() {
    print_status "Setting up virtual environment..."

    if [ -d "$VENV_DIR" ]; then
        print_warning "Virtual environment already exists at $VENV_DIR"
        read -p "Recreate? (y/N): " -n 1 -r
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

    if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" || "$OSTYPE" == "win32" ]]; then
        PIP_CMD="$VENV_DIR/Scripts/pip.exe"
    else
        PIP_CMD="$VENV_DIR/bin/pip"
    fi

    # Upgrade pip
    $PIP_CMD install --upgrade pip -q

    # Install requirements
    if [ -f "$REQUIREMENTS" ]; then
        $PIP_CMD install -r "$REQUIREMENTS" -q
        print_success "Installed dependencies from requirements.txt"
    else
        print_error "requirements.txt not found at $REQUIREMENTS"
        exit 1
    fi

    # Install huggingface_hub for model download
    $PIP_CMD install huggingface_hub -q
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

    # Check if model already exists
    if [ -f "$MODELS_DIR/$MODEL" ]; then
        print_success "Model already exists: $MODELS_DIR/$MODEL"
        return
    fi

    # Determine Python path
    if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" || "$OSTYPE" == "win32" ]]; then
        PYTHON_VENV="$VENV_DIR/Scripts/python.exe"
    else
        PYTHON_VENV="$VENV_DIR/bin/python"
    fi

    # Download using Python script
    print_status "Downloading from Hugging Face Hub..."

    DOWNLOAD_SCRIPT="
from huggingface_hub import hf_hub_download
import sys

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

    if $PYTHON_VENV -c "$DOWNLOAD_SCRIPT"; then
        print_success "Model downloaded to $MODELS_DIR/$MODEL"
    else
        print_error "Failed to download model"
        echo ""
        echo "You can download it manually from:"
        echo "  https://huggingface.co/ChristianAzinn/embeddinggemma-300m-qat"
        echo ""
        echo "Or skip this step with: SKIP_MODEL=true $0"
        exit 1
    fi
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
    echo -e "${BLUE}Project Root:${NC} $PROJECT_ROOT"
    echo -e "${BLUE}Virtual Env:${NC}  $VENV_DIR"
    echo -e "${BLUE}Models Dir:${NC}   $MODELS_DIR"
    echo -e "${BLUE}Model:${NC}        $MODEL"
    echo ""

    check_python
    setup_venv
    install_deps
    download_model

    echo ""
    echo -e "${GREEN}====================================${NC}"
    echo -e "${GREEN}🎉 Setup Complete!${NC}"
    echo -e "${GREEN}====================================${NC}"
    echo ""
    echo "To start the embedding service:"
    echo ""
    echo "  ${YELLOW}make start-embedding${NC}"
    echo ""
    echo "Or manually:"
    echo "  ${YELLOW}cd $PROJECT_ROOT/services/embedding${NC}"
    echo "  ${YELLOW}source .venv/bin/activate${NC}"
    echo "  ${YELLOW}uvicorn main:app --host 0.0.0.0 --port 8000${NC}"
    echo ""
    echo "With GGUF model:"
    echo "  ${YELLOW}EMBEDDING_MODEL=$MODEL make start-embedding${NC}"
    echo ""
}

# Run main
main
