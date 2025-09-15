# Multi-stage build for SSH Vault Keeper
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
ARG VERSION=1.1.0
ARG BUILD_TIME
ARG GIT_HASH

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-X 'github.com/rzago/ssh-vault-keeper/internal/cmd.Version=${VERSION}' \
              -X 'github.com/rzago/ssh-vault-keeper/internal/cmd.BuildTime=${BUILD_TIME}' \
              -X 'github.com/rzago/ssh-vault-keeper/internal/cmd.GitHash=${GIT_HASH}' \
              -w -s" \
    -o ssh-vault-keeper cmd/main.go

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 appuser && \
    adduser -D -u 1001 -G appuser appuser

# Create directories
RUN mkdir -p /home/appuser/.ssh-vault-keeper && \
    chown -R appuser:appuser /home/appuser

# Copy binary from builder
COPY --from=builder /app/ssh-vault-keeper /usr/local/bin/ssh-vault-keeper

# Set permissions
RUN chmod +x /usr/local/bin/ssh-vault-keeper

# Switch to non-root user
USER appuser

# Set working directory
WORKDIR /home/appuser

# Default configuration
ENV SSH_VAULT_LOGGING_LEVEL=info
ENV SSH_VAULT_LOGGING_FORMAT=json

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ssh-vault-keeper version || exit 1

# Entry point
ENTRYPOINT ["ssh-vault-keeper"]

# Default command
CMD ["--help"]

# Labels
LABEL org.opencontainers.image.title="SSH Vault Keeper"
LABEL org.opencontainers.image.description="Securely backup SSH keys to HashiCorp Vault"
LABEL org.opencontainers.image.vendor="SSH Vault Keeper"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.source="https://github.com/rzago/ssh-vault-keeper"
