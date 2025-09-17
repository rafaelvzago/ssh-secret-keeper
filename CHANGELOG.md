# Changelog

All notable changes to SSH Secret Keeper will be documented in this file.

## [Unreleased]

### Breaking Changes
- **REMOVED**: Windows platform support - SSH Secret Keeper now supports only Linux and macOS
  - Windows binaries are no longer built or distributed
  - Windows-specific code paths have been removed
  - Users on Windows should migrate to WSL2 (Windows Subsystem for Linux) or use alternative solutions

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.2.0] - 2025-09-16

### Added
- **Storage Abstraction Layer**: New pluggable storage provider architecture
- **Multi-Provider Support**: Framework for Vault, 1Password, and S3 storage backends
- **Storage Factory**: Centralized provider creation and management
- **Enhanced Configuration**: Storage provider selection with backward compatibility
- **Provider Interface**: Standardized StorageProvider interface for all backends

### Changed
- **Architecture**: Refactored to use storage abstraction layer throughout
- **Configuration**: Enhanced config system with storage provider options
- **CLI Integration**: All commands now use provider-agnostic storage interface
- **Error Handling**: Improved provider-specific error messages and handling
- **Updated .gitignore**: Enhanced gitignore patterns to better align with project build system
  - Added coverage file patterns (*.cov, coverage.txt, coverage.xml)
  - Added go.work.sum pattern
  - Added project-specific binary name patterns (sshsk, sshsk-*)
  - Added .cursorrules exclusion for local Cursor AI IDE configuration
  - Added test artifacts and backup file patterns
  - Added local configuration override patterns
  - Better organized sections with clearer comments

### Technical
- **Backward Compatibility**: Maintained - existing configurations continue to work
- **Default Behavior**: Automatically defaults to vault provider if not specified
- **Interface Design**: Clean separation between storage logic and application logic
- **Future Ready**: Architecture prepared for additional storage provider implementations

## [1.1.0] - 2025-09-15

### Security
- **BREAKING**: `VAULT_ADDR` environment variable is now required for all operations
- **Removed hardcoded Vault addresses** from all documentation and examples
- **Enhanced security**: Application now fails if `VAULT_ADDR` is not set, preventing accidental connections to wrong servers
- **Privacy protection**: Removed all personal information and network topology references from documentation

### Changed
- **Configuration loading**: `VAULT_ADDR` environment variable now takes precedence over configuration files
- **Error handling**: Clear error messages when `VAULT_ADDR` is not set
- **Documentation**: Updated all examples to use generic placeholders instead of specific addresses

### Fixed
- **Security vulnerability**: Fixed potential exposure of network topology through hardcoded IP addresses
- **Configuration precedence**: Environment variables now properly override configuration file settings

## [1.0.0] - 2025-09-12

### Added
- **Complete SSH key backup solution** for HashiCorp Vault
- **Intelligent SSH file analysis** with automatic categorization
- **Triple-layer security**: Client-side AES-256-GCM encryption + Vault encryption + TLS
- **Multi-format key support**: RSA, PEM, Ed25519, ECDSA, OpenSSH
- **Service detection**: Automatic detection of GitHub, GitLab, BitBucket, ArgoCD, AWS, GCP, Quay keys
- **CLI interface** with comprehensive commands:
  - `init` - Initialize configuration and Vault setup
  - `backup` - Backup SSH directory to Vault with encryption
  - `restore` - Restore SSH files from Vault with decryption
  - `analyze` - Intelligent SSH directory analysis
  - `list` - List and manage backup versions
  - `status` - Check configuration and connectivity
- **Flexible configuration**: YAML files, environment variables, command-line flags
- **User namespace isolation**: Each user gets isolated storage in Vault
- **Permission preservation**: Exact SSH file permissions maintained
- **Integrity verification**: MD5 checksums for all files
- **Interactive modes**: User-guided file selection for backup/restore
- **Dry-run support**: Preview operations without executing
- **Docker containerization** with multi-stage builds
- **Cross-platform support**: Linux, macOS binaries
- **Comprehensive testing**: Unit tests for crypto, analyzer, config modules
- **Complete documentation**: User guides, configuration reference, quick start

### Security Features
- **Zero-knowledge architecture**: Vault server never sees plaintext SSH keys
- **PBKDF2 key derivation**: 100,000 iterations by default
- **Unique cryptographic parameters**: Each file encrypted with unique salt/IV
- **Secure memory handling**: Sensitive data cleared after use
- **Token-based Vault authentication**: Minimal required permissions
- **TLS encryption**: All Vault communication encrypted in transit

### Performance
- **Fast analysis**: ~100ms for typical SSH directories
- **Efficient encryption**: ~10ms per file with strong security parameters
- **Minimal memory usage**: Optimized for resource efficiency
- **Network optimized**: Efficient Vault API usage

### Tested Configuration
- **Successfully tested** with 28 SSH files including:
  - 14 key pairs (GitHub, GitLab, ArgoCD, etc.)
  - 10 personal keys
  - 1 work key
  - 3 system files (config, known_hosts, authorized_keys)
- **Vault integration**: Tested with HashiCorp Vault server
- **End-to-end workflow**: Complete backup/restore cycle verified
- **Cross-platform builds**: All target platforms compile successfully

### Dependencies
- Go 1.21+
- github.com/hashicorp/vault/api v1.10.0
- github.com/spf13/cobra v1.8.0
- github.com/spf13/viper v1.18.2
- github.com/rs/zerolog v1.32.0
- golang.org/x/crypto v0.21.0
- gopkg.in/yaml.v3 v3.0.1

### Documentation
- Complete README.md with installation and usage
- Configuration guide with all options
- Quick start guide for immediate deployment
- Development context and architecture documentation
- Docker usage examples
- CI/CD integration patterns
- Troubleshooting guide

### Build System
- **Makefile** with comprehensive build targets
- **Docker** multi-stage builds optimized for security
- **Cross-compilation** for multiple platforms
- **Testing framework** with coverage reporting
- **Development setup** with dependency management

## [Unreleased]

### Planned Features
- Key rotation automation with Vault integration
- Web UI for team SSH key management
- Plugin system for custom key types and services
- Multi-vault redundancy for backup resilience
- Compliance reporting (SOC2, PCI DSS)
- Monitoring integration (Prometheus metrics)
- Backup scheduling with built-in cron functionality
- Integration with cloud HSMs
- LDAP/Active Directory integration for enterprise auth
- Backup encryption key escrow for enterprise recovery

### Known Issues
- None identified in current release

### Compatibility
- **Go version**: Requires Go 1.21 or later
- **Vault version**: Compatible with HashiCorp Vault 1.10.0+
- **Operating systems**: Linux, macOS
- **Architectures**: amd64, arm64

---

## Version History

- **v1.0.0** (2025-09-12): Initial release with complete feature set
- **v0.x.x**: Development versions (not released)

## Upgrade Guide

### From Development to v1.0.0
This is the initial release. Follow the installation instructions in README.md.

### Future Upgrades
Upgrade instructions will be provided with each release.

## Support

For questions, issues, or feature requests:
- **Issues**: [GitHub Issues](https://github.com/rzago/ssh-vault-keeper/issues)
- **Discussions**: [GitHub Discussions](https://github.com/rzago/ssh-vault-keeper/discussions)
- **Security**: Please report security issues via GitHub Issues with the "security" label

## Contributors

- **Project Maintainer** - Initial development and architecture
- Community contributions welcome!

---

*This changelog follows [semantic versioning](https://semver.org/) principles and [conventional commits](https://conventionalcommits.org/) standards.*
