#!/bin/bash
# Setup and run embedding service for PicoClaw
# This script installs dependencies and starts the embedding service

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVICE_DIR="$SCRIPT_DIR/../services/embedding"
VENV_DIR="$SERVICE_DIR/.venv"

echo "=== PicoClaw Embedding Service Setup ==="
echo ""

# Check if Python 3 is available
if ! command -v python3 &> /dev/null; then
    echo "Error: Python 3 is not installed"
    exit 1
fi

# Create virtual environment if it doesn't exist
if [ ! -d "$VENV_DIR" ]; then
    echo "Creating virtual environment..."
    python3 -m venv "$VENV_DIR"
fi

# Activate virtual environment
echo "Activating virtual environment..."
source "$VENV_DIR/bin/activate"

# Install dependencies
echo "Installing dependencies..."
pip install -q --upgrade pip

# Install core dependencies first (fast)
echo "Installing core dependencies..."
pip install -q fastapi uvicorn pydantic numpy sentence-transformers

# Optional: Install llama-cpp-python for GGUF support
# This is slow to compile, so we skip it by default
# Uncomment if you need GGUF support:
# echo "Installing llama-cpp-python (this may take a while)..."
# pip install -q llama-cpp-python

echo ""
echo "=== Starting Embedding Service ==="
echo "Service will be available at: http://localhost:8000"
echo "API Documentation: http://localhost:8000/docs"
echo ""
echo "Default model: sentence-transformers/all-MiniLM-L6-v2 (384 dimensions)"
echo ""
echo "To use a different model, set the EMBEDDING_MODEL environment variable:"
echo "  EMBEDDING_MODEL=sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2 ./setup-embedding-service.sh"
echo ""

# Run the service
cd "$SERVICE_DIR"
python -m uvicorn main:app --host 0.0.0.0 --port 8000 --reload
