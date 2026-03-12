#!/bin/bash
#
# Start the Python Embedding Service for PicoClaw
#
# This script starts the FastAPI-based embedding service that provides
# high-quality text embeddings for the RAG system.
#
# Usage:
#   ./start-embedding-service.sh [port]
#
# Environment Variables:
#   EMBEDDING_PORT - Port to run the service on (default: 18190)
#   EMBEDDING_MODEL - Model to use (default: sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2)
#   EMBEDDING_DIMENSION - Embedding dimension (default: 384)
#   MODELS_DIR - Directory to store/load models (default: ~/.picoclaw/models)
#   PICOCLAW_HOME - Base directory for picoclaw (default: ~/.picoclaw)
#   PYTHON - Python executable to use (default: auto-detect from venv)
#

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

# Cross-platform path setup
if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" || "$OSTYPE" == "win32" ]]; then
    # Windows paths
    EMBEDDING_DIR="$PICOCLAW_HOME/services/embedding"
    VENV_DIR="$EMBEDDING_DIR/.venv"
    VENV_PYTHON="$VENV_DIR/Scripts/python.exe"
    VENV_PIP="$VENV_DIR/Scripts/pip.exe"
else
    # Unix paths
    EMBEDDING_DIR="$PICOCLAW_HOME/services/embedding"
    VENV_DIR="$EMBEDDING_DIR/.venv"
    VENV_PYTHON="$VENV_DIR/bin/python"
    VENV_PIP="$VENV_DIR/bin/pip"
fi

MODELS_DIR="${MODELS_DIR:-$PICOCLAW_HOME/models}"

# Configuration
PORT="${1:-${EMBEDDING_PORT:-18190}}"
HOST="${EMBEDDING_HOST:-0.0.0.0}"
MODEL="${EMBEDDING_MODEL:-sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2}"
DIMENSION="${EMBEDDING_DIMENSION:-384}"
WORKERS="${EMBEDDING_WORKERS:-1}"

# Colors for output (disable on Windows if not supported)
if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" || "$OSTYPE" == "win32" ]]; then
    RED=''
    GREEN=''
    YELLOW=''
    NC=''
else
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    NC='\033[0m'
fi

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if embedding service is installed
if [ ! -d "$EMBEDDING_DIR" ]; then
    log_error "Embedding service directory not found: $EMBEDDING_DIR"
    log_info "Please run: make setup-embedding-user"
    log_info "Or: PICOCLAW_HOME=$PICOCLAW_HOME ./scripts/setup-embedding-user.sh"
    exit 1
fi

if [ ! -d "$VENV_DIR" ]; then
    log_error "Virtual environment not found: $VENV_DIR"
    log_info "Please run: make setup-embedding-user"
    exit 1
fi

cd "$EMBEDDING_DIR"

# Check if Python is available in venv
if [ ! -f "$VENV_PYTHON" ]; then
    log_error "Python not found in virtual environment: $VENV_PYTHON"
    log_info "Please reinstall the embedding service: make setup-embedding-user"
    exit 1
fi

log_info "Using Python: $VENV_PYTHON"

# Check Python version
PYTHON_VERSION=$($VENV_PYTHON --version 2>&1 | awk '{print $2}')
log_info "Python version: $PYTHON_VERSION"

# Install/update dependencies
if [ -f "requirements.txt" ]; then
    log_info "Installing/updating dependencies..."
    "$VENV_PYTHON" -m pip install -q --upgrade pip
    "$VENV_PYTHON" -m pip install -q -r requirements.txt
fi

# Export environment variables
export EMBEDDING_MODEL="$MODEL"
export EMBEDDING_DIMENSION="$DIMENSION"
export MODELS_DIR="$MODELS_DIR"

log_info "Starting Embedding Service"
log_info "  Host: $HOST"
log_info "  Port: $PORT"
log_info "  Model: $MODEL"
log_info "  Dimension: $DIMENSION"
log_info "  Workers: $WORKERS"
log_info "  Models Dir: $MODELS_DIR"
log_info "  Service Dir: $EMBEDDING_DIR"

# Check if port is already in use
if command -v lsof &> /dev/null; then
    if lsof -Pi :"$PORT" -sTCP:LISTEN -t >/dev/null 2>&1; then
        log_warn "Port $PORT is already in use"
        log_info "The embedding service may already be running"
        log_info "Check with: curl http://localhost:$PORT/health"
        exit 0
    fi
elif command -v netstat &> /dev/null; then
    if netstat -an | grep -q ":$PORT "; then
        log_warn "Port $PORT may already be in use"
        log_info "Check with: curl http://localhost:$PORT/health"
    fi
fi

# Start the service
log_info "Starting uvicorn server..."
log_info "Health check: curl http://localhost:$PORT/health"
log_info "Press Ctrl+C to stop"

trap 'log_info "Shutting down embedding service..."; exit 0' INT TERM

"$VENV_PYTHON" -m uvicorn main:app \
    --host "$HOST" \
    --port "$PORT" \
    --workers "$WORKERS" \
    --log-level info \
    --access-log \
    --reload
