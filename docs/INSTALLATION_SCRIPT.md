# SSH Secret Keeper Installation Script

This document describes the installation script (`install.sh`) for SSH Secret Keeper.

## Overview

The installation script provides an automated, cross-platform way to install SSH Secret Keeper. It automatically detects your operating system and architecture, downloads the appropriate binary from GitHub releases, verifies checksums, and installs it to the appropriate location.

## Features

- **Auto-detection**: Automatically detects OS (Linux/macOS) and architecture (amd64/arm64)
- **Security**: Verifies SHA256 checksums of downloaded binaries
- **Flexible installation**: Supports system-wide and user-local installation
- **Update handling**: Detects existing installations and handles updates
- **Fallback support**: Can build from source if pre-built binaries aren't available
- **Comprehensive error handling**: Clear error messages and graceful failure handling

## Usage

### Basic Installation

```bash
# Install latest version (system-wide, requires sudo)
curl -sSL https://raw.githubusercontent.com/rafaelvzago/ssh-secret-keeper/refs/heads/main/install.sh | bash
```

### Installation Options

```bash
# Install to user directory (no sudo required)
curl -sSL https://raw.githubusercontent.com/rafaelvzago/ssh-secret-keeper/refs/heads/main/install.sh | bash -s -- --user

# Install specific version
curl -sSL https://raw.githubusercontent.com/rafaelvzago/ssh-secret-keeper/refs/heads/main/install.sh | bash -s -- --version 1.2.0

# Install to custom directory
curl -sSL https://raw.githubusercontent.com/rafaelvzago/ssh-secret-keeper/refs/heads/main/install.sh | bash -s -- --install-dir /opt/bin

# View help
curl -sSL https://raw.githubusercontent.com/rafaelvzago/ssh-secret-keeper/refs/heads/main/install.sh | bash -s -- --help
```

### Local Installation

```bash
# Download and run locally
wget https://github.com/rafaelvzago/ssh-secret-keeper/raw/main/install.sh
chmod +x install.sh
./install.sh [OPTIONS]
```

## Command Line Options

| Option | Description | Example |
|--------|-------------|---------|
| `--version VERSION` | Install specific version | `--version 1.2.0` |
| `--install-dir DIR` | Custom installation directory | `--install-dir /opt/bin` |
| `--user` | Install to user directory (~/.local/bin) | `--user` |
| `--help` | Show help message | `--help` |
| `--debug` | Enable debug output | `--debug` |

## Installation Locations

The script uses the following priority for installation:

1. **System-wide installation** (`/usr/local/bin`):
   - Preferred location for system-wide access
   - Requires sudo privileges
   - Binary available to all users

2. **User-local installation** (`~/.local/bin`):
   - Used when `--user` flag is specified
   - Used as fallback when system installation fails
   - No sudo required
   - May need to add to PATH

3. **Custom directory**:
   - Used when `--install-dir` is specified
   - User is responsible for PATH management

## Supported Platforms

| OS | Architecture | Status |
|----|--------------|--------|
| Linux | amd64 | ✅ Supported |
| Linux | arm64 | ✅ Supported |
| macOS | amd64 | ✅ Supported |
| macOS | arm64 | ✅ Supported |

## Requirements

### Required Tools
- `curl` or `wget` (for downloading)
- `tar` (for extracting archives)

### Optional Tools
- `sha256sum` or `shasum` (for checksum verification)
- `git` and `go` 1.21+ (for building from source)
- `make` (for building from source with Makefile)
- `sudo` (for system-wide installation)

## Security Features

### Checksum Verification
The script automatically downloads and verifies SHA256 checksums for all binaries:

```bash
# Downloaded files:
ssh-secret-keeper-1.2.0-linux-amd64.tar.gz
checksums.txt

# Verification process:
1. Extract expected checksum from checksums.txt
2. Calculate actual checksum of downloaded file
3. Compare checksums and fail if they don't match
```

### Safe Downloads
- Uses HTTPS URLs for all downloads
- Verifies GitHub releases API responses
- Creates temporary directories with secure permissions
- Cleans up temporary files on exit

## Error Handling

The script provides comprehensive error handling with clear messages:

### Common Errors

| Error | Cause | Solution |
|-------|-------|----------|
| "Unsupported OS" | Running on Windows or other unsupported OS | Use WSL on Windows or build from source |
| "Unsupported architecture" | Running on 32-bit or other unsupported arch | Use a supported platform or build from source |
| "Failed to download" | Network issues or invalid version | Check internet connection and version |
| "Checksum verification failed" | Corrupted download | Retry installation |
| "Neither curl nor wget found" | Missing download tools | Install curl or wget |

### Debug Mode

Enable debug mode for troubleshooting:

```bash
curl -sSL https://raw.githubusercontent.com/rafaelvzago/ssh-secret-keeper/refs/heads/main/install.sh | bash -s -- --debug
```

Debug mode provides:
- Detailed platform detection information
- Download URLs and file paths
- Installation directory decisions
- Checksum verification details

## Fallback: Build from Source

If pre-built binaries aren't available, the script can build from source:

### Requirements for Source Build
- Go 1.21 or later
- Git
- Make (optional, but recommended)

### Build Process
1. Clone the repository
2. Use `make build` if available, otherwise `go build`
3. Install the built binary

## PATH Management

### System Installation
When installing to `/usr/local/bin`, the binary is typically already in PATH.

### User Installation
When installing to `~/.local/bin`, you may need to add it to PATH:

```bash
# Add to your shell profile (~/.bashrc, ~/.zshrc, etc.)
export PATH="$PATH:$HOME/.local/bin"

# Reload your shell
source ~/.bashrc  # or restart terminal
```

The script will warn you if the installation directory isn't in PATH.

## CI/CD Integration

The installation script is perfect for CI/CD environments:

### GitHub Actions
```yaml
- name: Install SSH Secret Keeper
  run: curl -sSL https://raw.githubusercontent.com/rafaelvzago/ssh-secret-keeper/refs/heads/main/install.sh | bash
```

### GitLab CI
```yaml
before_script:
  - curl -sSL https://raw.githubusercontent.com/rafaelvzago/ssh-secret-keeper/refs/heads/main/install.sh | bash
```

### Jenkins
```groovy
sh 'curl -sSL https://raw.githubusercontent.com/rafaelvzago/ssh-secret-keeper/refs/heads/main/install.sh | bash'
```

## Testing

The repository includes a test script (`test-install.sh`) that validates:
- Script syntax
- Help message functionality
- Argument parsing
- Platform detection
- Prerequisites checking
- Error handling

Run tests:
```bash
./test-install.sh
```

## Troubleshooting

### Installation Issues

1. **Permission denied**:
   ```bash
   # Try user installation
   curl -sSL https://raw.githubusercontent.com/rafaelvzago/ssh-secret-keeper/refs/heads/main/install.sh | bash -s -- --user
   ```

2. **Network issues**:
   ```bash
   # Check connectivity
   curl -I https://api.github.com/repos/rafaelvzago/ssh-secret-keeper/releases/latest
   ```

3. **Checksum failures**:
   ```bash
   # Retry installation (might be temporary network issue)
   curl -sSL https://raw.githubusercontent.com/rafaelvzago/ssh-secret-keeper/refs/heads/main/install.sh | bash
   ```

4. **Platform not supported**:
   ```bash
   # Try building from source
   git clone https://github.com/rafaelvzago/ssh-secret-keeper
   cd ssh-secret-keeper
   make build
   sudo make install
   ```

### Verification

After installation, verify it works:

```bash
# Check installation
which sshsk
sshsk --version

# Test basic functionality
sshsk --help
```

## Development

The installation script is located at `/install.sh` in the repository root.

### Script Structure
- Platform detection functions
- Version resolution from GitHub API
- Download and verification functions
- Installation logic with fallbacks
- Error handling and cleanup
- Command-line argument parsing

### Contributing
When modifying the installation script:
1. Update the test script if needed
2. Test on multiple platforms
3. Update this documentation
4. Ensure error handling is comprehensive

## Security Considerations

- The script downloads and executes code from GitHub
- Always review scripts before running with `curl | bash`
- For maximum security, download and inspect the script first:

```bash
# Download and inspect
curl -sSL https://github.com/rafaelvzago/ssh-secret-keeper/raw/main/install.sh > install.sh
less install.sh  # Review the script
chmod +x install.sh
./install.sh
```

- The script verifies checksums to ensure binary integrity
- Uses HTTPS for all network communication
- Creates temporary directories with secure permissions
