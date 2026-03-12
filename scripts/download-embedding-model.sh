#!/bin/bash
# Download embedding model for offline use
# This script pre-downloads the model so it doesn't need to be downloaded at runtime
#
# Environment Variables:
#   PICOCLAW_HOME - Base directory for picoclaw (default: ~/.picoclaw)

set -e

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
PICOCLAW_HOME="${PICOCLAW_HOME:-$USER_HOME/.picoclaw}"
VENV_DIR="$PICOCLAW_HOME/services/embedding/.venv"

# Cross-platform Python path
if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" || "$OSTYPE" == "win32" ]]; then
    PYTHON_CMD="$VENV_DIR/Scripts/python.exe"
else
    PYTHON_CMD="$VENV_DIR/bin/python"
fi

# Default model
MODEL_NAME="${1:-sentence-transformers/all-MiniLM-L6-v2}"

echo "=== Downloading Embedding Model ==="
echo "PICOCLAW_HOME: $PICOCLAW_HOME"
echo "Model: $MODEL_NAME"
echo ""

# Check if Python 3 is available
if ! command -v python3 &> /dev/null; then
    echo "Error: Python 3 is not installed"
    exit 1
fi

# Create virtual environment if it doesn't exist
if [ ! -d "$VENV_DIR" ]; then
    echo "Creating virtual environment at $VENV_DIR..."
    python3 -m venv "$VENV_DIR"
fi

# Activate virtual environment and install dependencies
echo "Installing dependencies..."
if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" || "$OSTYPE" == "win32" ]]; then
    "$VENV_DIR/Scripts/pip.exe" install -q --upgrade pip
    "$VENV_DIR/Scripts/pip.exe" install -q sentence-transformers
else
    "$VENV_DIR/bin/pip" install -q --upgrade pip
    "$VENV_DIR/bin/pip" install -q sentence-transformers
fi

echo ""
echo "Downloading model: $MODEL_NAME"
echo "This may take a few minutes depending on your internet connection..."
echo ""

# Download the model using Python
"$PYTHON_CMD" << EOF
from sentence_transformers import SentenceTransformer
import os

model_name = "$MODEL_NAME"
cache_dir = os.path.expanduser("~/.cache/torch/sentence_transformers")

print(f"Downloading {model_name}...")
print(f"Cache directory: {cache_dir}")

# Download the model
model = SentenceTransformer(model_name)

# Test the model
print("\nTesting model...")
test_embedding = model.encode("This is a test sentence.")
print(f"Model loaded successfully!")
print(f"Embedding dimension: {len(test_embedding)}")

print(f"\nModel cached at: {cache_dir}")
EOF

echo ""
echo "=== Download Complete ==="
echo "Model is ready to use offline"
echo ""
echo "You can now start the embedding service with:"
echo "  PICOCLAW_HOME=$PICOCLAW_HOME ./scripts/start-embedding-service.sh"
