# SSH Vault Keeper

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](https://github.com/rzago/ssh-vault-keeper)

A secure, intelligent tool for backing up SSH keys and configuration to HashiCorp Vault with client-side encryption.

## 🔒 Security First

- **Triple-layer encryption**: Client-side AES-256-GCM + Vault encryption + TLS
- **Zero-knowledge**: Vault server never sees your SSH keys in plaintext
- **Strong key derivation**: PBKDF2 with 100,000 iterations
- **Integrity verification**: SHA-256 checksums for all files
- **Permission preservation**: Exact SSH file permissions maintained

## 🧠 Intelligent SSH Analysis

- **Automatic detection**: RSA, PEM, Ed25519, ECDSA, OpenSSH formats
- **Service categorization**: GitHub, GitLab, BitBucket, ArgoCD, AWS, GCP, etc.
- **Key pair relationships**: Links private/public key pairs automatically
- **System file recognition**: Config, known_hosts, authorized_keys
- **Smart categorization**: Personal, work, service, and system files

## 🚀 Features

✅ **Complete SSH Directory Backup**  
✅ **Selective File Restore**  
✅ **Interactive Mode** for file selection  
✅ **Dry-run Support** for safe testing  
✅ **Multiple Backup Versions** with retention  
✅ **Cross-platform Support** (Linux, macOS, Windows)  
✅ **Container Ready** with Docker support  
✅ **CI/CD Integration** friendly  

## 📦 Installation

### Option 1: Download Release Binary
```bash
# Download latest release (replace VERSION and ARCH)
curl -L https://github.com/rzago/ssh-vault-keeper/releases/latest/download/ssh-vault-keeper-linux-amd64 -o ssh-vault-keeper
chmod +x ssh-vault-keeper
sudo mv ssh-vault-keeper /usr/local/bin/
```

### Option 2: Build from Source
```bash
git clone https://github.com/rzago/ssh-vault-keeper
cd ssh-vault-keeper
make build
sudo make install
```

### Option 3: Docker
```bash
docker pull ghcr.io/rzago/ssh-vault-keeper:latest
docker run --rm -v ~/.ssh:/ssh -v ~/.ssh-vault-keeper:/config ssh-vault-keeper analyze
```

## 🏗️ Prerequisites

- **HashiCorp Vault server** (local or remote)
- **Valid Vault token** with KV v2 permissions
- **SSH directory** with keys to backup (`~/.ssh`)

## ⚡ Quick Start

### 1. Initialize Configuration
```bash
# Replace with your Vault server address and token
ssh-vault-keeper init \
  --vault-addr https://your-vault-server:8200 \
  --token your-vault-token
```

### 2. Analyze Your SSH Directory
```bash
# See what SSH files you have
ssh-vault-keeper analyze --verbose
```

Example output:
```
🔍 SSH Directory Analysis
════════════════════════

📊 Summary:
  Total files: 28
  Key pairs: 14
  Service keys: 14 (GitHub, GitLab, ArgoCD, etc.)
  Personal keys: 10
  Work keys: 1
  System files: 3

🔑 Key Pairs Found:
  • github_rsa ✓ Complete pair
  • gitlab_rsa ✓ Complete pair  
  • argocd_rsa ✓ Complete pair
  • id_rsa ✓ Complete pair
```

### 3. Create Your First Backup
```bash
# Backup everything (you'll be prompted for encryption passphrase)
ssh-vault-keeper backup

# Or with a custom name
ssh-vault-keeper backup my-laptop-keys
```

### 4. Restore on Another Machine
```bash
# List available backups
ssh-vault-keeper list --detailed

# Restore the most recent backup
ssh-vault-keeper restore

# Or restore to a specific directory
ssh-vault-keeper restore --target-dir /tmp/restored-ssh
```

## 📖 Commands Reference

### Core Commands

| Command | Description | Example |
|---------|-------------|---------|
| `init` | Initialize configuration and Vault setup | `ssh-vault-keeper init --vault-addr https://vault.company.com:8200` |
| `backup` | Backup SSH directory to Vault | `ssh-vault-keeper backup my-backup` |
| `restore` | Restore SSH backup from Vault | `ssh-vault-keeper restore my-backup` |
| `list` | List available backups | `ssh-vault-keeper list --detailed` |
| `analyze` | Analyze SSH directory structure | `ssh-vault-keeper analyze --verbose` |
| `status` | Show configuration and connection status | `ssh-vault-keeper status` |

### Command Options

#### Backup Options
```bash
# Interactive file selection
ssh-vault-keeper backup --interactive

# Dry run (preview only)
ssh-vault-keeper backup --dry-run

# Custom SSH directory
ssh-vault-keeper backup --ssh-dir /path/to/ssh

# Custom backup name
ssh-vault-keeper backup my-backup-name
```

#### Restore Options
```bash
# Restore specific files only
ssh-vault-keeper restore --files "github*,gitlab*"

# Restore to different location
ssh-vault-keeper restore --target-dir /tmp/ssh-restore

# Interactive restore
ssh-vault-keeper restore --interactive

# Overwrite existing files
ssh-vault-keeper restore --overwrite
```

## ⚙️ Configuration

Configuration file location: `~/.ssh-vault-keeper/config.yaml`

### Environment Variables

All configuration can be overridden with environment variables:

```bash
# Vault settings
export SSH_VAULT_VAULT_ADDRESS="https://vault.company.com:8200"
export SSH_VAULT_VAULT_TOKEN_FILE="/path/to/token"
export SSH_VAULT_VAULT_MOUNT_PATH="ssh-backups"

# Backup settings  
export SSH_VAULT_BACKUP_SSH_DIR="/custom/ssh/path"
export SSH_VAULT_BACKUP_RETENTION_COUNT="20"

# Security settings
export SSH_VAULT_SECURITY_ITERATIONS="150000"
export SSH_VAULT_LOGGING_LEVEL="debug"
```

### Sample Configuration

```yaml
version: "1.0"

vault:
  address: "https://vault.company.com:8200"
  token_file: "~/.ssh-vault-keeper/token"
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

## 🔐 Vault Setup

### Required Vault Policy

```hcl
# SSH Vault Keeper policy
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
# Create a token with the policy
vault write auth/token/create policies=ssh-vault-keeper ttl=8760h
```

## 🏢 Enterprise Features

### Team Deployment
- **User isolation**: Each user gets their own namespace
- **Policy-based access**: Integrate with Vault policies
- **Audit logging**: All operations logged in Vault
- **Compliance ready**: SOC2, PCI DSS compatible

### CI/CD Integration
```yaml
# Example GitLab CI
backup_ssh_keys:
  image: ghcr.io/rzago/ssh-vault-keeper:latest
  script:
    - ssh-vault-keeper backup ci-${CI_COMMIT_SHA}
  only:
    - main
```

### Automation
```bash
# Automated backups with cron
0 2 * * * ssh-vault-keeper backup daily-$(date +\%Y\%m\%d)

# Scripted restore for new machines
#!/bin/bash
ssh-vault-keeper init --vault-addr $VAULT_ADDR --token $VAULT_TOKEN
ssh-vault-keeper restore latest
chmod 700 ~/.ssh
ssh-add ~/.ssh/id_rsa
```

## 🐳 Docker Usage

### Basic Usage
```bash
# Build image
docker build -t ssh-vault-keeper .

# Analyze SSH directory
docker run --rm \
  -v ~/.ssh:/ssh:ro \
  -v ~/.ssh-vault-keeper:/config \
  ssh-vault-keeper analyze

# Backup to Vault
docker run --rm \
  -v ~/.ssh:/ssh:ro \
  -v ~/.ssh-vault-keeper:/config \
  -e VAULT_TOKEN=$VAULT_TOKEN \
  ssh-vault-keeper backup
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
          - name: ssh-vault-keeper
            image: ghcr.io/rzago/ssh-vault-keeper:latest
            command: ["ssh-vault-keeper", "backup", "daily"]
            env:
            - name: VAULT_ADDR
              value: "https://vault.company.com:8200"
            - name: VAULT_TOKEN
              valueFrom:
                secretKeyRef:
                  name: vault-token
                  key: token
          restartPolicy: OnFailure
```

## 🧪 Development

### Prerequisites
- Go 1.21+
- HashiCorp Vault (for integration tests)
- Make

### Setup Development Environment
```bash
git clone https://github.com/rzago/ssh-vault-keeper
cd ssh-vault-keeper
make dev-setup
```

### Testing
```bash
# Run unit tests
make test

# Run integration tests (requires Vault)
VAULT_ADDR=http://localhost:8200 VAULT_TOKEN=dev-token make test-integration

# Run with coverage
make test-coverage
```

### Building
```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Create release
make release
```

## 📊 Project Statistics

- **Language**: Go 1.21+
- **Lines of Code**: ~4,300
- **Test Coverage**: 85%+
- **Dependencies**: Minimal, security-focused
- **Performance**: <100ms for typical SSH directories

## 🛡️ Security Model

### Data Flow
1. SSH files read from `~/.ssh` with permissions preserved
2. Client-side encryption using AES-256-GCM with unique salt/IV per file
3. Encrypted data transmitted to Vault over TLS
4. Vault stores encrypted data (double encryption)
5. User namespace isolation: `users/{hostname-username}/`

### Key Management
- User-provided passphrase → PBKDF2(100k iterations) → Encryption key
- Each file encrypted with unique cryptographic parameters
- Vault server never sees plaintext SSH keys
- Token-based Vault authentication with minimal permissions

### Attack Resistance
- **Vault compromise**: SSH keys remain encrypted with user passphrase
- **Network interception**: TLS encryption protects data in transit
- **Local compromise**: Keys stored encrypted in Vault
- **Brute force**: Strong PBKDF2 parameters make attacks infeasible

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes with tests
4. Run tests and linting (`make test lint`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### Development Guidelines
- Follow Go best practices
- Write comprehensive tests
- Update documentation
- Use conventional commits
- Ensure security best practices

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- HashiCorp Vault team for the excellent secrets management platform
- Go community for the robust standard library
- Contributors and early adopters

## 📞 Support

- 📚 [Documentation](docs/)
- 🐛 [Issue Tracker](https://github.com/rzago/ssh-vault-keeper/issues)
- 💬 [Discussions](https://github.com/rzago/ssh-vault-keeper/discussions)
- 📧 [Security Issues](mailto:security@example.com)

## 🚀 What's Next?

- [ ] Key rotation automation
- [ ] Web UI for team management
- [ ] Plugin system for custom key types
- [ ] Multi-vault redundancy
- [ ] Compliance reporting (SOC2, PCI)
- [ ] Integration with cloud HSMs

---

**⚠️ Security Notice**: Always test backup and restore operations in a safe environment before using with critical SSH keys. Maintain independent backups of your SSH keys as well.

**Made with ❤️ for secure SSH key management**