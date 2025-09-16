# SSH Secret Keeper - Quick Start Guide

This guide will get you up and running with SSH Secret Keeper in minutes.

## Prerequisites

- HashiCorp Vault server
- Valid Vault token with appropriate permissions
- SSH directory with keys to backup (`~/.ssh`)

## Step 1: Build the Application

```bash
cd /path/to/your/sshsk
make build
```

Or install directly:
```bash
make install
```

## Step 2: Initialize Configuration

```bash
# Initialize with your Vault server
sshsk init --vault-addr https://your-vault-server:8200 --token YOUR_VAULT_TOKEN
```

This will:
- Create configuration at `~/.sshsk/config.yaml`
- Store your token securely at `~/.sshsk/token`
- Test Vault connection
- Create necessary Vault mounts

## Step 3: Analyze Your SSH Directory

Before backing up, see what SSH files you have:

```bash
sshsk analyze --verbose
```

Example output for your SSH directory:
```
🔍 SSH Directory Analysis
════════════════════════

📊 Summary:
  Total files: 25
  Key pairs: 12
  Service keys: 8
  Personal keys: 4
  Work keys: 3
  System files: 3

🔑 Key Pairs:
  • github - ✓ Complete pair
  • gitlab - ✓ Complete pair
  • bitbucket - ✓ Complete pair
  • argocd - ✓ Complete pair
  • cloud_key1 - ✓ Complete pair
  • id_rsa - ✓ Complete pair
  • local - ✓ Complete pair
  • user-cert - ⚠️  Private key only

📂 Categories:
  service (8 files):
    🔐 service1_rsa [service1]
    🔑 service1_rsa.pub [service1]
    🔐 service2_rsa [service2]
    🔑 service2_rsa.pub [service2]
    ...

  personal (4 files):
    🔐 id_rsa
    🔑 id_rsa.pub
    🔐 local_rsa
    🔑 local_rsa.pub

  work (3 files):
    🔐 work_key1.rsa
    🔐 work_key2.rsa

⚙️ System Files:
  ⚙️ config (917 bytes)
  🌐 known_hosts (42556 bytes)
  🎫 authorized_keys (789 bytes)
```

## Step 4: Create Your First Backup

```bash
# Backup everything (you'll be prompted for encryption passphrase)
sshsk backup

# Or with a custom name
sshsk backup pre-migration
```

The backup process:
1. Analyzes your SSH directory
2. Shows summary of what will be backed up
3. Prompts for encryption passphrase
4. Encrypts all files client-side
5. Stores in Vault with your hostname/username namespace

## Step 5: Test Restore (Dry Run)

```bash
# See what would be restored
sshsk restore --dry-run

# Or restore to a test directory
mkdir ~/test-restore
sshsk restore --target-dir ~/test-restore
```

## Step 6: List and Manage Backups

```bash
# List all backups
sshsk list --detailed

# Check system status
sshsk status
```

## Common Use Cases

### Backup Before System Changes
```bash
sshsk backup pre-upgrade-$(date +%Y%m%d)
```

### Selective Restore
```bash
# Restore only specific files
sshsk restore --files "github*,gitlab*"

# Interactive restore
sshsk restore --interactive
```

### New Machine Setup
```bash
# On new machine after installing sshsk
sshsk init --vault-addr https://your-vault-server:8200 --token YOUR_TOKEN
sshsk list
sshsk restore
chmod 700 ~/.ssh
ssh-add -l
```

### Backup Automation
```bash
# Add to crontab for weekly backups
0 2 * * 0 /usr/local/bin/sshsk backup weekly-$(date +\%Y\%m\%d) --passphrase-file ~/.sshsk/backup-passphrase
```

## Configuration Examples

Your configuration at `~/.sshsk/config.yaml`:

```yaml
version: "1.0"
vault:
  address: "https://your-vault-server:8200"
  token_file: "~/.sshsk/token"
  mount_path: "ssh-backups"

backup:
  ssh_dir: "~/.ssh"
  hostname_prefix: true
  retention_count: 10

security:
  algorithm: "AES-256-GCM"
  iterations: 100000
  per_file_encrypt: true
  verify_integrity: true
```

## Troubleshooting

### Connection Issues
```bash
# Test Vault connection
sshsk status --vault

# Common issues:
# - Token expired: Get new token and update ~/.sshsk/token
# - Network issues: Check if your Vault server is accessible
# - Mount issues: Ensure mount_path exists in Vault
```

### Permission Issues
```bash
# Fix SSH directory permissions
chmod 700 ~/.ssh
chmod 600 ~/.ssh/id_rsa ~/.ssh/*_rsa ~/.ssh/config
chmod 644 ~/.ssh/*.pub

# Fix token file permissions
chmod 600 ~/.sshsk/token
```

### Backup Issues
```bash
# Check what would be backed up
sshsk backup --dry-run

# Verbose logging for debugging
sshsk --verbose backup
```

## Security Best Practices

1. **Passphrase Management**
   - Use a strong, unique passphrase for encryption
   - Consider using a password manager
   - Don't store passphrases in scripts

2. **Token Security**
   - Use dedicated tokens with minimal required permissions
   - Rotate tokens regularly
   - Don't share tokens between machines

3. **Network Security**
   - Use HTTPS for Vault in production
   - Consider VPN for remote access
   - Monitor Vault access logs

4. **Backup Verification**
   - Regularly test restore operations
   - Verify backup integrity
   - Keep offline backups as well

## Next Steps

- Set up regular backup automation
- Configure Vault policies for team access
- Integrate with your CI/CD pipeline
- Monitor backup success/failures

## Support

- Run `sshsk --help` for command reference
- Check logs for detailed error messages
- Review configuration with `sshsk status`
