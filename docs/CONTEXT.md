# SSH Vault Keeper - Development Context

## Current User Environment
- **User**: rzago
- **SSH Directory**: Complex structure with 25+ files
- **Key Types**: RSA, PEM, OpenSSH format
- **Services**: GitHub, GitLab, BitBucket, ArgoCD, Quay, GKE
- **Vault**: HashiCorp Vault server (K8s cluster)

## SSH File Analysis

### Identified File Patterns from User Directory:

**Service Keys:**
- `github_rsa` + `github_rsa.pub` - GitHub access
- `gitlab_rsa` + `gitlab_rsa.pub` - GitLab access  
- `gitlab_redhat_rsa` + `gitlab_redhat_rsa.pub` - GitLab Red Hat instance
- `bitbucket_rsa` + `bitbucket_rsa.pub` - BitBucket access
- `argocd_rsa` + `argocd_rsa.pub` - ArgoCD access
- `quay_installer` + `quay_installer.pub` - Quay installer

**Cloud Keys:**
- `gke_redhat_rsa` + `gke_redhat_rsa.pub` - Google Kubernetes Engine

**Personal/Default Keys:**
- `id_rsa` + `id_rsa.pub` - Default SSH key pair
- `local_rsa` + `local_rsa.pub` - Local access keys

**Work/Corporate Keys:**
- `rht_classroom.rsa` - Red Hat classroom environment
- `id_rsa_grade` - Grading system access

**Certificates/PEM Files:**
- `rzago-psi.pem` - Personal certificate
- `cci` + `cci.pem` + `cci.pub` - CI/CD related certificates

**System Files:**
- `config` - SSH client configuration
- `known_hosts` - Known host keys
- `known_hosts.old` - Backup of known hosts
- `authorized_keys` - Authorized public keys

### Permission Patterns Observed:
- Private keys: 600 (-rw-------)
- Public keys: 644 (-rw-r--r--)  
- Config files: 600 (-rw-------)
- System files: 600-644 depending on type

## Architecture Decisions

### 1. Client-Side Encryption
- **Why**: Never trust vault server with plaintext SSH keys
- **Algorithm**: AES-256-GCM with PBKDF2 key derivation
- **Implementation**: Each file encrypted separately for granular access

### 2. Categorized Storage  
- **Service keys**: Grouped by service (GitHub, GitLab, etc.)
- **Personal keys**: User's personal/default keys
- **Work keys**: Corporate/work-related keys
- **System files**: SSH config, known_hosts, etc.

### 3. Flexible Key Detection
- **Pattern-based**: Detects keys by filename patterns and content
- **Service mapping**: Maps key names to services automatically
- **Extensible**: Easy to add new key types and patterns

### 4. Vault Schema
```
ssh-backups/data/users/{hostname-username}/
├── backups/
│   ├── backup-20250112-143022/
│   └── backup-20250113-091505/
└── metadata/
    └── backup_info
```

## Current Implementation Status

### Completed Components:
- ✅ Configuration management (YAML + environment variables)
- ✅ SSH file analysis and categorization  
- ✅ Client-side encryption (AES-256-GCM)
- ✅ Vault client integration
- ✅ CLI commands (init, backup, restore, analyze, list, status)
- ✅ Key pair detection and relationship mapping

### Core Features:
- ✅ Intelligent SSH directory analysis
- ✅ Service-based key categorization  
- ✅ Complete backup/restore workflow
- ✅ Interactive file selection
- ✅ Dry-run support
- ✅ Integrity verification
- ✅ Flexible configuration

### Security Features:
- ✅ Triple-layer encryption (client + vault + TLS)
- ✅ Secure passphrase handling
- ✅ File permission preservation
- ✅ Checksum verification
- ✅ Memory cleanup for sensitive data

## Next Development Priorities

### Testing & Quality:
- Unit tests for all core components
- Integration tests with test Vault instance
- End-to-end workflow tests
- Security testing (encryption/decryption)

### Documentation:
- User guide with examples
- API documentation
- Security model documentation  
- Troubleshooting guide

### Advanced Features:
- Key rotation support
- Backup scheduling
- Compliance reporting
- Multi-vault support
- Plugin system for custom detectors

## Development Guidelines

### Code Quality:
- Follow Go best practices
- Comprehensive error handling
- Structured logging with zerolog
- Testable interfaces and dependency injection
- Security-first approach

### User Experience:
- Clear, informative output
- Interactive prompts when needed
- Dry-run mode for all operations
- Helpful error messages and recommendations
- Progress indicators for long operations

### Security:
- Never log sensitive data
- Clear secrets from memory after use
- Preserve exact file permissions
- Validate all user inputs
- Use secure defaults
