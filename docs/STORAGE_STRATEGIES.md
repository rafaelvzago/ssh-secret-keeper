# SSH Secret Keeper - Storage Strategies Guide

This guide explains the different storage strategies available in SSH Secret Keeper and how to migrate between them for cross-machine and cross-user compatibility.

## Overview

SSH Secret Keeper supports multiple storage strategies that determine how backups are organized in Vault. The choice of strategy affects whether you can restore backups across different machines and user accounts.

## Storage Strategies

### 1. Universal Storage (Recommended)

**Path Pattern**: `shared/{namespace}/backups/{backup-name}` or `shared/backups/{backup-name}`

**Benefits**:
- ✅ **Cross-machine restore**: Backup on laptop, restore on desktop
- ✅ **Cross-user restore**: Backup as user1, restore as user2
- ✅ **Team sharing**: Multiple users can access shared backup namespace
- ✅ **Simple management**: All backups in one location
- ✅ **Container friendly**: Perfect for containerized environments

**Use Cases**:
- Personal use across multiple machines
- Team environments with shared SSH keys
- Container deployments
- CI/CD pipelines

**Configuration**:
```yaml
vault:
  storage_strategy: "universal"
  backup_namespace: ""  # Optional: "personal", "team-a", etc.
```

### 2. User-Scoped Storage

**Path Pattern**: `users/{username}/backups/{backup-name}`

**Benefits**:
- ✅ **Cross-machine restore**: Same user can restore on any machine
- ✅ **User isolation**: Each user has their own backup space
- ⚠️ **Limited sharing**: Only accessible to the same user

**Use Cases**:
- Shared Vault with multiple users
- When user isolation is required
- Cross-machine restore for same user

**Configuration**:
```yaml
vault:
  storage_strategy: "user"
```

### 3. Machine-User Scoped (Legacy)

**Path Pattern**: `users/{hostname-username}/backups/{backup-name}`

**Benefits**:
- ✅ **Maximum isolation**: Each machine-user combination is separate
- ❌ **No cross-machine restore**: Backups tied to specific machine
- ❌ **No cross-user restore**: Backups tied to specific user

**Use Cases**:
- Legacy installations (existing behavior)
- Maximum security isolation requirements
- When cross-machine restore is not needed

**Configuration**:
```yaml
vault:
  storage_strategy: "machine-user"
```

### 4. Custom Storage

**Path Pattern**: `{custom-prefix}/backups/{backup-name}`

**Benefits**:
- ✅ **Flexible organization**: Use any prefix you want
- ✅ **Team/project organization**: Group by team, project, environment
- ⚠️ **Requires configuration**: Must set custom prefix

**Use Cases**:
- Team-based organization ("team-devops", "team-frontend")
- Project-based organization ("project-alpha", "project-beta")
- Environment-based organization ("dev", "staging", "prod")

**Configuration**:
```yaml
vault:
  storage_strategy: "custom"
  custom_prefix: "team-devops"  # Required for custom strategy
```

## Cross-Machine and Cross-User Compatibility

### Path Normalization

SSH Secret Keeper now normalizes file paths to enable cross-user compatibility:

- `/home/user1/.ssh` → `~/.ssh`
- `/home/user2/.ssh` → `~/.ssh`
- `/Users/alice/.ssh` → `~/.ssh` (macOS)
- `C:\Users\bob\.ssh` → `~\.ssh` (Windows)

This allows backups created by one user to be restored by another user on the same or different machines.

### Backup Metadata

New backup format (v2.0) includes:
- **Normalized paths**: Cross-user compatible paths
- **Original user**: For reference and warnings
- **Path version**: Indicates normalization support
- **Cross-user compatibility**: Metadata flag

## Migration Between Strategies

### Migration Commands

```bash
# Show current storage strategy and available options
sshsk migrate-status

# Migrate from legacy machine-user to universal (dry run)
sshsk migrate --from machine-user --to universal --dry-run

# Perform actual migration with cleanup
sshsk migrate --from machine-user --to universal --cleanup

# Migrate to user-scoped for shared Vault
sshsk migrate --from machine-user --to user
```

### Migration Process

1. **Validation**: Checks source and destination compatibility
2. **Backup copying**: Copies all backups to new location
3. **Metadata update**: Adds migration information
4. **Cleanup** (optional): Removes source backups after successful migration

### Migration Examples

#### Example 1: Personal Cross-Machine Setup

**Scenario**: You have SSH keys backed up from your laptop and want to restore them on your desktop.

**Current**: `users/laptop-alice/backups/my-keys`
**Target**: `shared/backups/my-keys`

```bash
# Check what would be migrated
sshsk migrate --from machine-user --to universal --dry-run

# Perform migration
sshsk migrate --from machine-user --to universal --cleanup
```

**Result**: You can now restore backups on any machine or user account.

#### Example 2: Team Environment

**Scenario**: Multiple team members need access to shared SSH keys.

**Current**: `users/server1-alice/backups/team-keys`
**Target**: `shared/team-devops/backups/team-keys`

```bash
# Update configuration
# vault:
#   storage_strategy: "universal"
#   backup_namespace: "team-devops"

# Migrate existing backups
sshsk migrate --from machine-user --to universal
```

**Result**: All team members can access shared keys using the same backup names.

#### Example 3: User Isolation

**Scenario**: Shared Vault server with multiple users who need isolation.

**Current**: `users/machine1-alice/backups/keys`
**Target**: `users/alice/backups/keys`

```bash
# Migrate to user-scoped storage
sshsk migrate --from machine-user --to user --cleanup
```

**Result**: Alice can restore her keys on any machine, but other users cannot access them.

## Configuration Examples

### Universal Storage with Namespace

```yaml
# ~/.ssh-secret-keeper/config.yaml
vault:
  address: "https://vault.company.com:8200"
  storage_strategy: "universal"
  backup_namespace: "personal"  # Optional organization

backup:
  ssh_dir: "~/.ssh"
  normalize_paths: true
  cross_machine_restore: true
```

### Team-Based Custom Storage

```yaml
# ~/.ssh-secret-keeper/config.yaml
vault:
  address: "https://vault.company.com:8200"
  storage_strategy: "custom"
  custom_prefix: "team-devops"

backup:
  ssh_dir: "~/.ssh"
  normalize_paths: true
  cross_machine_restore: true
```

### User-Scoped for Shared Vault

```yaml
# ~/.ssh-secret-keeper/config.yaml
vault:
  address: "https://vault.company.com:8200"
  storage_strategy: "user"

backup:
  ssh_dir: "~/.ssh"
  normalize_paths: true
  cross_machine_restore: true
```

## Security Considerations

### Access Control Policies

Different strategies require different Vault ACL policies:

#### Universal Storage Policy
```hcl
# Allow access to shared namespace
path "ssh-backups/data/shared/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

path "ssh-backups/metadata/shared/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
```

#### User-Scoped Policy
```hcl
# Allow access only to user's own namespace
path "ssh-backups/data/users/{{identity.entity.name}}/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

path "ssh-backups/metadata/users/{{identity.entity.name}}/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
```

### Cross-User Restore Security

When restoring backups across users:

1. **File permissions are preserved** from the original backup
2. **SSH directory permissions** are enforced (0700)
3. **Warning messages** are logged for cross-user operations
4. **Original user information** is stored for audit purposes

## Troubleshooting

### Common Migration Issues

#### 1. Backup Name Conflicts
**Problem**: Destination already has backups with the same names
**Solution**: Use unique backup names or migrate to a namespace

#### 2. Permission Denied
**Problem**: Vault token doesn't have access to destination path
**Solution**: Update Vault policies or use appropriate token

#### 3. Large Migration
**Problem**: Many backups to migrate
**Solution**: Use `--dry-run` first, migrate in batches if needed

### Validation Errors

```bash
# Check migration feasibility
sshsk migrate --from machine-user --to universal --dry-run

# View detailed validation results
sshsk migrate-status
```

### Cross-User Restore Issues

#### 1. Permission Warnings
**Problem**: Restored files have incorrect permissions
**Solution**: Check SSH directory permissions (should be 0700)

#### 2. Path Resolution Errors
**Problem**: Cannot resolve `~/.ssh` path
**Solution**: Ensure target user has a valid home directory

## Best Practices

### 1. Strategy Selection

- **Personal use**: Choose `universal` for maximum flexibility
- **Team sharing**: Use `universal` with namespaces or `custom` with team prefixes
- **Multi-user Vault**: Use `user` for isolation with cross-machine capability
- **Maximum security**: Keep `machine-user` if cross-machine restore isn't needed

### 2. Migration Planning

- Always run `--dry-run` first
- Test with a few backups before migrating all
- Keep source backups until restore is verified
- Update team documentation after migration

### 3. Backup Naming

- Use descriptive names: `dev-keys-2024-01`, `prod-access-keys`
- Include dates for versioning: `backup-20240117`
- Avoid machine-specific names when using universal storage

### 4. Security

- Review Vault policies after changing strategies
- Monitor cross-user restore operations
- Use appropriate authentication methods
- Regular backup verification

## Environment Variables

Override configuration with environment variables:

```bash
# Storage strategy
export SSH_VAULT_STORAGE_STRATEGY="universal"

# Custom prefix for custom strategy
export SSH_VAULT_CUSTOM_PREFIX="team-devops"

# Backup namespace for universal strategy
export SSH_VAULT_BACKUP_NAMESPACE="personal"

# Enable path normalization
export SSH_VAULT_BACKUP_NORMALIZE_PATHS="true"

# Enable cross-machine restore
export SSH_VAULT_BACKUP_CROSS_MACHINE_RESTORE="true"
```

## Conclusion

The new storage strategies in SSH Secret Keeper provide flexibility for different use cases while maintaining security. The universal storage strategy is recommended for most users as it enables cross-machine and cross-user restore capabilities while keeping backup management simple.

For existing users, migration from the legacy machine-user strategy is straightforward and can be done safely with the built-in migration tools.
