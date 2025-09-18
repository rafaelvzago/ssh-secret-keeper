# SSH Secret Keeper - Self-Update Feature

## Overview

SSH Secret Keeper includes a built-in self-update mechanism that allows you to easily update to the latest version without manual download and installation.

## Features

- **Automatic Platform Detection**: Detects your OS and architecture automatically
- **Version Checking**: Compare current version with latest release
- **Secure Downloads**: Downloads from official GitHub releases with checksum verification
- **Backup & Rollback**: Creates backup of current binary before update
- **Pre-release Support**: Option to include beta/pre-release versions
- **Non-interactive Mode**: Supports automated updates in CI/CD environments

## Basic Usage

### Check for Updates

Check if a new version is available without installing:

```bash
sshsk update --check
```

Output:
```
🔍 Checking for updates...
📌 Current version: v1.0.3
🎯 Latest version:  v1.0.4 (released 2 days ago)
🆕 Update available: Yes

📝 Release Notes:
   - Fixed permission handling bug
   - Added support for Ed25519 keys
   - Improved error messages

📦 An update is available!
Run 'sshsk update' to install the latest version.
```

### Update to Latest Version

Update to the latest stable release:

```bash
sshsk update
```

The update process will:
1. Check for the latest release
2. Download the appropriate binary for your platform
3. Verify the download integrity (checksum)
4. Create a backup of the current binary
5. Replace the binary with the new version
6. Verify the new binary works

### Update to Specific Version

Update to a specific version:

```bash
sshsk update --version v1.0.5
```

### Include Pre-release Versions

Check for and install pre-release/beta versions:

```bash
sshsk update --pre-release
```

## Command Options

| Flag | Description |
|------|-------------|
| `--check` | Check for updates without installing |
| `--version VERSION` | Update to a specific version |
| `--force` | Force update even if already on latest version |
| `--pre-release` | Include pre-release versions |
| `--no-backup` | Don't create backup of current binary |
| `--skip-checksum` | Skip checksum verification (not recommended) |
| `--skip-verify` | Skip new binary verification |
| `-y, --yes` | Skip confirmation prompt (non-interactive) |

## Non-Interactive Updates

For automated environments, use the `-y` flag to skip confirmation:

```bash
sshsk update -y
```

## Permission Requirements

### Linux/macOS

If SSH Secret Keeper is installed in a system directory (e.g., `/usr/local/bin`), you may need to run the update with sudo:

```bash
sudo sshsk update
```

### Windows

If installed in `Program Files`, you may need to run the command as Administrator.

## Configuration

You can configure update behavior in your `config.yaml`:

```yaml
update:
  check_on_startup: false         # Check for updates when running commands
  auto_update: false              # Automatically install updates (not recommended)
  channel: "stable"               # Update channel: stable, beta
  check_interval: "24h"           # How often to check for updates
  github_repo: "rafaelvzago/ssh-secret-keeper"  # GitHub repository
```

## Platform Support

The update feature automatically detects and downloads the correct binary for:

- **Linux**: amd64, arm64
- **macOS**: amd64 (Intel), arm64 (Apple Silicon)
- **Windows**: amd64

## Security

### Checksum Verification

By default, the update process verifies the SHA256 checksum of downloaded binaries against the published checksums. This ensures the integrity of the download.

### Backup & Rollback

Before replacing the binary, a backup is created with the `.backup` extension. If the update fails, the backup is automatically restored.

To manually restore a backup:
```bash
mv /usr/local/bin/sshsk.backup /usr/local/bin/sshsk
```

### HTTPS Only

All downloads are performed over HTTPS from GitHub's secure infrastructure.

## Troubleshooting

### Permission Denied

If you get a permission error:
```bash
Error: insufficient permissions: try running with sudo
```

Solution: Run with elevated privileges:
```bash
sudo sshsk update
```

### Network Issues

If you're behind a proxy or have network restrictions:
1. Ensure you can access `https://github.com`
2. Configure proxy settings if needed:
   ```bash
   export HTTPS_PROXY=http://proxy.example.com:8080
   sshsk update
   ```

### Version Not Found

If a specific version is not found:
```bash
Error: release v1.0.99 not found
```

Check available versions at: https://github.com/rafaelvzago/ssh-secret-keeper/releases

### Checksum Mismatch

If checksum verification fails:
```bash
Error: checksum mismatch: expected abc123..., got def456...
```

This could indicate:
- Corrupted download (try again)
- Network interference
- Security issue (do not use `--skip-checksum`)

## Manual Update

If the self-update feature doesn't work in your environment, you can always update manually:

1. Download the latest release from [GitHub Releases](https://github.com/rafaelvzago/ssh-secret-keeper/releases)
2. Extract the archive
3. Replace your existing binary
4. Verify the installation: `sshsk version`

## CI/CD Integration

### GitHub Actions

```yaml
- name: Update SSH Secret Keeper
  run: |
    sshsk update --check
    sshsk update -y
```

### GitLab CI

```yaml
update-sshsk:
  script:
    - sshsk update --check
    - sshsk update -y
```

### Jenkins

```groovy
sh 'sshsk update --check'
sh 'sshsk update -y'
```

## Development

### Testing Updates

For development/testing, you can update to a specific pre-release:

```bash
# Check pre-release versions
sshsk update --check --pre-release

# Update to specific pre-release
sshsk update --version v1.0.5-beta.1 --pre-release
```

### Building from Source

If you need features not yet released:

```bash
git clone https://github.com/rafaelvzago/ssh-secret-keeper.git
cd ssh-secret-keeper
make build
make install
```

## FAQ

**Q: How often should I update?**
A: Check for updates monthly or when you encounter issues that might be fixed in newer versions.

**Q: Is it safe to auto-update?**
A: While the feature exists, we recommend manual updates after reviewing release notes.

**Q: Can I downgrade to an older version?**
A: Yes, use `--version` with an older version number and `--force` flag.

**Q: What happens if an update fails?**
A: The original binary is automatically restored from backup.

**Q: Can I update without internet access?**
A: No, the update feature requires internet access to download from GitHub.

## Support

For issues with the update feature:
1. Check the [troubleshooting section](#troubleshooting)
2. Review [GitHub Issues](https://github.com/rafaelvzago/ssh-secret-keeper/issues)
3. Create a new issue if your problem isn't addressed

## Related Documentation

- [Installation Guide](INSTALLATION_SCRIPT.md)
- [Quick Start](QUICK_START.md)
- [Configuration](CONFIGURATION.md)
- [Release Process](RELEASE_PROCESS.md)
