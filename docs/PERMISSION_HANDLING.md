# SSH Permission Handling

## Overview

SSH Vault Keeper provides **perfect permission preservation** for SSH files, ensuring that your SSH keys maintain the exact same permissions after backup and restore operations. This is critical for SSH security, as SSH clients will reject keys with incorrect permissions.

## Permission Preservation Features

### ✅ What is Preserved

1. **Exact File Permissions** (owner, group, world read/write/execute)
2. **File Modification Times** 
3. **SSH Directory Permissions** (automatically set to 0700)
4. **Permission Metadata** stored in backup

### 🔍 Permission Validation

The tool automatically validates permissions and provides warnings for:

- **Private keys with world/group readable permissions** (SSH will reject these)
- **SSH directory not set to 0700**
- **Unusual permission combinations**
- **Permission mismatches after restore**

## SSH Permission Best Practices

### Recommended Permissions

| File Type | Permission | Octal | Description |
|-----------|------------|-------|-------------|
| SSH Directory | `drwx------` | `0700` | Only user can access |
| Private Keys | `-rw-------` | `0600` | Only user can read/write |
| Public Keys | `-rw-r--r--` | `0644` | World readable (optional: 0600) |
| SSH Config | `-rw-------` | `0600` | Only user can read/write |
| Known Hosts | `-rw-r--r--` | `0644` | World readable (optional: 0600) |
| Authorized Keys | `-rw-------` | `0600` | Only user can read/write |

### Critical Security Rules

1. **Private keys MUST be 0600** - SSH will refuse to use keys readable by others
2. **SSH directory MUST be 0700** - SSH will warn or refuse if others can access
3. **Never use 0777** - This allows anyone to read your SSH keys

## Implementation Details

### During Backup

```go
// Permissions captured during file analysis
keyInfo.Permissions = info.Mode()  // Full file mode including permissions

// Stored in FileData structure
type FileData struct {
    Permissions os.FileMode `json:"permissions"`
    // ... other fields
}
```

### During Restore

```go
// Exact permissions restored
os.WriteFile(targetPath, content, fileData.Permissions)

// Permissions verified after restoration
actualPerms := stat.Mode().Perm()
expectedPerms := fileData.Permissions & os.ModePerm
```

### Permission Validation

The tool performs multi-layer permission validation:

1. **Pre-restore validation** - Checks backup data
2. **During restore** - Warns about critical permission issues
3. **Post-restore verification** - Confirms permissions match

## Usage Examples

### Backup with Permission Summary

```bash
$ ssh-vault-keeper backup my-keys

✓ Backup 'my-keys' completed successfully
Files backed up: 28
Total size: 125440 bytes

Permission Preservation:
• 0600: 14 files
• 0644: 10 files  
• 0700: 3 files

✅ All file permissions have been preserved in the backup
```

### Restore with Permission Verification

```bash
$ ssh-vault-keeper restore my-keys

Restoring files to /home/user/.ssh...
Verifying file permissions...
✓ All file permissions verified

✓ Restore completed successfully
Files restored: 28

Permission Summary:
• SSH directory: /home/user/.ssh (0700)
• Private keys: 14 files (0600)
• Public keys: 10 files (0644/0600)

💡 Next steps:
  ssh-add -l    # Check SSH agent
  ssh-add /home/user/.ssh/id_rsa  # Add key to agent
```

### Dry Run with Permission Preview

```bash
$ ssh-vault-keeper restore my-keys --dry-run

[DRY RUN] Would restore file: id_rsa
  Target: /home/user/.ssh/id_rsa
  Permissions: -rw-------
  
[DRY RUN] Would restore file: id_rsa.pub
  Target: /home/user/.ssh/id_rsa.pub
  Permissions: -rw-r--r--
```

## Troubleshooting Permission Issues

### Common Warnings

#### "Permission validation warning"
**Cause**: File has unusual permissions  
**Action**: Review the specific warning in logs, usually safe to continue

#### "CRITICAL: Private key has world/group readable permissions"
**Cause**: Private key is readable by others (SSH will reject)  
**Action**: Fix immediately - this is a security risk

#### "SSH directory permission issue"
**Cause**: SSH directory is not 0700  
**Action**: Directory will be automatically fixed during restore

### Manual Permission Fixes

```bash
# Fix SSH directory permissions
chmod 700 ~/.ssh

# Fix private key permissions  
chmod 600 ~/.ssh/id_rsa ~/.ssh/*_rsa

# Fix public key permissions (optional)
chmod 644 ~/.ssh/*.pub

# Verify permissions
ls -la ~/.ssh/
```

### Permission Verification Commands

```bash
# Check current SSH permissions
ls -la ~/.ssh/

# Verify SSH directory
stat -c "%n %a" ~/.ssh

# Verify specific files
stat -c "%n %a" ~/.ssh/id_rsa ~/.ssh/id_rsa.pub

# Test SSH key acceptance
ssh-add -t ~/.ssh/id_rsa
```

## Advanced Features

### Container Environments

When running in containers, ensure proper volume mounting preserves permissions:

```bash
# Docker with proper permission preservation
docker run --rm \
  -v ~/.ssh:/ssh:ro \
  -v ~/.ssh-vault-keeper:/config \
  --user $(id -u):$(id -g) \
  ghcr.io/rzago/ssh-vault-keeper backup

# Podman with SELinux context
podman run --rm \
  -v ~/.ssh:/ssh:ro,Z \
  -v ~/.ssh-vault-keeper:/config:Z \
  --user $(id -u):$(id -g) \
  ghcr.io/rzago/ssh-vault-keeper backup
```

### CI/CD Considerations

For automated environments:

```yaml
# GitLab CI with permission preservation
backup_ssh:
  script:
    - chmod 700 $HOME/.ssh  # Ensure directory is secure
    - ssh-vault-keeper backup "ci-$CI_COMMIT_SHA"
    - ssh-vault-keeper verify-permissions  # Optional validation
```

## Security Considerations

### Why Permissions Matter

1. **SSH Client Enforcement**: SSH refuses to use keys with incorrect permissions
2. **Security Defense**: Prevents unauthorized access to private keys
3. **Compliance**: Many security policies require specific SSH permissions
4. **Audit Requirements**: Proper permissions are often audited

### Permission-Related Attacks

- **Privilege Escalation**: Incorrect permissions can allow unauthorized key access
- **Key Theft**: World-readable private keys can be stolen by other users
- **Directory Traversal**: Incorrect SSH directory permissions expose key structure

## Testing and Validation

### Automated Testing

The tool includes comprehensive permission tests:

```bash
# Run permission-specific tests
go test -v ./internal/ssh -run TestPermission

# Integration tests with permission validation
make test-permissions
```

### Manual Validation

```bash
# Test backup and restore cycle
ssh-vault-keeper backup test-permissions
ssh-vault-keeper restore test-permissions --target-dir /tmp/test-ssh

# Compare original and restored permissions
diff <(ls -la ~/.ssh/) <(ls -la /tmp/test-ssh/)
```

## Best Practices Summary

1. **Always verify permissions** after restore operations
2. **Use dry-run mode** to preview permission changes
3. **Monitor logs** for permission warnings
4. **Test SSH functionality** after restore
5. **Follow the principle of least privilege** for SSH files
6. **Regular permission audits** of SSH directories

---

**Security Note**: Incorrect SSH permissions are a common cause of SSH connection failures and security vulnerabilities. Always ensure your SSH files have the correct permissions for both security and functionality.
