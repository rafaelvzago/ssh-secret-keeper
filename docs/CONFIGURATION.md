# SSH Secret Keeper Configuration Guide

This guide covers all configuration options for SSH Secret Keeper.

## Configuration Sources

SSH Secret Keeper loads configuration from multiple sources in this order of precedence:

1. **Command line flags** (highest priority)
2. **Environment variables**
3. **Configuration file**
4. **Default values** (lowest priority)

## Configuration File

Default location: `~/.sshsk/config.yaml`

### Complete Configuration Example

```yaml
# SSH Secret Keeper Configuration
version: "1.0"

# Vault connection settings
vault:
  # Vault server address - CHANGE THIS TO YOUR VAULT SERVER
  address: "https://vault.company.com:8200"

  # Path to Vault token file
  token_file: "~/.sshsk/token"

  # KV v2 mount path in Vault for SSH backups
  mount_path: "ssh-backups"

  # Vault namespace (Enterprise only)
  namespace: ""

  # Skip TLS verification (not recommended for production)
  tls_skip_verify: false

  # NEW: Storage strategy configuration
  storage_strategy: "universal"    # Options: "universal", "user", "machine-user", "custom"
  backup_namespace: ""             # Optional namespace for universal strategy
  custom_prefix: ""                # Required for custom strategy

# Backup behavior settings
backup:
  # SSH directory to backup
  ssh_dir: "~/.ssh"

  # Include hostname in Vault storage path (legacy, disabled by default with universal storage)
  hostname_prefix: false

  # Number of backup versions to retain
  retention_count: 10

  # NEW: Path normalization and cross-machine compatibility
  normalize_paths: true            # Enable cross-user compatibility
  cross_machine_restore: true      # Enable cross-machine restore

  # Files to include in backup (glob patterns)
  include_patterns:
    - "*.rsa"
    - "*.pem"
    - "*.pub"
    - "id_rsa*"
    - "config"
    - "known_hosts*"
    - "authorized_keys"
    - "*_rsa"
    - "*_rsa.pub"
    - "*.ed25519"
    - "*.ecdsa"

  # Files to exclude from backup (glob patterns)
  exclude_patterns:
    - "*.tmp"
    - "*.bak"
    - "*.old"
    - "*.orig"
    - "*.swp"

  # Predefined categories for SSH files
  categories:
    services:
      - "service1_rsa"
      - "service2_rsa"
      - "service3_rsa"
      - "service4_rsa"
      - "service5_rsa"
    personal:
      - "id_rsa"
      - "local_rsa"
    work:
      - "*work*"
      - "*corp*"
      - "*office*"
    cloud:
      - "*cloud*"
      - "*aws*"
      - "*gcp*"
      - "*azure*"

# Security settings
security:
  # Encryption algorithm (currently only AES-256-GCM supported)
  algorithm: "AES-256-GCM"

  # Key derivation function (currently only PBKDF2 supported)
  key_derivation: "PBKDF2"

  # PBKDF2 iterations (higher = more secure but slower)
  iterations: 100000

  # Encrypt each file separately (recommended)
  per_file_encrypt: true

  # Verify file integrity after decryption
  verify_integrity: true

# Logging configuration
logging:
  # Log level: debug, info, warn, error
  level: "info"

  # Log format: console, json
  format: "console"

# Key detection settings
detectors:
  # Enabled detector types
  enabled:
    - "rsa"
    - "pem"
    - "openssh"
    - "config"
    - "hosts"
    - "ed25519"
    - "ecdsa"

  # Custom patterns file (optional)
  custom_patterns_file: ""

  # Service to category mapping
  service_mapping:
    github: "git_hosting"
    gitlab: "git_hosting"
    bitbucket: "git_hosting"
    argocd: "automation"
    jenkins: "automation"
    aws: "cloud"
    gcp: "cloud"
    azure: "cloud"
    quay: "container_registry"

  # Purpose assignment rules (glob patterns)
  purpose_rules:
    "*work*": "work"
    "*corp*": "work"
    "*office*": "work"
    "*personal*": "personal"
    "*github*": "service"
    "*gitlab*": "service"
    "*aws*": "cloud"
    "*gcp*": "cloud"
```

## Storage Strategies (NEW)

SSH Secret Keeper supports multiple storage strategies for different use cases:

### Universal Storage (Default)
**Path**: `shared/{namespace}/backups/{backup-name}`

```yaml
vault:
  storage_strategy: "universal"
  backup_namespace: "personal"  # Optional namespace
```

**Benefits**: Cross-machine restore, cross-user restore, team sharing, container friendly

### User-Scoped Storage
**Path**: `users/{username}/backups/{backup-name}`

```yaml
vault:
  storage_strategy: "user"
```

**Benefits**: Cross-machine restore for same user, user isolation in shared Vault

### Machine-User Storage (Legacy)
**Path**: `users/{hostname-username}/backups/{backup-name}`

```yaml
vault:
  storage_strategy: "machine-user"
backup:
  hostname_prefix: true  # Optional with this strategy
```

**Benefits**: Maximum isolation per machine-user (no cross-machine restore)

### Custom Storage
**Path**: `{custom-prefix}/backups/{backup-name}`

```yaml
vault:
  storage_strategy: "custom"
  custom_prefix: "team-devops"  # Required
```

**Benefits**: Team/project organization, flexible prefix configuration

## Migration Between Strategies

### Check Current Strategy
```bash
sshsk migrate-status
```

### Migrate to Universal Storage
```bash
# Preview migration
sshsk migrate --from machine-user --to universal --dry-run

# Perform migration with cleanup
sshsk migrate --from machine-user --to universal --cleanup
```

## Backup Modes and Options

### Command Line Options

#### Basic Backup Modes
```bash
# Quick backup with auto-generated name
sshsk backup

# Named backup for organization
sshsk backup "my-backup-name"

# Interactive file selection
sshsk backup --interactive "selective-backup"

# Preview mode (dry run)
sshsk backup --dry-run
sshsk backup --dry-run "test-backup"

# Custom SSH directory
sshsk backup --ssh-dir "/custom/ssh/path"

# Combined options
sshsk backup --interactive --dry-run "preview-selective"
```

#### Restore Options
```bash
# Basic restore
sshsk restore
sshsk restore "backup-name"

# Cross-machine restore (works with universal storage)
sshsk restore "laptop-backup" --target-dir "~/.ssh"

# File filtering
sshsk restore --files "github*,gitlab*"
sshsk restore --interactive

# Safety options
sshsk restore --dry-run
sshsk restore --overwrite
sshsk restore --select
```

## Environment Variables

All configuration options can be overridden with environment variables using the prefix `SSH_VAULT_` and replacing dots with underscores:

### Vault Configuration
```bash
export SSH_VAULT_ADDRESS="https://vault.company.com:8200"
export SSH_VAULT_TOKEN_FILE="/path/to/token"
export SSH_VAULT_MOUNT_PATH="ssh-backups"
export SSH_VAULT_STORAGE_STRATEGY="universal"
export SSH_VAULT_BACKUP_NAMESPACE="team-devops"
export SSH_VAULT_CUSTOM_PREFIX="my-team"
```

### Backup Configuration
```bash
export SSH_VAULT_BACKUP_SSH_DIR="/custom/ssh/path"
export SSH_VAULT_BACKUP_RETENTION_COUNT="20"
export SSH_VAULT_BACKUP_NORMALIZE_PATHS="true"
export SSH_VAULT_BACKUP_CROSS_MACHINE_RESTORE="true"
export SSH_VAULT_BACKUP_HOSTNAME_PREFIX="false"
```

### Security Configuration
```bash
export SSH_VAULT_SECURITY_ALGORITHM="AES-256-GCM"
export SSH_VAULT_SECURITY_ITERATIONS="100000"
export SSH_VAULT_SECURITY_PER_FILE_ENCRYPT="true"
export SSH_VAULT_SECURITY_VERIFY_INTEGRITY="true"
```

### Logging Configuration
```bash
export SSH_VAULT_LOGGING_LEVEL="debug"
export SSH_VAULT_LOGGING_FORMAT="json"
```

### Vault Configuration
```bash
# Vault server address (standard HashiCorp Vault environment variable)
export VAULT_ADDR="https://vault.company.com:8200"

# Alternative: SSH Secret Keeper specific environment variable
export SSH_SECRET_VAULT_ADDRESS="https://vault.company.com:8200"

# Token file path
export SSH_SECRET_VAULT_TOKEN_FILE="/path/to/vault/token"

# Mount path in Vault
export SSH_SECRET_VAULT_MOUNT_PATH="ssh-backups"

# Vault namespace (Enterprise)
export SSH_SECRET_VAULT_NAMESPACE="team-namespace"

# Skip TLS verification (not recommended)
export SSH_SECRET_VAULT_TLS_SKIP_VERIFY="false"
```

**Note**: The `VAULT_ADDR` environment variable is **required** and takes precedence over the configuration file and `SSH_SECRET_VAULT_ADDRESS` environment variable, following HashiCorp Vault's standard convention. The application will fail to start if `VAULT_ADDR` is not set.

### Authentication Priority

The application uses the following priority for Vault authentication:

1. **VAULT_TOKEN environment variable** (highest priority)
2. **Token file** (as specified in configuration or SSH_SECRET_VAULT_TOKEN_FILE)
3. **Error** (if neither is available)

For enhanced security, especially in containerized or CI/CD environments, using `VAULT_TOKEN` environment variable is recommended over token files.

### Backup Configuration
```bash
# SSH directory to backup
export SSH_SECRET_BACKUP_SSH_DIR="/custom/ssh/path"

# Include hostname in path
export SSH_SECRET_BACKUP_HOSTNAME_PREFIX="true"

# Number of backups to keep
export SSH_SECRET_BACKUP_RETENTION_COUNT="20"
```

### Security Configuration
```bash
# PBKDF2 iterations
export SSH_SECRET_SECURITY_ITERATIONS="150000"

# Per-file encryption
export SSH_SECRET_SECURITY_PER_FILE_ENCRYPT="true"

# Integrity verification
export SSH_SECRET_SECURITY_VERIFY_INTEGRITY="true"
```

### Logging Configuration
```bash
# Log level
export SSH_SECRET_LOGGING_LEVEL="debug"

# Log format
export SSH_SECRET_LOGGING_FORMAT="json"
```

## Command Line Flags

Global flags available for all commands:

```bash
# Configuration file path
--config /path/to/config.yaml

# Enable verbose logging
--verbose

# Suppress all output except errors
--quiet
```

## Vault-Specific Configuration

### Vault Server Examples

#### Local Development
```yaml
vault:
  address: "http://localhost:8200"
  token_file: "~/.vault-token"
  mount_path: "ssh-backups"
  tls_skip_verify: true  # OK for dev only
```

#### Production with HTTPS
```yaml
vault:
  address: "https://vault.company.com:8200"
  token_file: "~/.sshsk/token"
  mount_path: "ssh-backups"
  namespace: "team/ssh"  # Enterprise feature
  tls_skip_verify: false
```

#### High Availability Cluster
```yaml
vault:
  address: "https://vault-cluster.company.com:8200"
  token_file: "/etc/sshsk/token"
  mount_path: "ssh-backups"
```

### Token Management

#### Development Token
```bash
# Create a dev token (short-lived)
vault write auth/token/create policies=sshsk ttl=24h

# Save token
echo "hvs.ABCD..." > ~/.sshsk/token
chmod 600 ~/.sshsk/token
```

#### Production Token
```bash
# Create a renewable token
vault write auth/token/create \
  policies=sshsk \
  ttl=30d \
  renewable=true \
  explicit_max_ttl=90d

# Save token with proper permissions
echo "hvs.PROD..." > ~/.sshsk/token
chmod 600 ~/.sshsk/token
chown $(whoami):$(whoami) ~/.sshsk/token
```

## SSH Directory Configuration

### Custom SSH Directory
```yaml
backup:
  ssh_dir: "/custom/path/to/ssh"
```

### Include/Exclude Patterns

#### Common Patterns
```yaml
backup:
  include_patterns:
    # Standard SSH keys
    - "*.rsa"
    - "*.pem"
    - "*.pub"
    - "id_rsa*"

    # Ed25519 keys
    - "*ed25519*"
    - "*.ed25519"

    # ECDSA keys
    - "*ecdsa*"
    - "*.ecdsa"

    # System files
    - "config"
    - "known_hosts*"
    - "authorized_keys"

    # Service-specific patterns
    - "*github*"
    - "*gitlab*"
    - "*aws*"

  exclude_patterns:
    # Temporary files
    - "*.tmp"
    - "*.temp"
    - "*.swp"
    - "*~"

    # Backup files
    - "*.bak"
    - "*.old"
    - "*.orig"

    # Editor files
    - ".*.swp"
    - "*#"
```

#### Team-Specific Patterns
```yaml
backup:
  include_patterns:
    # Company-specific
    - "*company*"
    - "*corp*"

    # Project-specific
    - "*project1*"
    - "*staging*"
    - "*prod*"

  exclude_patterns:
    # Test files
    - "*test*"
    - "*demo*"
```

## Security Configuration

### Encryption Strength
```yaml
security:
  # Strong configuration for sensitive environments
  iterations: 200000  # Higher iterations = more secure
  per_file_encrypt: true
  verify_integrity: true
```

### Performance vs Security Trade-offs
```yaml
security:
  # Balanced configuration for most use cases
  iterations: 100000  # Good security, reasonable performance

  # Fast configuration for development
  iterations: 50000   # Faster but less secure
```

## Logging Configuration

### Development Logging
```yaml
logging:
  level: "debug"
  format: "console"
```

### Production Logging
```yaml
logging:
  level: "info"
  format: "json"  # Better for log aggregation
```

### Minimal Logging
```yaml
logging:
  level: "warn"
  format: "console"
```

## Service Detection Configuration

### Custom Service Patterns
```yaml
detectors:
  service_mapping:
    # Git hosting
    github: "git_hosting"
    gitlab: "git_hosting"
    bitbucket: "git_hosting"

    # Cloud providers
    aws: "cloud"
    gcp: "cloud"
    azure: "cloud"

    # CI/CD
    jenkins: "automation"
    argocd: "automation"
    drone: "automation"

    # Container registries
    quay: "container_registry"
    docker: "container_registry"

    # Custom services
    myservice: "custom"

  purpose_rules:
    # Work patterns
    "*company*": "work"
    "*corp*": "work"
    "*enterprise*": "work"

    # Personal patterns
    "*personal*": "personal"
    "*home*": "personal"

    # Project patterns
    "*project*": "work"
    "*staging*": "work"
    "*prod*": "work"
```

## Troubleshooting Configuration

### Debug Configuration Loading
```bash
# Show effective configuration
sshsk status --verbose

# Test with custom config
sshsk --config /path/to/config.yaml status

# Override with environment
SSH_SECRET_LOGGING_LEVEL=debug sshsk status
```

### Common Issues

#### Vault Connection
```yaml
# If using self-signed certificates
vault:
  tls_skip_verify: true  # Not recommended for production

# Or provide custom CA
vault:
  ca_cert: "/path/to/ca.pem"
```

#### Permission Issues
```bash
# Fix token file permissions
chmod 600 ~/.sshsk/token

# Fix SSH directory permissions
chmod 700 ~/.ssh
chmod 600 ~/.ssh/id_rsa ~/.ssh/config
chmod 644 ~/.ssh/*.pub
```

#### Path Issues
```yaml
# Use absolute paths to avoid issues
vault:
  token_file: "/home/user/.sshsk/token"
backup:
  ssh_dir: "/home/user/.ssh"
```

This configuration guide should help you customize SSH Secret Keeper for your specific environment and security requirements.
