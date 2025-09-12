# SSH Vault Keeper Makefile

# Build information
VERSION ?= 1.0.0
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_HASH := $(shell git rev-parse --short HEAD)

# Go build flags
LDFLAGS := -ldflags "-X 'github.com/rzago/ssh-vault-keeper/internal/cmd.Version=$(VERSION)' \
                     -X 'github.com/rzago/ssh-vault-keeper/internal/cmd.BuildTime=$(BUILD_TIME)' \
                     -X 'github.com/rzago/ssh-vault-keeper/internal/cmd.GitHash=$(GIT_HASH)'"

# Build settings
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
BINARY_NAME := ssh-vault-keeper
BUILD_DIR := bin

# Default target
.DEFAULT_GOAL := build

# Development setup
.PHONY: dev-setup
dev-setup:
	@echo "🔧 Setting up development environment..."
	go mod download
	go mod tidy
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.55.2)
	@echo "✅ Development environment ready"

# Build
.PHONY: build
build:
	@echo "🏗️  Building $(BINARY_NAME) for $(GOOS)/$(GOARCH)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) cmd/main.go
	@echo "✅ Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for all platforms
.PHONY: build-all
build-all:
	@echo "🏗️  Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 make build
	@mv $(BUILD_DIR)/$(BINARY_NAME) $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64
	GOOS=linux GOARCH=arm64 make build  
	@mv $(BUILD_DIR)/$(BINARY_NAME) $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64
	GOOS=darwin GOARCH=amd64 make build
	@mv $(BUILD_DIR)/$(BINARY_NAME) $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64
	GOOS=darwin GOARCH=arm64 make build
	@mv $(BUILD_DIR)/$(BINARY_NAME) $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64
	GOOS=windows GOARCH=amd64 make build
	@mv $(BUILD_DIR)/$(BINARY_NAME) $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe
	@echo "✅ All builds complete"

# Run
.PHONY: run
run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

# Development run with sample config
.PHONY: dev-run
dev-run: build
	@echo "🚀 Running in development mode..."
	SSH_VAULT_LOGGING_LEVEL=debug ./$(BUILD_DIR)/$(BINARY_NAME) status

# Test
.PHONY: test
test:
	@echo "🧪 Running tests..."
	go test -v -race ./...
	@echo "✅ Tests passed"

# Test with coverage
.PHONY: test-coverage
test-coverage:
	@echo "🧪 Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

# Integration tests (requires Vault)
.PHONY: test-integration
test-integration:
	@echo "🧪 Running integration tests..."
	@if [ -z "$(VAULT_ADDR)" ]; then \
		echo "❌ VAULT_ADDR environment variable required"; \
		exit 1; \
	fi
	@if [ -z "$(VAULT_TOKEN)" ]; then \
		echo "❌ VAULT_TOKEN environment variable required"; \
		exit 1; \
	fi
	go test -v -tags=integration ./...
	@echo "✅ Integration tests passed"

# Lint
.PHONY: lint
lint:
	@echo "🔍 Running linter..."
	golangci-lint run
	@echo "✅ Linting passed"

# Format code
.PHONY: fmt
fmt:
	@echo "📝 Formatting code..."
	go fmt ./...
	goimports -w .
	@echo "✅ Code formatted"

# Generate
.PHONY: generate
generate:
	@echo "🔄 Running go generate..."
	go generate ./...
	@echo "✅ Generation complete"

# Clean
.PHONY: clean
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	@echo "✅ Clean complete"

# Install locally
.PHONY: install
install: build
	@echo "📦 Installing $(BINARY_NAME)..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "✅ Installed to /usr/local/bin/$(BINARY_NAME)"

# Uninstall
.PHONY: uninstall  
uninstall:
	@echo "🗑️  Uninstalling $(BINARY_NAME)..."
	sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "✅ Uninstalled"

# Docker build
.PHONY: docker
docker:
	@echo "🐳 Building Docker image..."
	docker build -t ssh-vault-keeper:$(VERSION) .
	docker build -t ssh-vault-keeper:latest .
	@echo "✅ Docker images built"

# Release (requires goreleaser)
.PHONY: release
release:
	@echo "🚀 Creating release..."
	@which goreleaser > /dev/null || (echo "❌ goreleaser not found. Install with: go install github.com/goreleaser/goreleaser@latest" && exit 1)
	goreleaser release --clean
	@echo "✅ Release created"

# Release snapshot (local testing)
.PHONY: release-snapshot
release-snapshot:
	@echo "📸 Creating snapshot release..."
	@which goreleaser > /dev/null || (echo "❌ goreleaser not found. Install with: go install github.com/goreleaser/goreleaser@latest" && exit 1)  
	goreleaser release --snapshot --clean
	@echo "✅ Snapshot release created"

# Initialize development Vault (requires Docker)
.PHONY: dev-vault
dev-vault:
	@echo "🏦 Starting development Vault server..."
	docker run --rm -d \
		--name ssh-vault-keeper-vault \
		--cap-add=IPC_LOCK \
		-p 8200:8200 \
		-e 'VAULT_DEV_ROOT_TOKEN_ID=dev-root-token' \
		-e 'VAULT_DEV_LISTEN_ADDRESS=0.0.0.0:8200' \
		vault:latest
	@echo "✅ Vault started at http://localhost:8200"
	@echo "   Root token: dev-root-token"
	@echo "   Stop with: make stop-dev-vault"

# Stop development Vault
.PHONY: stop-dev-vault
stop-dev-vault:
	@echo "🛑 Stopping development Vault server..."
	docker stop ssh-vault-keeper-vault || true
	@echo "✅ Vault stopped"

# Show help
.PHONY: help
help:
	@echo "SSH Vault Keeper - Available Commands"
	@echo "===================================="
	@echo ""
	@echo "Development:"
	@echo "  dev-setup          Set up development environment"
	@echo "  dev-run            Run in development mode"
	@echo "  dev-vault          Start development Vault server"
	@echo "  stop-dev-vault     Stop development Vault server"
	@echo ""
	@echo "Build & Install:"
	@echo "  build              Build for current platform"
	@echo "  build-all          Build for all platforms"
	@echo "  install            Install to /usr/local/bin"
	@echo "  uninstall          Remove from /usr/local/bin"
	@echo "  docker             Build Docker images"
	@echo ""
	@echo "Testing & Quality:"
	@echo "  test               Run unit tests"
	@echo "  test-coverage      Run tests with coverage report"
	@echo "  test-integration   Run integration tests (requires Vault)"
	@echo "  lint               Run linter"
	@echo "  fmt                Format code"
	@echo ""
	@echo "Release:"
	@echo "  release            Create release with goreleaser"
	@echo "  release-snapshot   Create snapshot release for testing"
	@echo ""
	@echo "Maintenance:"  
	@echo "  generate           Run go generate"
	@echo "  clean              Clean build artifacts"
	@echo "  help               Show this help"
	@echo ""
	@echo "Examples:"
	@echo "  make build VERSION=1.0.1"
	@echo "  VAULT_ADDR=http://localhost:8200 VAULT_TOKEN=dev-root-token make test-integration"
