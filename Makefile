# SSH Vault Keeper Makefile

# Build information
VERSION ?= 1.0.0
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_HASH := $(shell git rev-parse --short HEAD)
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
GIT_TAG := $(shell git describe --tags --exact-match 2>/dev/null || echo "")

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


# Build
.PHONY: build
build:
	@echo "Building $(BINARY_NAME) for $(GOOS)/$(GOARCH)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) cmd/main.go
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for all platforms
.PHONY: build-all
build-all:
	@echo "Building for all platforms..."
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
	@echo "All builds complete"

# Run
.PHONY: run
run: build
	./$(BUILD_DIR)/$(BINARY_NAME)


# Test
.PHONY: test
test:
	@echo "Running tests..."
	go test -v -race ./...
	@echo "Tests passed"


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
install: build
	@echo "Installing $(BINARY_NAME)..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "Installed to /usr/local/bin/$(BINARY_NAME)"

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
		-t ssh-vault-keeper:$(VERSION) \
		-t ssh-vault-keeper:latest \
		.
	@echo "Docker images built: ssh-vault-keeper:$(VERSION), ssh-vault-keeper:latest"

.PHONY: docker-build-branch
docker-build-branch:
	@echo "Building Docker images for branch $(GIT_BRANCH)..."
	$(eval BRANCH_TAG := $(shell echo $(GIT_BRANCH) | sed 's/[^a-zA-Z0-9._-]/-/g'))
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_HASH=$(GIT_HASH) \
		-t ssh-vault-keeper:$(BRANCH_TAG) \
		-t ssh-vault-keeper:$(BRANCH_TAG)-$(GIT_HASH) \
		$(if $(filter main master,$(GIT_BRANCH)),-t ssh-vault-keeper:latest) \
		.
	@echo "Docker images built: ssh-vault-keeper:$(BRANCH_TAG), ssh-vault-keeper:$(BRANCH_TAG)-$(GIT_HASH)"

.PHONY: podman-build
podman-build:
	@echo "Building Podman images..."
	podman build \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_HASH=$(GIT_HASH) \
		-t ssh-vault-keeper:$(VERSION) \
		-t ssh-vault-keeper:latest \
		.
	@echo "Podman images built: ssh-vault-keeper:$(VERSION), ssh-vault-keeper:latest"

.PHONY: podman-build-branch
podman-build-branch:
	@echo "Building Podman images for branch $(GIT_BRANCH)..."
	$(eval BRANCH_TAG := $(shell echo $(GIT_BRANCH) | sed 's/[^a-zA-Z0-9._-]/-/g'))
	podman build \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_HASH=$(GIT_HASH) \
		-t ssh-vault-keeper:$(BRANCH_TAG) \
		-t ssh-vault-keeper:$(BRANCH_TAG)-$(GIT_HASH) \
		$(if $(filter main master,$(GIT_BRANCH)),-t ssh-vault-keeper:latest) \
		.
	@echo "Podman images built: ssh-vault-keeper:$(BRANCH_TAG), ssh-vault-keeper:$(BRANCH_TAG)-$(GIT_HASH)"

# Legacy docker target for backward compatibility
.PHONY: docker
docker: docker-build

# Create git tag based on current branch and version
.PHONY: tag-release
tag-release:
	@echo "Creating release tag v$(VERSION) for branch $(GIT_BRANCH)..."
	@if [ "$(GIT_BRANCH)" != "main" ] && [ "$(GIT_BRANCH)" != "master" ]; then \
		echo "Warning: Creating release from non-main branch ($(GIT_BRANCH))"; \
	fi
	@if git rev-parse "v$(VERSION)" >/dev/null 2>&1; then \
		echo "Tag v$(VERSION) already exists. Use 'make tag-release VERSION=new-version'"; \
		exit 1; \
	fi
	git tag -a "v$(VERSION)" -m "Release version $(VERSION) from branch $(GIT_BRANCH)"
	@echo "Release tag v$(VERSION) created"
	@echo "Push with: git push origin v$(VERSION)"

# Release (requires goreleaser)
.PHONY: release
release: tag-release
	@echo "Creating release..."
	@which goreleaser > /dev/null || (echo "goreleaser not found. Install with: go install github.com/goreleaser/goreleaser@latest" && exit 1)
	goreleaser release --clean
	@echo "Release created"

# Release with container images
.PHONY: release-with-images
release-with-images: container-build-branch release
	@echo "Release with container images created"

# Release snapshot (local testing)
.PHONY: release-snapshot
release-snapshot:
	@echo "Creating snapshot release..."
	@which goreleaser > /dev/null || (echo "goreleaser not found. Install with: go install github.com/goreleaser/goreleaser@latest" && exit 1)
	goreleaser release --snapshot --clean
	@echo "Snapshot release created"

# Show help
.PHONY: help
help:
	@echo "SSH Vault Keeper - Available Commands"
	@echo "===================================="
	@echo ""
	@echo "Build & Install:"
	@echo "  build              Build for current platform"
	@echo "  build-all          Build for all platforms"
	@echo "  install            Install to /usr/local/bin"
	@echo "  uninstall          Remove from /usr/local/bin"
	@echo "  container-build    Build container image (auto-detect Docker/Podman)"
	@echo "  container-build-branch Build container image tagged with current branch"
	@echo "  docker-build       Build Docker images specifically"
	@echo "  docker-build-branch Build Docker images tagged with current branch"
	@echo "  podman-build       Build Podman images specifically"
	@echo "  podman-build-branch Build Podman images tagged with current branch"
	@echo "  docker             Build Docker images (legacy alias)"
	@echo ""
	@echo "Testing:"
	@echo "  test               Run unit tests"
	@echo ""
	@echo "Release:"
	@echo "  tag-release        Create git tag for release (use VERSION=x.y.z)"
	@echo "  release            Create release with goreleaser (includes tagging)"
	@echo "  release-with-images Create release with container images"
	@echo "  release-snapshot   Create snapshot release for testing"
	@echo ""
	@echo "Maintenance:"
	@echo "  generate           Run go generate"
	@echo "  clean              Clean build artifacts"
	@echo "  help               Show this help"
	@echo ""
	@echo "Examples:"
	@echo "  make build VERSION=1.0.1"
	@echo "  make container-build-branch"
	@echo "  make tag-release VERSION=1.2.0"
	@echo "  make release-with-images VERSION=1.2.0"
	@echo "  make test"
