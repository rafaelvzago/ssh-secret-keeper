# SSH Vault Keeper Site - Development Makefile
.DEFAULT_GOAL := dev

# Node/npm commands
NPM := npm

# Development
.PHONY: install
install:
	@echo "Installing dependencies..."
	$(NPM) install
	@echo "✅ Dependencies installed"

.PHONY: dev
dev: install
	@echo "Starting development server..."
	$(NPM) run dev

# Build (for testing build process)
.PHONY: build
build: install
	@echo "Building for testing..."
	$(NPM) run build
	@echo "✅ Build complete in dist/"

.PHONY: preview
preview: build
	@echo "Starting preview server..."
	$(NPM) run preview

# Quality assurance
.PHONY: lint
lint: install
	@echo "Running ESLint..."
	$(NPM) run lint

.PHONY: lint-fix
lint-fix: install
	@echo "Running ESLint with auto-fix..."
	$(NPM) run lint:fix

.PHONY: type-check
type-check: install
	@echo "Running TypeScript type check..."
	$(NPM) run type-check

# Maintenance
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf dist node_modules
	@echo "✅ Clean complete"

.PHONY: clean-build
clean-build:
	@echo "Cleaning build directory..."
	rm -rf dist
	@echo "✅ Build directory cleaned"

# Help
.PHONY: help
help:
	@echo "SSH Vault Keeper Site - Development Commands"
	@echo "==========================================="
	@echo ""
	@echo "Development:"
	@echo "  dev                Start development server (default)"
	@echo "  install            Install dependencies"
	@echo "  build              Build for testing"
	@echo "  preview            Preview build"
	@echo ""
	@echo "Quality:"
	@echo "  lint               Run ESLint"
	@echo "  lint-fix           Run ESLint with auto-fix"
	@echo "  type-check         Run TypeScript type check"
	@echo ""
	@echo "Maintenance:"
	@echo "  clean              Clean all artifacts"
	@echo "  clean-build        Clean build directory only"
	@echo "  help               Show this help"
	@echo ""
	@echo "Examples:"
	@echo "  make               # Start development (default)"
	@echo "  make dev           # Start development server"
	@echo "  make build         # Test build process"
	@echo "  make lint-fix      # Fix linting issues"
