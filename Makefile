# SSH Secret Keeper Makefile

# Build information - Git tag-driven versioning
VERSION := $(shell git describe --tags --exact-match 2>/dev/null | sed 's/^v//' || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_HASH := $(shell git rev-parse --short HEAD)
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
GIT_TAG := $(shell git describe --tags --exact-match 2>/dev/null || echo "")

# Go build flags
LDFLAGS := -ldflags "-X 'github.com/rzago/ssh-secret-keeper/internal/cmd.Version=$(VERSION)' \
                     -X 'github.com/rzago/ssh-secret-keeper/internal/cmd.BuildTime=$(BUILD_TIME)' \
                     -X 'github.com/rzago/ssh-secret-keeper/internal/cmd.GitHash=$(GIT_HASH)'"

# Build settings
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
BINARY_NAME := sshsk
BUILD_DIR := bin

# Default target
.DEFAULT_GOAL := build


# Build
.PHONY: build
build:
	@echo "Building $(BINARY_NAME) for $(GOOS)/$(GOARCH)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) cmd/main.go
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for all platforms including container images
.PHONY: build-all
build-all: build-binaries container-build-all
	@echo "All builds complete (binaries + all container images)"

# Build binaries for all platforms
.PHONY: build-binaries
build-binaries:
	@echo "Building binaries for all platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 make build
	@mv $(BUILD_DIR)/$(BINARY_NAME) $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64
	GOOS=linux GOARCH=arm64 make build
	@mv $(BUILD_DIR)/$(BINARY_NAME) $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64
	GOOS=darwin GOARCH=amd64 make build
	@mv $(BUILD_DIR)/$(BINARY_NAME) $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64
	GOOS=darwin GOARCH=arm64 make build
	@mv $(BUILD_DIR)/$(BINARY_NAME) $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64
	@echo "All binary builds complete"

# Run
.PHONY: run
run: build
	./$(BUILD_DIR)/$(BINARY_NAME)


# Test
.PHONY: test
test:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	@echo "Tests passed"

# Test coverage report
.PHONY: test-coverage
test-coverage: test
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Test coverage percentage check
.PHONY: test-coverage-check
test-coverage-check: test
	@echo "Checking coverage percentage..."
	@go tool cover -func=coverage.out | grep total | awk '{print "Total coverage: " $$3}'
	@go tool cover -func=coverage.out | grep total | awk '{coverage=$$3; gsub(/%/, "", coverage); if(coverage < 0) {print "❌ Coverage below 0% target: " coverage "%"; exit 1} else {print "✅ Coverage check passed: " coverage "%"}}'

# Test with short flag for quick feedback
.PHONY: test-short
test-short:
	@echo "Running short tests..."
	go test -short -race ./...

# Benchmark tests
.PHONY: test-bench
test-bench:
	@echo "Running benchmark tests..."
	go test -bench=. -benchmem ./...


# Generate
.PHONY: generate
generate:
	@echo "Running go generate..."
	go generate ./...
	@echo "Generation complete"

# Clean
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	@echo "Clean complete"

# Install locally
.PHONY: install
install:
	@echo "Installing $(BINARY_NAME) for $(shell go env GOOS)/$(shell go env GOARCH)..."
	@CURRENT_GOOS=$$(go env GOOS); \
	CURRENT_GOARCH=$$(go env GOARCH); \
	ARCH_BINARY="$(BUILD_DIR)/$(BINARY_NAME)-$${CURRENT_GOOS}-$${CURRENT_GOARCH}"; \
	INSTALL_NAME="$(BINARY_NAME)"; \
	GENERIC_BINARY="$(BUILD_DIR)/$(BINARY_NAME)"; \
	if [ -f "$$ARCH_BINARY" ]; then \
		echo "Found architecture-specific binary: $$ARCH_BINARY"; \
		sudo cp "$$ARCH_BINARY" /usr/local/bin/$$INSTALL_NAME; \
	elif [ -f "$$GENERIC_BINARY" ]; then \
		echo "Found generic binary: $$GENERIC_BINARY"; \
		sudo cp "$$GENERIC_BINARY" /usr/local/bin/$$INSTALL_NAME; \
	else \
		echo "No binary found. Building for current platform..."; \
		mkdir -p $(BUILD_DIR); \
		CGO_ENABLED=0 GOOS=$$CURRENT_GOOS GOARCH=$$CURRENT_GOARCH go build $(LDFLAGS) -o $$GENERIC_BINARY cmd/main.go; \
		echo "Build complete: $$GENERIC_BINARY"; \
		sudo cp "$$GENERIC_BINARY" /usr/local/bin/$$INSTALL_NAME; \
	fi; \
	echo "Installed to /usr/local/bin/$$INSTALL_NAME"

# Uninstall
.PHONY: uninstall
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "Uninstalled"

# Container builds (Docker and Podman support)
.PHONY: container-build
container-build:
	@echo "Building container image (auto-detecting runtime)..."
	@if command -v podman > /dev/null 2>&1; then \
		echo "Using Podman..."; \
		$(MAKE) podman-build; \
	elif command -v docker > /dev/null 2>&1; then \
		echo "Using Docker..."; \
		$(MAKE) docker-build; \
	else \
		echo "Neither Docker nor Podman found. Please install one."; \
		exit 1; \
	fi

# Build container images with both Docker and Podman
.PHONY: container-build-all
container-build-all:
	@echo "Building container images with both Docker and Podman..."
	@DOCKER_AVAILABLE=false; PODMAN_AVAILABLE=false; \
	if command -v docker > /dev/null 2>&1; then DOCKER_AVAILABLE=true; fi; \
	if command -v podman > /dev/null 2>&1; then PODMAN_AVAILABLE=true; fi; \
	if [ "$$DOCKER_AVAILABLE" = "false" ] && [ "$$PODMAN_AVAILABLE" = "false" ]; then \
		echo "Neither Docker nor Podman found. Please install at least one."; \
		exit 1; \
	fi; \
	if [ "$$DOCKER_AVAILABLE" = "true" ]; then \
		echo "Building Docker images..."; \
		$(MAKE) docker-build; \
	else \
		echo "Docker not available, skipping Docker build"; \
	fi; \
	if [ "$$PODMAN_AVAILABLE" = "true" ]; then \
		echo "Building Podman images..."; \
		$(MAKE) podman-build; \
	else \
		echo "Podman not available, skipping Podman build"; \
	fi; \
	echo "Container build complete (Docker: $$DOCKER_AVAILABLE, Podman: $$PODMAN_AVAILABLE)"

# Build container image based on current git branch
.PHONY: container-build-branch
container-build-branch:
	@echo "Building container image for branch: $(GIT_BRANCH)..."
	@if command -v podman > /dev/null 2>&1; then \
		echo "Using Podman..."; \
		$(MAKE) podman-build-branch; \
	elif command -v docker > /dev/null 2>&1; then \
		echo "Using Docker..."; \
		$(MAKE) docker-build-branch; \
	else \
		echo "Neither Docker nor Podman found. Please install one."; \
		exit 1; \
	fi

.PHONY: docker-build
docker-build:
	@echo "Building Docker images..."
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_HASH=$(GIT_HASH) \
		-t ssh-secret-keeper:$(VERSION) \
		-t ssh-secret-keeper:latest \
		.
	@echo "Docker images built: ssh-secret-keeper:$(VERSION), ssh-secret-keeper:latest"

.PHONY: docker-build-branch
docker-build-branch:
	@echo "Building Docker images for branch $(GIT_BRANCH)..."
	$(eval BRANCH_TAG := $(shell echo $(GIT_BRANCH) | sed 's/[^a-zA-Z0-9._-]/-/g'))
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_HASH=$(GIT_HASH) \
		-t ssh-secret-keeper:$(BRANCH_TAG) \
		-t ssh-secret-keeper:$(BRANCH_TAG)-$(GIT_HASH) \
		$(if $(filter main master,$(GIT_BRANCH)),-t ssh-secret-keeper:latest) \
		.
	@echo "Docker images built: ssh-secret-keeper:$(BRANCH_TAG), ssh-secret-keeper:$(BRANCH_TAG)-$(GIT_HASH)"

.PHONY: podman-build
podman-build:
	@echo "Building Podman images..."
	podman build \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_HASH=$(GIT_HASH) \
		-t ssh-secret-keeper:$(VERSION) \
		-t ssh-secret-keeper:latest \
		.
	@echo "Podman images built: ssh-secret-keeper:$(VERSION), ssh-secret-keeper:latest"

.PHONY: podman-build-branch
podman-build-branch:
	@echo "Building Podman images for branch $(GIT_BRANCH)..."
	$(eval BRANCH_TAG := $(shell echo $(GIT_BRANCH) | sed 's/[^a-zA-Z0-9._-]/-/g'))
	podman build \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_HASH=$(GIT_HASH) \
		-t ssh-secret-keeper:$(BRANCH_TAG) \
		-t ssh-secret-keeper:$(BRANCH_TAG)-$(GIT_HASH) \
		$(if $(filter main master,$(GIT_BRANCH)),-t ssh-secret-keeper:latest) \
		.
	@echo "Podman images built: ssh-secret-keeper:$(BRANCH_TAG), ssh-secret-keeper:$(BRANCH_TAG)-$(GIT_HASH)"

# Legacy docker target for backward compatibility
.PHONY: docker
docker: docker-build

# Release preparation and validation
.PHONY: release-prepare
release-prepare:
	@echo "Preparing release from git tag..."
	@echo "Current branch: $(GIT_BRANCH)"
	@echo "Current commit: $(GIT_HASH)"
	@if [ -z "$(GIT_TAG)" ]; then \
		echo "❌ No git tag found on current commit. Create a tag first:"; \
		echo "  git tag -a v1.2.3 -m 'Release version 1.2.3'"; \
		exit 1; \
	fi
	@echo "✅ Git tag found: $(GIT_TAG)"
	@echo "✅ Version will be: $(VERSION)"
	@if git diff --quiet HEAD; then \
		echo "✅ Working directory is clean"; \
	else \
		echo "❌ Working directory has uncommitted changes"; \
		exit 1; \
	fi
	@which goreleaser > /dev/null || (echo "❌ goreleaser not found. Install with: go install github.com/goreleaser/goreleaser@latest" && exit 1)
	@echo "✅ Release preparation complete"

# Validate release configuration
.PHONY: release-check
release-check:
	@echo "Validating release configuration..."
	@which goreleaser > /dev/null || (echo "❌ goreleaser not found. Install with: go install github.com/goreleaser/goreleaser@latest" && exit 1)
	goreleaser check
	@echo "✅ GoReleaser configuration is valid"

# Create git tag
.PHONY: tag-release
tag-release:
	@if [ -z "$(V)" ]; then \
		echo "❌ Please specify version: make tag-release V=1.2.3"; \
		exit 1; \
	fi
	@echo "Creating release tag v$(V)..."
	@if git rev-parse "v$(V)" >/dev/null 2>&1; then \
		echo "❌ Tag v$(V) already exists"; \
		exit 1; \
	fi
	@if git diff --quiet HEAD; then \
		echo "✅ Working directory is clean"; \
	else \
		echo "❌ Working directory has uncommitted changes"; \
		exit 1; \
	fi
	git tag -a "v$(V)" -m "Release version $(V) from branch $(GIT_BRANCH)"
	@echo "✅ Release tag v$(V) created"
	@echo "Push with: git push origin v$(V)"

# Push tag to trigger release
.PHONY: push-tag
push-tag:
	@if [ -z "$(V)" ]; then \
		echo "❌ Please specify version: make push-tag V=1.2.3"; \
		exit 1; \
	fi
	@if ! git rev-parse "v$(V)" >/dev/null 2>&1; then \
		echo "❌ Tag v$(V) does not exist. Create it first with 'make tag-release V=$(V)'"; \
		exit 1; \
	fi
	@echo "Pushing tag v$(V) to trigger release..."
	git push origin "v$(V)"
	@echo "✅ Tag pushed. Release workflow should start automatically."
	@echo "Monitor the release at: https://github.com/rafaelvzago/ssh-vault-keeper/releases"

# Local release (requires existing tag)
.PHONY: release-local
release-local:
	@echo "Creating local release..."
	@which goreleaser > /dev/null || (echo "❌ goreleaser not found. Install with: go install github.com/goreleaser/goreleaser@latest" && exit 1)
	@if ! git describe --exact-match --tags $$(git log -n1 --pretty='%h') >/dev/null 2>&1; then \
		echo "❌ Current commit is not tagged. Create and checkout a tag first."; \
		exit 1; \
	fi
	goreleaser release --clean
	@echo "✅ Release created"

# Complete release workflow
.PHONY: release
release:
	@if [ -z "$(V)" ]; then \
		echo "❌ Please specify version: make release V=1.2.3"; \
		exit 1; \
	fi
	@echo "Starting complete release workflow for v$(V)..."
	make test
	make tag-release V=$(V)
	make push-tag V=$(V)
	@echo "✅ Complete release workflow finished"
	@echo "Monitor the release at: https://github.com/rafaelvzago/ssh-vault-keeper/releases"

# Release with container images
.PHONY: release-with-images
release-with-images:
	@if [ -z "$(V)" ]; then \
		echo "❌ Please specify version: make release-with-images V=1.2.3"; \
		exit 1; \
	fi
	@echo "Starting release with container images for v$(V)..."
	make test
	make container-build-branch
	make tag-release V=$(V)
	make push-tag V=$(V)
	@echo "✅ Release with container images completed"

# Release snapshot (local testing)
.PHONY: release-snapshot
release-snapshot:
	@echo "Creating snapshot release..."
	@which goreleaser > /dev/null || (echo "❌ goreleaser not found. Install with: go install github.com/goreleaser/goreleaser@latest" && exit 1)
	goreleaser release --snapshot --clean
	@echo "✅ Snapshot release created in dist/"

# Show help
.PHONY: help
help:
	@echo "SSH Secret Keeper - Available Commands"
	@echo "===================================="
	@echo ""
	@echo "Build & Install:"
	@echo "  build              Build for current platform"
	@echo "  build-all          Build for all platforms + all container images"
	@echo "  build-binaries     Build binaries for all platforms only"
	@echo "  install            Install to /usr/local/bin (auto-detects architecture)"
	@echo "  uninstall          Remove from /usr/local/bin"
	@echo "  container-build    Build container image (auto-detect Docker/Podman)"
	@echo "  container-build-all Build container images with both Docker and Podman"
	@echo "  container-build-branch Build container image tagged with current branch"
	@echo "  docker-build       Build Docker images specifically"
	@echo "  docker-build-branch Build Docker images tagged with current branch"
	@echo "  podman-build       Build Podman images specifically"
	@echo "  podman-build-branch Build Podman images tagged with current branch"
	@echo "  docker             Build Docker images (legacy alias)"
	@echo ""
	@echo "Testing:"
	@echo "  test               Run unit tests with coverage"
	@echo "  test-coverage      Generate HTML coverage report"
	@echo "  test-coverage-check Show coverage percentage (no threshold)"
	@echo "  test-short         Run tests with short flag"
	@echo "  test-bench         Run benchmark tests"
	@echo ""
	@echo "Release Management:"
	@echo "  release-prepare    Validate environment for release"
	@echo "  release-check      Validate GoReleaser configuration"
	@echo "  tag-release        Create git tag for release (use VERSION=x.y.z)"
	@echo "  push-tag          Push tag to trigger automated release"
	@echo "  release           Complete release workflow (test + tag + push)"
	@echo "  release-local     Create release from existing tag (local)"
	@echo "  release-with-images Complete release with container images"
	@echo "  release-snapshot  Create snapshot release for testing"
	@echo ""
	@echo "Maintenance:"
	@echo "  generate           Run go generate"
	@echo "  clean              Clean build artifacts"
	@echo "  help               Show this help"
	@echo ""
	@echo "Examples:"
	@echo "  make build VERSION=1.0.1"
	@echo "  make build-all                 # Builds binaries + all container images"
	@echo "  make build-binaries            # Builds only binaries"
	@echo "  make container-build           # Auto-detect and build with one runtime"
	@echo "  make container-build-all       # Build with both Docker and Podman"
	@echo "  make container-build-branch"
	@echo "  make test"
	@echo ""
	@echo "Release Examples:"
	@echo "  make release-check                    # Validate configuration"
	@echo "  make release-prepare                  # Prepare for release (requires git tag)"
	@echo "  make release V=1.2.1                 # Complete release workflow"
	@echo "  make tag-release V=1.2.1             # Create tag only"
	@echo "  make push-tag V=1.2.1                # Push existing tag"
	@echo "  make release-snapshot                 # Test release locally"
