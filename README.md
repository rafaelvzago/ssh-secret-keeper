# SSH Secret Keeper

[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](https://github.com/rzago/ssh-secret-keeper)

A secure, intelligent tool for backing up SSH keys and configuration to HashiCorp Vault with client-side encryption and **perfect permission preservation**.

## Security First

- **Triple-layer encryption**: Client-side AES-256-GCM + Vault encryption + TLS
- **Zero-knowledge**: Vault server never sees your SSH keys in plaintext
- **Strong key derivation**: PBKDF2 with 100,000 iterations
- **Integrity verification**: MD5 checksums for all files
- **Perfect permission preservation**: Exact SSH file permissions maintained and verified
- **Permission validation**: Critical warnings for insecure SSH key permissions
- **Directory security**: SSH directory automatically secured to 0700 permissions

## Intelligent SSH Analysis

- **Automatic detection**: RSA, PEM, Ed25519, ECDSA, OpenSSH formats
- **Service categorization**: GitHub, GitLab, BitBucket, ArgoCD, AWS, GCP, etc.
- **Key pair relationships**: Links private/public key pairs automatically
- **System file recognition**: Config, known_hosts, authorized_keys
- **Smart categorization**: Personal, work, service, and system files

## Features

- Complete SSH Directory Backup with metadata
- Selective File Restore with permission verification
- Interactive Mode for file and backup selection
- Dry-run Support for safe testing
- Multiple Backup Versions with retention
- Cross-platform Support (Linux, macOS, Windows)
- Container Ready with Docker and Podman support
- CI/CD Integration friendly
- Permission Security validation and warnings

## Installation

### Option 1: Download Release Binary
```bash
# Download latest release (replace VERSION and ARCH)
curl -L https://github.com/rzago/ssh-secret-keeper/releases/latest/download/sshsk-linux-amd64 -o sshsk
chmod +x sshsk
sudo mv sshsk /usr/local/bin/
```

### Option 2: Build from Source
```bash
git clone https://github.com/rzago/sshsk
cd sshsk
make build
sudo make install
```

### Option 3: Container (Docker/Podman)
```bash
# Using Docker
docker pull ghcr.io/rzago/ssh-secret-keeper:latest
docker run --rm -v ~/.ssh:/ssh -v ~/.ssh-secret-keeper:/config ghcr.io/rzago/ssh-secret-keeper analyze

# Using Podman
podman pull ghcr.io/rzago/ssh-secret-keeper:latest
podman run --rm -v ~/.ssh:/ssh -v ~/.ssh-secret-keeper:/config ghcr.io/rzago/ssh-secret-keeper analyze
```

## Prerequisites

- HashiCorp Vault server (local or remote)
- Valid Vault token with KV v2 permissions
- SSH directory with keys to backup (~/.ssh)

## Quick Start

### Authentication Methods

SSH Secret Keeper supports two authentication approaches:

#### Option 1: Environment Variables Only (Recommended)
Perfect for containers, CI/CD, and environments where no local files should be stored.

```bash
# Set required environment variables
export VAULT_ADDR="https://your-vault-server:8200"  # Replace with your actual Vault server address
export VAULT_TOKEN="your-vault-token"              # Your Vault authentication token

# Initialize (no config files created)
sshsk init

# Use any command - everything works with environment variables
sshsk backup my-backup
sshsk status
```

#### Option 2: Configuration Files (Traditional)
Uses local configuration and token files for persistent setups.

```bash
# Set Vault address (required)
export VAULT_ADDR="https://your-vault-server:8200"

# Initialize with token flag (creates config and token files)
sshsk init --token "your-vault-token"
```

**Important Notes:**
- `VAULT_ADDR` environment variable is **required** for all operations
- `VAULT_TOKEN` environment variable takes priority over token files
- When using environment variables, no local files are created or required
- The application will fail with clear error messages if authentication is missing

### 1. Initialize Configuration

### 2. Analyze Your SSH Directory
```bash
# See what SSH files you have
sshsk analyze --verbose
```

Example output:
```
SSH Directory Analysis
======================

Summary:
  Total files: 28
  Key pairs: 14
  Service keys: 14 (GitHub, GitLab, ArgoCD, etc.)
  Personal keys: 10
  Work keys: 1
  System files: 3

Key Pairs Found:
  - service1_rsa (Complete pair) [0600/0644]
  - service2_rsa (Complete pair) [0600/0644]
  - service3_rsa (Complete pair) [0600/0644]
  - id_rsa (Complete pair) [0600/0644]

Permission Summary:
  - 0600: 14 files (private keys)
  - 0644: 14 files (public keys, config)
```

### 3. Create Your First Backup
```bash
# Backup everything (you'll be prompted for encryption passphrase)
sshsk backup

# Or with a custom name using variables
BACKUP_NAME="laptop-$(hostname)-$(date +%Y%m%d)"
sshsk backup "${BACKUP_NAME}"
```

### 4. Restore on Another Machine
```bash
# List available backups
sshsk list --detailed

# Restore the most recent backup
sshsk restore

# Or restore to a specific directory using variables
RESTORE_DIR="/tmp/restored-ssh-$(date +%Y%m%d)"
sshsk restore --target-dir "${RESTORE_DIR}"
```

## Commands Reference

### Core Commands

| Command | Description | Example |
|---------|-------------|---------|
| `init` | Initialize configuration and Vault setup | `sshsk init --vault-addr "${VAULT_ADDR}"` |
| `backup` | Backup SSH directory to Vault | `sshsk backup "${BACKUP_NAME}"` |
| `restore` | Restore SSH backup from Vault | `sshsk restore --select` |
| `list` | List available backups | `sshsk list --detailed` |
| `delete` | Delete a backup from Vault | `sshsk delete "${BACKUP_NAME}" --force` |
| `analyze` | Analyze SSH directory structure | `sshsk analyze --verbose` |
| `status` | Show configuration and connection status | `sshsk status --checksums` |

### Command Options

#### Backup Options
```bash
# Interactive file selection
sshsk backup --interactive

# Dry run (preview only)
sshsk backup --dry-run

# Custom SSH directory using variables
SSH_DIR="/path/to/custom/ssh"
sshsk backup --ssh-dir "${SSH_DIR}"

# Custom backup name with timestamp
BACKUP_NAME="backup-$(hostname)-$(date +%Y%m%d-%H%M%S)"
sshsk backup "${BACKUP_NAME}"
```

#### Delete Options
```bash
# Delete specific backup with confirmation
sshsk delete "backup-20240101-120000"

# Delete without confirmation prompt
sshsk delete "old-backup" --force

# Interactive backup selection for deletion
sshsk delete "" --interactive
```

#### Restore Options
```bash
# Interactive backup selection
sshsk restore --select

# Restore specific files only
sshsk restore --files "github*,gitlab*"

# Restore to different location using variables
TARGET_DIR="/tmp/ssh-restore-$(date +%Y%m%d)"
sshsk restore --target-dir "${TARGET_DIR}"

# Interactive file selection
sshsk restore --interactive

# Combine interactive backup and file selection
sshsk restore --select --interactive

# Overwrite existing files
sshsk restore --overwrite
```

#### Status Options
```bash
# Show basic status
sshsk status

# Show MD5 checksums for most recent backup
sshsk status --checksums

# Show detailed info for specific backup with checksums
sshsk status "backup-20240101-120000" --checksums

# Skip vault connection check
sshsk status --vault=false

# Skip SSH directory check
sshsk status --ssh=false
```

## Configuration

Configuration file location: `~/.ssh-secret-keeper/config.yaml`

### Environment Variables

#### Required Environment Variables

```bash
# Vault authentication - REQUIRED for all operations
export VAULT_ADDR="https://vault.company.com:8200"    # Vault server address (REQUIRED)
export VAULT_TOKEN="your-vault-token-here"            # Vault authentication token (REQUIRED)

# For Kubernetes clusters, set your cluster's Vault address:
# export VAULT_ADDR="http://your-vault-server:8200"
```

#### Optional Environment Variables

All other configuration can be overridden with environment variables:

```bash
# Vault settings (optional overrides)
export SSH_SECRET_VAULT_ADDRESS="https://vault.company.com:8200"  # Alternative to VAULT_ADDR
export SSH_SECRET_VAULT_TOKEN_FILE="/path/to/token"               # Token file path (if not using VAULT_TOKEN)
export SSH_SECRET_VAULT_MOUNT_PATH="ssh-backups"                  # Vault mount path
export SSH_SECRET_VAULT_NAMESPACE="your-namespace"                # Vault namespace (Enterprise)
export SSH_SECRET_VAULT_TLS_SKIP_VERIFY="false"                   # Skip TLS verification (not recommended)

# Backup settings
export SSH_SECRET_BACKUP_SSH_DIR="/custom/ssh/path"               # Custom SSH directory
export SSH_SECRET_BACKUP_RETENTION_COUNT="20"                     # Number of backups to keep
export SSH_SECRET_BACKUP_HOSTNAME_PREFIX="true"                   # Include hostname in paths

# Security settings
export SSH_SECRET_SECURITY_ALGORITHM="AES-256-GCM"                # Encryption algorithm
export SSH_SECRET_SECURITY_ITERATIONS="150000"                    # PBKDF2 iterations
export SSH_SECRET_SECURITY_PER_FILE_ENCRYPT="true"                # Encrypt each file separately
export SSH_SECRET_SECURITY_VERIFY_INTEGRITY="true"                # Verify checksums

# Logging settings
export SSH_SECRET_LOGGING_LEVEL="info"                            # Log level: debug, info, warn, error
export SSH_SECRET_LOGGING_FORMAT="console"                        # Log format: console, json
```

#### Authentication Priority

The application uses the following priority for authentication:

1. **VAULT_TOKEN environment variable** (highest priority)
2. **Token file** (fallback if VAULT_TOKEN not set)
3. **Error** (if neither is available)

#### Environment-Only Mode

When both `VAULT_ADDR` and `VAULT_TOKEN` are set as environment variables:
- ✅ No configuration files are created
- ✅ No token files are stored locally
- ✅ Perfect for containers and CI/CD environments
- ✅ Enhanced security (no secrets on disk)

### Sample Configuration

```yaml
version: "1.0"

vault:
  address: "https://vault.company.com:8200"
  token_file: "~/.sshsk/token"
  mount_path: "ssh-backups"
  tls_skip_verify: false

backup:
  ssh_dir: "~/.ssh"
  hostname_prefix: true
  retention_count: 10
  include_patterns:
    - "*.rsa"
    - "*.pem"
    - "*.pub"
    - "id_rsa*"
    - "config"
    - "known_hosts*"
    - "authorized_keys"
  exclude_patterns:
    - "*.tmp"
    - "*.bak"

security:
  algorithm: "AES-256-GCM"
  key_derivation: "PBKDF2"
  iterations: 100000
  per_file_encrypt: true
  verify_integrity: true
```

## Vault Setup

### Required Vault Policy

```hcl
# SSH Secret Keeper policy
path "ssh-backups/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

path "sys/mounts" {
  capabilities = ["read"]
}

path "sys/mounts/ssh-backups" {
  capabilities = ["create", "update"]
}
```

### Create Vault Token

```bash
# Set token TTL (time to live)
TOKEN_TTL="8760h"  # 1 year

# Create a token with the policy
vault write auth/token/create policies=sshsk ttl="${TOKEN_TTL}"
```

## Enterprise Features

### Team Deployment
- User isolation: Each user gets their own namespace
- Policy-based access: Integrate with Vault policies
- Audit logging: All operations logged in Vault
- Compliance ready: SOC2, PCI DSS compatible

### CI/CD Integration

#### GitLab CI/CD
```yaml
# .gitlab-ci.yml
variables:
  VAULT_ADDR: "https://vault.company.com:8200"

backup_ssh_keys:
  image: ghcr.io/rzago/sshsk:latest
  variables:
    BACKUP_NAME: "ci-${CI_COMMIT_SHA}-${CI_PIPELINE_ID}"
  before_script:
    # VAULT_TOKEN should be set as a CI/CD variable (masked)
    - echo "Using Vault at ${VAULT_ADDR}"
  script:
    - sshsk init  # No files created, uses env vars only
    - sshsk backup "${BACKUP_NAME}"
  only:
    - main

restore_ssh_keys:
  image: ghcr.io/rzago/sshsk:latest
  script:
    - sshsk init
    - sshsk restore --target-dir /tmp/ssh-keys
    - ls -la /tmp/ssh-keys/
  when: manual
```

#### GitHub Actions
```yaml
# .github/workflows/ssh-backup.yml
name: SSH Backup

on:
  push:
    branches: [ main ]

jobs:
  backup:
    runs-on: ubuntu-latest
    steps:
    - name: Backup SSH Keys
      env:
        VAULT_ADDR: ${{ vars.VAULT_ADDR }}
        VAULT_TOKEN: ${{ secrets.VAULT_TOKEN }}
      run: |
        curl -L https://github.com/rzago/sshsk/releases/latest/download/sshsk-linux-amd64 -o sshsk
        chmod +x sshsk
        ./sshsk init
        ./sshsk backup "github-${GITHUB_SHA}-${GITHUB_RUN_ID}"

  restore:
    runs-on: ubuntu-latest
    if: github.event_name == 'workflow_dispatch'
    steps:
    - name: Restore SSH Keys
      env:
        VAULT_ADDR: ${{ vars.VAULT_ADDR }}
        VAULT_TOKEN: ${{ secrets.VAULT_TOKEN }}
      run: |
        curl -L https://github.com/rzago/sshsk/releases/latest/download/sshsk-linux-amd64 -o sshsk
        chmod +x sshsk
        ./sshsk init
        ./sshsk restore --target-dir /tmp/ssh-keys
```

#### Jenkins Pipeline
```groovy
pipeline {
    agent any

    environment {
        VAULT_ADDR = 'https://vault.company.com:8200'
        VAULT_TOKEN = credentials('vault-token')
    }

    stages {
        stage('Backup SSH Keys') {
            steps {
                sh '''
                    curl -L https://github.com/rzago/sshsk/releases/latest/download/sshsk-linux-amd64 -o sshsk
                    chmod +x sshsk
                    ./sshsk init
                    ./sshsk backup "jenkins-${BUILD_NUMBER}-${GIT_COMMIT}"
                '''
            }
        }
    }
}
```

### Automation

#### Cron Jobs with Environment Variables
```bash
# /etc/cron.d/ssh-backup
# Daily backup at 2 AM using environment variables
0 2 * * * root /usr/bin/env VAULT_ADDR="https://vault.company.com:8200" VAULT_TOKEN="$(cat /etc/vault/token)" sshsk backup "daily-$(date +\%Y\%m\%d)" >> /var/log/ssh-backup.log 2>&1

# Weekly cleanup - keep only 30 days
0 3 * * 0 root /usr/bin/env VAULT_ADDR="https://vault.company.com:8200" VAULT_TOKEN="$(cat /etc/vault/token)" sshsk list | grep -E "daily-[0-9]{8}" | tail -n +30 | xargs -I {} sshsk delete {} --force
```

#### Machine Setup Script
```bash
#!/bin/bash
# setup-new-machine.sh - Automated SSH key restoration for new machines
set -e

# Configuration variables (set these in your environment or CI/CD)
VAULT_ADDR="${VAULT_ADDR:-https://vault.company.com:8200}"
VAULT_TOKEN="${VAULT_TOKEN}"
SSH_DIR="${HOME}/.ssh"
BACKUP_NAME="${BACKUP_NAME:-latest}"

# Validation
if [ -z "$VAULT_ADDR" ]; then
    echo "Error: VAULT_ADDR environment variable is required"
    exit 1
fi

if [ -z "$VAULT_TOKEN" ]; then
    echo "Error: VAULT_TOKEN environment variable is required"
    exit 1
fi

echo "Setting up SSH keys from Vault..."

# Download sshsk if not available
if ! command -v sshsk &> /dev/null; then
    echo "Downloading sshsk..."
    curl -L "https://github.com/rzago/sshsk/releases/latest/download/sshsk-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m | sed 's/x86_64/amd64/')" -o /tmp/sshsk
    chmod +x /tmp/sshsk
    SSH_SECRET_KEEPER="/tmp/sshsk"
else
    SSH_SECRET_KEEPER="sshsk"
fi

# Initialize (no config files created)
echo "Initializing with environment variables..."
$SSH_SECRET_KEEPER init

# Restore SSH keys
echo "Restoring SSH backup: $BACKUP_NAME"
$SSH_SECRET_KEEPER restore "$BACKUP_NAME" --target-dir "$SSH_DIR"

# Secure SSH directory
chmod 700 "$SSH_DIR"

# Add keys to SSH agent if available
if command -v ssh-add &> /dev/null && [ -n "$SSH_AUTH_SOCK" ]; then
    echo "Adding keys to SSH agent..."
    find "$SSH_DIR" -name "id_*" -not -name "*.pub" -exec ssh-add {} \; 2>/dev/null || true
fi

echo "✅ SSH keys restored successfully!"
echo "SSH directory: $SSH_DIR"
$SSH_SECRET_KEEPER status --ssh-only

# Cleanup temporary binary
if [ "$SSH_SECRET_KEEPER" = "/tmp/sshsk" ]; then
    rm -f /tmp/sshsk
fi
```

#### Docker Compose for Scheduled Backups
```yaml
# docker-compose.yml
version: '3.8'

services:
  ssh-backup:
    image: ghcr.io/rzago/sshsk:latest
    environment:
      - VAULT_ADDR=${VAULT_ADDR}
      - VAULT_TOKEN=${VAULT_TOKEN}
    volumes:
      - ~/.ssh:/ssh:ro
    command: >
      sh -c "
        sshsk init &&
        sshsk backup 'scheduled-$(date +%Y%m%d-%H%M%S)'
      "
    profiles:
      - backup

  ssh-restore:
    image: ghcr.io/rzago/sshsk:latest
    environment:
      - VAULT_ADDR=${VAULT_ADDR}
      - VAULT_TOKEN=${VAULT_TOKEN}
    volumes:
      - ./restored-ssh:/ssh
    command: >
      sh -c "
        sshsk init &&
        sshsk restore --target-dir /ssh
      "
    profiles:
      - restore
```

Usage:
```bash
# Set environment variables
export VAULT_ADDR="https://vault.company.com:8200"
export VAULT_TOKEN="your-vault-token"

# Run backup
docker-compose --profile backup up ssh-backup

# Run restore
docker-compose --profile restore up ssh-restore
```

## Container Usage (Docker/Podman)

### Environment Variables Only (Recommended)

Perfect for containers - no configuration files needed:

```bash
# Configuration variables
IMAGE_NAME="sshsk"
IMAGE_TAG="latest"
REGISTRY="ghcr.io/rzago"

# Set your Vault credentials
export VAULT_ADDR="https://your-vault-server:8200"
export VAULT_TOKEN="your-vault-token"

# Analyze SSH directory with Docker (environment variables only)
docker run --rm \
  -v ~/.ssh:/ssh:ro \
  -e VAULT_ADDR="${VAULT_ADDR}" \
  -e VAULT_TOKEN="${VAULT_TOKEN}" \
  "${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}" analyze

# Analyze SSH directory with Podman (environment variables only)
podman run --rm \
  -v ~/.ssh:/ssh:ro \
  -e VAULT_ADDR="${VAULT_ADDR}" \
  -e VAULT_TOKEN="${VAULT_TOKEN}" \
  "${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}" analyze

# Initialize and backup with Docker (no config files created)
docker run --rm \
  -v ~/.ssh:/ssh:ro \
  -e VAULT_ADDR="${VAULT_ADDR}" \
  -e VAULT_TOKEN="${VAULT_TOKEN}" \
  "${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}" init

docker run --rm \
  -v ~/.ssh:/ssh:ro \
  -e VAULT_ADDR="${VAULT_ADDR}" \
  -e VAULT_TOKEN="${VAULT_TOKEN}" \
  "${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}" backup "container-backup-$(date +%Y%m%d)"

# Initialize and backup with Podman (no config files created)
podman run --rm \
  -v ~/.ssh:/ssh:ro \
  -e VAULT_ADDR="${VAULT_ADDR}" \
  -e VAULT_TOKEN="${VAULT_TOKEN}" \
  "${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}" init

podman run --rm \
  -v ~/.ssh:/ssh:ro \
  -e VAULT_ADDR="${VAULT_ADDR}" \
  -e VAULT_TOKEN="${VAULT_TOKEN}" \
  "${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}" backup "container-backup-$(date +%Y%m%d)"

# Restore SSH keys to a new location
docker run --rm \
  -v /tmp/restored-ssh:/ssh \
  -e VAULT_ADDR="${VAULT_ADDR}" \
  -e VAULT_TOKEN="${VAULT_TOKEN}" \
  "${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}" restore --target-dir /ssh
```

### Traditional Configuration Files (Optional)

If you prefer using configuration files:

```bash
# Build image with Docker
docker build -t "${IMAGE_NAME}:${IMAGE_TAG}" .

# Build image with Podman
podman build -t "${IMAGE_NAME}:${IMAGE_TAG}" .

# Use with configuration files (traditional approach)
docker run --rm \
  -v ~/.ssh:/ssh:ro \
  -v ~/.sshsk:/config \
  -e VAULT_ADDR="${VAULT_ADDR}" \
  "${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}" analyze
```

### Kubernetes Deployment
```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: ssh-backup
spec:
  schedule: "0 2 * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: sshsk
            image: ghcr.io/rzago/sshsk:latest
            command:
            - /bin/sh
            - -c
            - |
              BACKUP_NAME="k8s-daily-$(date +%Y%m%d)"
              sshsk backup "${BACKUP_NAME}"
            env:
            - name: VAULT_ADDR
              value: "http://your-vault-server:8200"  # Your Kubernetes cluster Vault address
            - name: VAULT_TOKEN
              valueFrom:
                secretKeyRef:
                  name: vault-token
                  key: token
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
          restartPolicy: OnFailure
```

## Architecture & Design

### Clean Architecture
This project follows SOLID principles and clean architecture patterns for maintainability and extensibility:

#### SOLID Principles Implementation
- Single Responsibility Principle: Each service has one clear purpose
  - VaultStorageService: Handles only Vault storage operations
  - FileAnalysisService: Handles only SSH file analysis
  - EncryptionService: Handles only encryption/decryption operations
  - ValidationService: Handles only input validation
- Open/Closed Principle: New detectors can be added via registry pattern
- Liskov Substitution Principle: All implementations are substitutable via interfaces
- Interface Segregation Principle: Focused, role-specific interfaces
- Dependency Inversion Principle: All components depend on abstractions, not concretions

#### Service Architecture
```
┌─────────────────┐    ┌─────────────────────┐
│  CLI Commands   │────│ BackupOrchestrator  │
└─────────────────┘    └─────────────────────┘
                                │
                       ┌────────┼────────┐
                       │        │        │
              ┌────────▼──┐  ┌──▼──────┐ │
              │ Vault     │  │ Crypto  │ │
              │ Storage   │  │ Service │ │
              │ Service   │  └─────────┘ │
              └───────────┘              │
                       ┌─────────────────▼┐
                       │ File Services    │
                       │ - Analysis       │
                       │ - Read           │
                       │ - Restore        │
                       │ - Validation     │
                       └──────────────────┘
```

#### Key Design Patterns
- Service Registry Pattern: Pluggable SSH key detectors
- Dependency Injection: Constructor-based service wiring
- Strategy Pattern: Multiple encryption algorithms supported
- Factory Pattern: Service creation with proper configuration
- Repository Pattern: Abstract data access layer for Vault

### Code Quality Standards
- Test Coverage: 85%+ with unit and integration tests
- Error Handling: Structured, contextual error messages
- Logging: Structured logging with zerolog
- Validation: Comprehensive input validation throughout
- Security: Fail-safe defaults, secure by design

## Building

### Prerequisites
- Go 1.21+
- Make

### Build Instructions
```bash
git clone https://github.com/rzago/sshsk
cd sshsk

# Build for current platform
make build

# Build for all platforms
make build-all

# Create release
make release
```

### Testing
```bash
# Run unit tests
make test
```

## Project Statistics

- Language: Go 1.21+
- Lines of Code: ~4,300+
- Test Coverage: 85%+ (comprehensive test suite)
- Dependencies: Minimal, security-focused
- Performance: <100ms for typical SSH directories
- Architecture: Clean architecture with SOLID principles
- Compatibility: Linux, macOS, Windows, ARM64

## Security Model

### Authentication Security

#### Environment Variables (Recommended)
- ✅ **No secrets on disk**: Tokens only in memory
- ✅ **Container-friendly**: Perfect for ephemeral environments
- ✅ **CI/CD secure**: Integrates with secret management systems
- ✅ **Process isolation**: Environment variables are process-scoped
- ⚠️ **Process visibility**: Other processes with same user may see environment variables

#### Token Files (Traditional)
- ✅ **Persistent**: Survives process restarts
- ✅ **File permissions**: Protected with 0600 permissions
- ⚠️ **Disk storage**: Token stored on local filesystem
- ⚠️ **Backup exposure**: May be included in system backups

#### Security Best Practices

```bash
# ✅ Good: Use environment variables in containers
docker run --rm \
  -e VAULT_ADDR="https://vault.company.com:8200" \
  -e VAULT_TOKEN="$(vault write -field=token auth/token/create)" \
  sshsk backup

# ✅ Good: Use CI/CD secret management
# GitLab CI/CD Variables (masked)
# GitHub Actions Secrets
# Jenkins Credentials

# ⚠️ Avoid: Hardcoded tokens in scripts
export VAULT_TOKEN="hvs.hardcoded-token-here"  # Don't do this

# ✅ Good: Read from secure location
export VAULT_TOKEN="$(cat /run/secrets/vault-token)"

# ✅ Good: Use short-lived tokens
vault write auth/token/create ttl=1h policies=sshsk
```

### Data Flow
1. SSH files read from ~/.ssh with exact permissions captured (mode, timestamps)
2. Client-side encryption using AES-256-GCM with unique salt/IV per file
3. Permission metadata stored alongside encrypted content
4. Encrypted data transmitted to Vault over TLS
5. Vault stores encrypted data (double encryption)
6. User namespace isolation: users/{hostname-username}/
7. Permission restoration with validation and security warnings

### Key Management
- User-provided passphrase → PBKDF2(100k iterations) → Encryption key
- Each file encrypted with unique cryptographic parameters
- Vault server never sees plaintext SSH keys
- Token-based Vault authentication with minimal permissions
- Environment variable authentication takes priority over files

### Attack Resistance
- **Vault compromise**: SSH keys remain encrypted with user passphrase
- **Network interception**: TLS encryption protects data in transit
- **Local compromise**: Keys stored encrypted in Vault, no local token files (env var mode)
- **Process inspection**: Environment variables only visible to same user processes
- **Container escape**: No persistent secrets on container filesystem
- **Brute force**: Strong PBKDF2 parameters make attacks infeasible
- **Permission tampering**: Exact file permissions verified during restore
- **Directory security**: SSH directories automatically secured to proper permissions

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes with tests
4. Run tests (`make test`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request


### Extending the System
The architecture supports easy extension:

```go
// Add a new key detector
type CustomKeyDetector struct{}

func (d *CustomKeyDetector) Name() string { return "custom" }
func (d *CustomKeyDetector) Detect(filename string, content []byte) (*KeyInfo, bool) {
    // Implementation
}

// Register with service
analysisService := analyzer.NewService()
analysisService.RegisterDetector(&CustomKeyDetector{})
```

```go
// Add a new storage backend
type S3StorageService struct{}

func (s *S3StorageService) StoreBackup(ctx context.Context, name string, data map[string]interface{}) error {
    // Implementation
}

// Use with orchestrator
orchestrator := orchestrator.New(s3Storage, analysis, fileRead, fileRestore, encryption, validation)
```

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- HashiCorp Vault team for the excellent secrets management platform
- Go community for the robust standard library
- Contributors and early adopters

## Troubleshooting

### Common Issues

#### Permission Problems
If SSH files are restored with incorrect permissions:

```bash
# Check debug logs during restore
sshsk restore --dry-run --verbose

# Look for these messages:
# ✅ "✓ Parsed permissions from json.Number" (working correctly)
# ❌ "❌ CRITICAL: Missing or invalid permissions" (issue detected)
```

#### Vault Connection Issues
```bash
# Test Vault connectivity
sshsk status

# Check environment variables
echo "VAULT_ADDR: $VAULT_ADDR"
echo "VAULT_TOKEN: ${VAULT_TOKEN:0:10}..." # First 10 chars only

# Verify Vault token has correct permissions
vault token lookup
```

#### SSH Directory Issues
```bash
# Analyze SSH directory structure
sshsk analyze --verbose

# Check SSH directory permissions
ls -la ~/.ssh

# SSH directory should be 0700, keys should be 0600, public keys 0644
```

#### Container Issues
```bash
# Verify environment variables are passed correctly
docker run --rm -e VAULT_ADDR -e VAULT_TOKEN sshsk:latest status

# Check volume mounts
docker run --rm -v ~/.ssh:/ssh:ro sshsk:latest analyze
```

### Debug Mode

Enable detailed logging for troubleshooting:

```bash
# Set debug log level
export SSH_SECRET_LOGGING_LEVEL=debug

# Run with verbose output
sshsk [command] --verbose
```

### Getting Help

If you encounter issues:

1. **Check logs**: Run with `--verbose` flag
2. **Verify environment**: Use `sshsk status` command
3. **Test connectivity**: Ensure Vault is accessible
4. **Check permissions**: Verify SSH directory structure
5. **Review documentation**: Check command examples above

## Support

- [Documentation](docs/)
- [Issue Tracker](https://github.com/rzago/sshsk/issues)
- [Discussions](https://github.com/rzago/sshsk/discussions)
- [Security Issues](https://github.com/rzago/sshsk/issues/new?labels=security)

## Recent Updates

### v1.2.0 - Permission Preservation Fix (Latest)
- **🔧 Fixed**: Critical permission parsing bug for Vault data
- **✅ Resolved**: All SSH files now restore with correct original permissions
- **🔍 Enhanced**: Comprehensive debug logging for troubleshooting
- **🛡️ Improved**: Better error handling for data type mismatches

### Previous Features
- Perfect permission preservation with validation
- Intelligent SSH key detection and categorization
- Triple-layer encryption (client-side + Vault + TLS)
- Container and CI/CD ready
- Cross-platform support

## What's Next

- Key rotation automation
- Web UI for team management
- Plugin system for custom key types
- Multi-vault redundancy
- Compliance reporting (SOC2, PCI)
- Integration with cloud HSMs

---

Security Notice: Always test backup and restore operations in a safe environment before using with critical SSH keys. Maintain independent backups of your SSH keys as well.

Secure SSH key management solution