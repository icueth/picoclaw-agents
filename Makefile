.PHONY: all build install uninstall clean help test

# Build variables
BINARY_NAME=picoclaw
BUILD_DIR=build
CMD_DIR=cmd/$(BINARY_NAME)
MAIN_GO=$(CMD_DIR)/main.go

# Version
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT=$(shell git rev-parse --short=8 HEAD 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date +%FT%T%z)
GO_VERSION=$(shell $(GO) version | awk '{print $$3}')
INTERNAL=picoclaw/agent/cmd/picoclaw/internal
LDFLAGS=-ldflags "-X $(INTERNAL).version=$(VERSION) -X $(INTERNAL).gitCommit=$(GIT_COMMIT) -X $(INTERNAL).buildTime=$(BUILD_TIME) -X $(INTERNAL).goVersion=$(GO_VERSION) -s -w"

# Go variables
GO?=CGO_ENABLED=0 go
GOFLAGS?=-v -tags stdjson

# Golangci-lint
GOLANGCI_LINT?=golangci-lint

# Installation
# Auto-detect optimal install prefix based on OS and existing PATH entries
ifeq ($(PLATFORM),darwin)
    # macOS: prefer /usr/local for system-wide or ~/.local for user install
    ifeq ($(shell test -w /usr/local && echo yes),yes)
        DEFAULT_INSTALL_PREFIX=/usr/local
    else
        DEFAULT_INSTALL_PREFIX=$(HOME)/.local
    endif
else ifeq ($(PLATFORM),linux)
    # Linux: prefer ~/.local for user install, /usr/local if root
    ifeq ($(shell test -w /usr/local && echo yes),yes)
        DEFAULT_INSTALL_PREFIX=/usr/local
    else
        DEFAULT_INSTALL_PREFIX=$(HOME)/.local
    endif
else ifeq ($(OS),Windows_NT)
    # Windows: use LOCALAPPDATA or USERPROFILE
    DEFAULT_INSTALL_PREFIX=$(LOCALAPPDATA)/Programs
else
    DEFAULT_INSTALL_PREFIX=$(HOME)/.local
endif

INSTALL_PREFIX?=$(DEFAULT_INSTALL_PREFIX)
INSTALL_BIN_DIR=$(INSTALL_PREFIX)/bin
INSTALL_MAN_DIR=$(INSTALL_PREFIX)/share/man/man1
INSTALL_TMP_SUFFIX=.new

# Workspace and Skills - Cross-platform home detection
ifeq ($(OS),Windows_NT)
    USER_HOME?=$(USERPROFILE)
else
    USER_HOME?=$(HOME)
endif

PICOCLAW_HOME?=$(USER_HOME)/.picoclaw
WORKSPACE_DIR?=$(PICOCLAW_HOME)/workspace
WORKSPACE_SKILLS_DIR=$(WORKSPACE_DIR)/skills
BUILTIN_SKILLS_DIR=$(CURDIR)/skills

# OS detection
UNAME_S:=$(shell uname -s)
UNAME_M:=$(shell uname -m)

# Platform-specific settings
ifeq ($(UNAME_S),Linux)
	PLATFORM=linux
	ifeq ($(UNAME_M),x86_64)
		ARCH=amd64
	else ifeq ($(UNAME_M),aarch64)
		ARCH=arm64
	else ifeq ($(UNAME_M),armv81)
		ARCH=arm64
	else ifeq ($(UNAME_M),loongarch64)
		ARCH=loong64
	else ifeq ($(UNAME_M),riscv64)
		ARCH=riscv64
	else
		ARCH=$(UNAME_M)
	endif
else ifeq ($(UNAME_S),Darwin)
	PLATFORM=darwin
	ifeq ($(UNAME_M),x86_64)
		ARCH=amd64
	else ifeq ($(UNAME_M),arm64)
		ARCH=arm64
	else
		ARCH=$(UNAME_M)
	endif
else
	PLATFORM=$(UNAME_S)
	ARCH=$(UNAME_M)
endif

BINARY_PATH=$(BUILD_DIR)/$(BINARY_NAME)-$(PLATFORM)-$(ARCH)

# Default target
all: build

## generate: Run generate
generate:
	@echo "Run generate..."
	@rm -r ./$(CMD_DIR)/workspace 2>/dev/null || true
	@$(GO) generate ./...
	@echo "Run generate complete"

## build: Build the picoclaw binary for current platform
build: generate
	@echo "Building $(BINARY_NAME) for $(PLATFORM)/$(ARCH)..."
	@mkdir -p $(BUILD_DIR)
	@$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BINARY_PATH) ./$(CMD_DIR)
	@echo "Build complete: $(BINARY_PATH)"
	@ln -sf $(BINARY_NAME)-$(PLATFORM)-$(ARCH) $(BUILD_DIR)/$(BINARY_NAME)

## dev: Quick build and update binary in ~/.local/bin (user preference)
dev: build
	@echo "Updating binary in $(HOME)/.local/bin/$(BINARY_NAME)..."
	@mkdir -p $(HOME)/.local/bin
	@rm -f $(HOME)/.local/bin/$(BINARY_NAME)
	@cp -f $(BUILD_DIR)/$(BINARY_NAME) $(HOME)/.local/bin/$(BINARY_NAME)
	@chmod +x $(HOME)/.local/bin/$(BINARY_NAME)
	@echo "✅ Binary updated: $(HOME)/.local/bin/$(BINARY_NAME)"

## build-whatsapp-native: Build with WhatsApp native (whatsmeow) support; larger binary
build-whatsapp-native: generate
## @echo "Building $(BINARY_NAME) with WhatsApp native for $(PLATFORM)/$(ARCH)..."
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build -tags whatsapp_native $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./$(CMD_DIR)
	GOOS=linux GOARCH=arm GOARM=7 $(GO) build -tags whatsapp_native $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm ./$(CMD_DIR)
	GOOS=linux GOARCH=arm64 $(GO) build -tags whatsapp_native $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./$(CMD_DIR)
	GOOS=linux GOARCH=loong64 $(GO) build -tags whatsapp_native $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-loong64 ./$(CMD_DIR)
	GOOS=linux GOARCH=riscv64 $(GO) build -tags whatsapp_native $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-riscv64 ./$(CMD_DIR)
	GOOS=darwin GOARCH=arm64 $(GO) build -tags whatsapp_native $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./$(CMD_DIR)
	GOOS=windows GOARCH=amd64 $(GO) build -tags whatsapp_native $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./$(CMD_DIR)
## @$(GO) build $(GOFLAGS) -tags whatsapp_native $(LDFLAGS) -o $(BINARY_PATH) ./$(CMD_DIR)
	@echo "Build complete"
##	@ln -sf $(BINARY_NAME)-$(PLATFORM)-$(ARCH) $(BUILD_DIR)/$(BINARY_NAME)

## build-linux-arm: Build for Linux ARMv7 (e.g. Raspberry Pi Zero 2 W 32-bit)
build-linux-arm: generate
	@echo "Building for linux/arm (GOARM=7)..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm GOARM=7 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm ./$(CMD_DIR)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)-linux-arm"

## build-linux-arm64: Build for Linux ARM64 (e.g. Raspberry Pi Zero 2 W 64-bit)
build-linux-arm64: generate
	@echo "Building for linux/arm64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./$(CMD_DIR)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64"

## build-pi-zero: Build for Raspberry Pi Zero 2 W (32-bit and 64-bit)
build-pi-zero: build-linux-arm build-linux-arm64
	@echo "Pi Zero 2 W builds: $(BUILD_DIR)/$(BINARY_NAME)-linux-arm (32-bit), $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 (64-bit)"

## build-all: Build picoclaw for all platforms
build-all: generate
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./$(CMD_DIR)
	GOOS=linux GOARCH=arm GOARM=7 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm ./$(CMD_DIR)
	GOOS=linux GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./$(CMD_DIR)
	GOOS=linux GOARCH=loong64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-loong64 ./$(CMD_DIR)
	GOOS=linux GOARCH=riscv64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-riscv64 ./$(CMD_DIR)
	GOOS=linux GOARCH=arm GOARM=7 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-armv7 ./$(CMD_DIR)
	GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./$(CMD_DIR)
	GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./$(CMD_DIR)
	@echo "All builds complete"

## install: Install picoclaw to system (zero-config, no external services needed)
install: build verify-skills
	@echo "Installing $(BINARY_NAME)..."
	@echo "Platform: $(PLATFORM), Arch: $(ARCH)"
	@echo "Install prefix: $(INSTALL_PREFIX)"
	@echo "PICOCLAW_HOME: $(PICOCLAW_HOME)"
	@echo ""
	@mkdir -p $(INSTALL_BIN_DIR)
	# Copy binary with temporary suffix to ensure atomic update
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_BIN_DIR)/$(BINARY_NAME)$(INSTALL_TMP_SUFFIX)
	@chmod +x $(INSTALL_BIN_DIR)/$(BINARY_NAME)$(INSTALL_TMP_SUFFIX)
	@mv -f $(INSTALL_BIN_DIR)/$(BINARY_NAME)$(INSTALL_TMP_SUFFIX) $(INSTALL_BIN_DIR)/$(BINARY_NAME)
	@echo "✅ Installed binary to $(INSTALL_BIN_DIR)/$(BINARY_NAME)"
	@echo ""
	@echo "===================================="
	@echo "✅ PicoClaw installed successfully!"
	@echo "===================================="
	@echo ""
	@echo "📦 Binary: $(INSTALL_BIN_DIR)/$(BINARY_NAME)"
	@echo "🏠 PICOCLAW_HOME: $(PICOCLAW_HOME)"
	@echo "💾 Database: $(PICOCLAW_HOME)/picoclaw.db"
	@echo "🔍 Search: Built-in FTS5 keyword search (no external services needed)"
	@echo ""
	# Auto-add to PATH if needed
	$(call check_and_update_path)
	@echo ""
	@echo "🚀 Running onboard to setup configuration..."
	@$(INSTALL_BIN_DIR)/$(BINARY_NAME) onboard || true
	@echo ""
	@echo "Next steps:"
	@echo "  1. Edit config: $(PICOCLAW_HOME)/config.json"
	@echo "  2. Start services: picoclaw gateway"
	@echo ""
	@echo "⚠️  Important: Ensure '$(INSTALL_BIN_DIR)' is in your PATH"

# Function to check and update PATH
define check_and_update_path
	@echo "🔧 Checking PATH configuration..."
	@if [ "$(PLATFORM)" = "darwin" ]; then \
		if [ "$(INSTALL_PREFIX)" = "$(HOME)/.local" ]; then \
			if [ -f "$(HOME)/.zshrc" ]; then \
				if ! grep -q '$(INSTALL_BIN_DIR)' "$(HOME)/.zshrc" 2>/dev/null; then \
					echo 'export PATH="$(INSTALL_BIN_DIR):$$PATH"' >> "$(HOME)/.zshrc"; \
					echo "✅ Added $(INSTALL_BIN_DIR) to PATH in ~/.zshrc"; \
					echo "   Run: source ~/.zshrc"; \
				else \
					echo "✅ PATH already configured in ~/.zshrc"; \
				fi; \
			elif [ -f "$(HOME)/.bash_profile" ]; then \
				if ! grep -q '$(INSTALL_BIN_DIR)' "$(HOME)/.bash_profile" 2>/dev/null; then \
					echo 'export PATH="$(INSTALL_BIN_DIR):$$PATH"' >> "$(HOME)/.bash_profile"; \
					echo "✅ Added $(INSTALL_BIN_DIR) to PATH in ~/.bash_profile"; \
					echo "   Run: source ~/.bash_profile"; \
				else \
					echo "✅ PATH already configured in ~/.bash_profile"; \
				fi; \
			fi; \
		fi; \
	elif [ "$(PLATFORM)" = "linux" ]; then \
		if [ "$(INSTALL_PREFIX)" = "$(HOME)/.local" ]; then \
			SHELL_NAME=$$(basename "$$SHELL"); \
			if [ "$$SHELL_NAME" = "zsh" ] && [ -f "$(HOME)/.zshrc" ]; then \
				if ! grep -q '$(INSTALL_BIN_DIR)' "$(HOME)/.zshrc" 2>/dev/null; then \
					echo 'export PATH="$(INSTALL_BIN_DIR):$$PATH"' >> "$(HOME)/.zshrc"; \
					echo "✅ Added $(INSTALL_BIN_DIR) to PATH in ~/.zshrc"; \
					echo "   Run: source ~/.zshrc"; \
				else \
					echo "✅ PATH already configured in ~/.zshrc"; \
				fi; \
			elif [ -f "$(HOME)/.bashrc" ]; then \
				if ! grep -q '$(INSTALL_BIN_DIR)' "$(HOME)/.bashrc" 2>/dev/null; then \
					echo 'export PATH="$(INSTALL_BIN_DIR):$$PATH"' >> "$(HOME)/.bashrc"; \
					echo "✅ Added $(INSTALL_BIN_DIR) to PATH in ~/.bashrc"; \
					echo "   Run: source ~/.bashrc"; \
				else \
					echo "✅ PATH already configured in ~/.bashrc"; \
				fi; \
			fi; \
		elif [ "$(INSTALL_PREFIX)" = "/usr/local" ]; then \
			echo "✅ Using system prefix /usr/local (should already be in PATH)"; \
		fi; \
	else \
		echo "⚠️  Please ensure '$(INSTALL_BIN_DIR)' is in your PATH manually"; \
	fi
endef

## uninstall: Remove picoclaw from system
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@rm -f $(INSTALL_BIN_DIR)/$(BINARY_NAME)
	@echo "Removed binary from $(INSTALL_BIN_DIR)/$(BINARY_NAME)"
	@echo "Note: Only the executable file has been deleted."
	@echo "If you need to delete all configurations (config.json, workspace, etc.), run 'make uninstall-all'"

## uninstall-all: Remove picoclaw and all data
uninstall-all:
	@echo "Removing workspace and skills..."
	@rm -rf $(PICOCLAW_HOME)
	@echo "Removed workspace: $(PICOCLAW_HOME)"
	@echo "Complete uninstallation done!"

## clean: Remove build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete"

## vet: Run go vet for static analysis
vet: generate
	@$(GO) vet ./...

## test: Test Go code
test: generate
	@$(GO) test ./...

## fmt: Format Go code
fmt:
	@$(GOLANGCI_LINT) fmt

## lint: Run linters
lint:
	@$(GOLANGCI_LINT) run

## fix: Fix linting issues
fix:
	@$(GOLANGCI_LINT) run --fix

## deps: Download dependencies
deps:
	@$(GO) mod download
	@$(GO) mod verify

## update-deps: Update dependencies
update-deps:
	@$(GO) get -u ./...
	@$(GO) mod tidy

## check: Run vet, fmt, and verify dependencies
check: deps fmt vet test

## run: Build and run picoclaw
run: build
	@$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

## docker-build: Build Docker image (minimal Alpine-based)
docker-build:
	@echo "Building minimal Docker image (Alpine-based)..."
	docker compose -f docker/docker-compose.yml build picoclaw-agent picoclaw-gateway

## docker-build-full: Build Docker image with full MCP support (Node.js 24)
docker-build-full:
	@echo "Building full-featured Docker image (Node.js 24)..."
	docker compose -f docker/docker-compose.full.yml build picoclaw-agent picoclaw-gateway

## docker-test: Test MCP tools in Docker container
docker-test:
	@echo "Testing MCP tools in Docker..."
	@chmod +x scripts/test-docker-mcp.sh
	@./scripts/test-docker-mcp.sh

## docker-run: Run picoclaw gateway in Docker (Alpine-based)
docker-run:
	docker compose -f docker/docker-compose.yml --profile gateway up

## docker-run-full: Run picoclaw gateway in Docker (full-featured)
docker-run-full:
	docker compose -f docker/docker-compose.full.yml --profile gateway up

## docker-run-agent: Run picoclaw agent in Docker (interactive, Alpine-based)
docker-run-agent:
	docker compose -f docker/docker-compose.yml run --rm picoclaw-agent

## docker-run-agent-full: Run picoclaw agent in Docker (interactive, full-featured)
docker-run-agent-full:
	docker compose -f docker/docker-compose.full.yml run --rm picoclaw-agent

## docker-clean: Clean Docker images and volumes
docker-clean:
	docker compose -f docker/docker-compose.yml down -v
	docker compose -f docker/docker-compose.full.yml down -v
	docker rmi picoclaw:latest picoclaw:full 2>/dev/null || true

## kill: Kill all picoclaw processes
kill:
	@echo "Killing picoclaw processes..."
	@pkill -f "picoclaw gateway" 2>/dev/null || echo "No picoclaw gateway running"
	@pkill -f "picoclaw agent" 2>/dev/null || echo "No picoclaw agent running"
	@echo "All processes killed!"

## kill-force: Force kill all picoclaw processes (SIGKILL)
kill-force:
	@echo "Force killing picoclaw processes..."
	@pkill -9 -f "picoclaw gateway" 2>/dev/null || echo "No picoclaw gateway running"
	@pkill -9 -f "picoclaw agent" 2>/dev/null || echo "No picoclaw agent running"
	@echo "All processes force killed!"

## status: Check status of picoclaw processes
status:
	@echo "Checking picoclaw processes..."
	@echo "Picoclaw gateway:"
	@pgrep -f "picoclaw gateway" | head -5 || echo "  Not running"
	@echo "Picoclaw agent:"
	@pgrep -f "picoclaw agent" | head -5 || echo "  Not running"
	@echo ""
	@echo "Port usage:"
	@lsof -i :18790 2>/dev/null | grep LISTEN || echo "  Port 18790: Not in use"

## build-rag-tester: Build RAG memory tester
build-rag-tester:
	@echo "Building RAG memory tester..."
	@mkdir -p $(BUILD_DIR)
	@$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/rag-tester ./cmd/rag-tester
	@echo "Build complete: $(BUILD_DIR)/rag-tester"

## test-rag: Run RAG memory system tests
test-rag: build-rag-tester
	@echo "Running RAG memory system tests..."
	@$(BUILD_DIR)/rag-tester -config=$(PICOCLAW_HOME)/config.json

## test-rag-list: List available RAG test cases
test-rag-list: build-rag-tester
	@echo "Available RAG test cases:"
	@$(BUILD_DIR)/rag-tester -list

## test-rag-specific: Run specific RAG test (use TEST=name)
test-rag-specific: build-rag-tester
	@if [ -z "$(TEST)" ]; then \
		echo "Error: TEST variable not set"; \
		echo "Usage: make test-rag-specific TEST='Basic RAG Save and Retrieve'"; \
		exit 1; \
	fi
	@echo "Running RAG test: $(TEST)"
	@$(BUILD_DIR)/rag-tester -config=$(PICOCLAW_HOME)/config.json -test="$(TEST)"

## verify-skills: Verify and install missing builtin skills
verify-skills:
	@echo "Verifying and installing missing builtin skills..."
	@$(BUILD_DIR)/$(BINARY_NAME) skills install-builtin
	@echo "✅ Skills verification complete!"

## install-check: Check installation paths and environment
install-check:
	@echo "=== PicoClaw Installation Check ==="
	@echo ""
	@echo "Platform:"
	@echo "  OS:        $(PLATFORM)"
	@echo "  Arch:      $(ARCH)"
	@echo "  OS Type:   $(UNAME_S)"
	@echo ""
	@echo "Paths:"
	@echo "  HOME:           $(USER_HOME)"
	@echo "  PICOCLAW_HOME:  $(PICOCLAW_HOME)"
	@echo "  INSTALL_PREFIX: $(INSTALL_PREFIX)"
	@echo "  BIN_DIR:        $(INSTALL_BIN_DIR)"
	@echo "  BINARY_PATH:    $(BINARY_PATH)"
	@echo ""
	@echo "Environment Variables:"
	@echo "  PICOCLAW_HOME=$(PICOCLAW_HOME)"
	@echo "  INSTALL_PREFIX=$(INSTALL_PREFIX)"
	@echo ""
	@echo "Checking directories..."
	@test -d "$(USER_HOME)" && echo "  ✓ HOME exists" || echo "  ✗ HOME not found"
	@test -d "$(PICOCLAW_HOME)" && echo "  ✓ PICOCLAW_HOME exists" || echo "  ○ PICOCLAW_HOME will be created"
	@test -d "$(INSTALL_BIN_DIR)" && echo "  ✓ BIN_DIR exists" || echo "  ○ BIN_DIR will be created"
	@test -f "$(BINARY_PATH)" && echo "  ✓ Binary built" || echo "  ✗ Binary not built (run 'make build')"
	@echo ""
	@echo "PATH check:"
	@printf "  Current PATH contains BIN_DIR: "
	@if echo "$(PATH)" | grep -q "$(INSTALL_BIN_DIR)"; then echo "YES"; else echo "NO (will be added during install)"; fi
	@echo ""

## install-system: Install picoclaw to system-wide location (/usr/local)
install-system:
	$(MAKE) install INSTALL_PREFIX=/usr/local

## install-user: Install picoclaw to user home (~/.local)
install-user:
	$(MAKE) install INSTALL_PREFIX=$(HOME)/.local

## reinstall: Reinstall picoclaw (clean and install)
reinstall: clean uninstall build verify-skills install

## help: Show this help message
help:
	@echo "picoclaw Makefile - Cross-Platform Build System"
	@echo ""
	@echo "Usage:"
	@echo "  make [target] [VARIABLE=value]"
	@echo ""
	@echo "Installation Targets:"
	@echo "  install              Install picoclaw (auto-detects best location)"
	@echo "  install-user         Install to ~/.local/bin (user-only)"
	@echo "  install-system       Install to /usr/local/bin (system-wide, requires root)"
	@echo "  reinstall            Clean build and reinstall"
	@echo "  install-check        Check installation paths and environment"
	@echo "  uninstall            Remove picoclaw binary"
	@echo "  uninstall-all        Remove picoclaw and all data (~/.picoclaw)"
	@echo ""
	@echo "Build Targets:"
	@echo "  build                Build for current platform ($(PLATFORM)/$(ARCH))"
	@echo "  build-all            Build for all platforms"
	@echo "  build-whatsapp-native Build with WhatsApp native support"
	@echo "  clean                Remove build artifacts"
	@echo ""
	@echo "Development Targets:"
	@echo "  test-rag             Run RAG memory system tests"
	@echo "  test-rag-list        List available RAG test cases"
	@echo ""
	@echo "Maintenance Targets:"
	@echo "  kill                 Kill all picoclaw processes"
	@echo "  kill-force           Force kill all processes"
	@echo "  status               Check process status"
	@echo "  test                 Run tests"
	@echo "  lint                 Run linters"
	@echo ""
	@echo "Cross-Platform Installation Examples:"
	@echo "  # macOS - user install (recommended)"
	@echo "  make install-user"
	@echo ""
	@echo "  # Ubuntu/Debian - user install (recommended)"
	@echo "  make install-user"
	@echo ""
	@echo "  # Ubuntu/Debian - system install (requires sudo)"
	@echo "  sudo make install-system"
	@echo ""
	@echo "  # Windows - using Git Bash or WSL"
	@echo "  make install PICOCLAW_HOME='C:/Users/YourName/.picoclaw'"
	@echo ""
	@echo "  # Custom install location"
	@echo "  make install INSTALL_PREFIX=/opt/picoclaw"
	@echo ""
	@echo "Environment Variables:"
	@echo "  PICOCLAW_HOME         # Data/config directory (default: ~/.picoclaw)"
	@echo "  INSTALL_PREFIX        # Installation prefix (default: auto-detect)"
	@echo "  WORKSPACE_DIR         # Workspace directory (default: ~/.picoclaw/workspace)"
	@echo "  VERSION               # Version string (default: git describe)"
	@echo ""
	@echo "Current Configuration:"
	@echo "  Platform: $(PLATFORM)/$(ARCH)"
	@echo "  Binary: $(BINARY_PATH)"
	@echo "  PICOCLAW_HOME: $(PICOCLAW_HOME)"
	@echo "  Install Prefix: $(INSTALL_PREFIX)"
