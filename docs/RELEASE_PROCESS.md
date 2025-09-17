# SSH Secret Keeper Release Process

This document describes the complete release process for SSH Secret Keeper.

## Overview

The project uses GoReleaser for automated releases with GitHub Actions. The release process is designed to be safe, automated, and consistent.

## Release Types

### 1. Automated Release (Recommended)
Triggered by pushing a git tag to the repository.

### 2. Manual Release
For special cases or local testing.

## Prerequisites

Before creating a release, ensure:

1. **GoReleaser is installed** (for local releases):
   ```bash
   go install github.com/goreleaser/goreleaser@latest
   ```

2. **All tests pass**:
   ```bash
   make test
   ```

3. **Working directory is clean** (no uncommitted changes)

4. **You're on the main/master/developer branch** (recommended)

## Automated Release Process

### Step 1: Prepare the Release

```bash
# Validate the environment and configuration
make release-prepare VERSION=x.y.z

# This will check:
# - Working directory is clean
# - Current branch (warns if not main/master/developer)
# - Tag doesn't already exist
# - GoReleaser is available and configuration is valid
```

### Step 2: Create and Push Tag

```bash
# Create the tag locally
make tag-release VERSION=x.y.z

# Push the tag to trigger automated release
make push-tag VERSION=x.y.z
```

Or combine both steps:

```bash
# Complete automated release workflow
make release VERSION=x.y.z
```

### Step 3: Monitor the Release

After pushing the tag:
1. GitHub Actions will automatically start the release workflow
2. Monitor progress at: https://github.com/rafaelvzago/ssh-secret-keeper/actions
3. Release will be published at: https://github.com/rafaelvzago/ssh-secret-keeper/releases

## Manual/Local Release Process

### For Local Testing

```bash
# Create a snapshot release (no tag required)
make release-snapshot

# This creates binaries in dist/ for local testing
```

### For Manual Release (with existing tag)

```bash
# Ensure you're on a tagged commit
git checkout v1.2.3

# Create release from existing tag
make release-local
```

## Release Workflow Details

### What Happens During Release

1. **Validation**:
   - Environment checks
   - Configuration validation
   - Git state verification

2. **Building**:
   - Cross-platform binaries (Linux, macOS)
   - Architecture support (amd64, arm64)
   - Static binaries with version information

3. **Packaging**:
   - tar.gz archives for Unix systems
   - Checksums for all artifacts

4. **Publishing**:
   - GitHub release with release notes
   - Binary attachments
   - Automatic changelog generation

### Included Files in Release

Each release archive contains:
- `sshsk` binary
- `README.md`
- `LICENSE`
- `CHANGELOG.md`
- `configs/config.example.yaml`

## Release Configuration

The release process is configured in:
- `.goreleaser.yml` - GoReleaser configuration
- `.github/workflows/release.yml` - GitHub Actions workflow
- `Makefile` - Release targets and validation

## Version Management

### Version Format
Follow semantic versioning: `vMAJOR.MINOR.PATCH`

Examples:
- `v1.0.0` - Major release
- `v1.1.0` - Minor release (new features)
- `v1.1.1` - Patch release (bug fixes)

### Pre-release Versions
For pre-release versions, use:
- `v1.0.0-alpha.1` - Alpha release
- `v1.0.0-beta.1` - Beta release
- `v1.0.0-rc.1` - Release candidate

## Container Images

Container images are built locally during the release process. To publish container images:

1. Build images:
   ```bash
   make container-build-branch
   ```

2. Tag and push to registry (manual step):
   ```bash
   # Example for GitHub Container Registry
   docker tag ssh-secret-keeper:latest ghcr.io/rafaelvzago/ssh-secret-keeper:latest
   docker push ghcr.io/rafaelvzago/ssh-secret-keeper:latest
   ```

## Troubleshooting

### Common Issues

1. **"Tag already exists"**:
   ```bash
   # Delete the tag locally and remotely
   git tag -d v1.2.3
   git push origin --delete v1.2.3
   ```

2. **"Working directory has uncommitted changes"**:
   ```bash
   # Commit or stash your changes
   git add .
   git commit -m "Prepare for release"
   ```

3. **"GoReleaser not found"**:
   ```bash
   go install github.com/goreleaser/goreleaser@latest
   ```

4. **Release workflow fails**:
   - Check GitHub Actions logs
   - Verify GITHUB_TOKEN permissions
   - Ensure all tests pass

### Validation Commands

```bash
# Check release configuration
make release-check

# Test release process locally
make release-snapshot

# Validate environment before release
make release-prepare VERSION=x.y.z
```

## Release Checklist

Before creating a release:

- [ ] All tests pass (`make test`)
- [ ] Documentation is updated
- [ ] CHANGELOG.md is updated
- [ ] Version number follows semantic versioning
- [ ] Working directory is clean
- [ ] Release configuration is valid (`make release-check`)

After creating a release:

- [ ] Verify release on GitHub
- [ ] Test download and installation
- [ ] Update documentation if needed
- [ ] Announce release to users

## Emergency Procedures

### Rollback a Release

1. **Delete the GitHub release**:
   - Go to GitHub releases page
   - Delete the problematic release

2. **Delete the git tag**:
   ```bash
   git tag -d vX.Y.Z
   git push origin --delete vX.Y.Z
   ```

3. **Fix issues and create new release**

### Hotfix Release

For critical bug fixes:

1. Create hotfix branch from the release tag
2. Apply minimal fixes
3. Create new patch release
4. Follow normal release process

## Contact

For questions about the release process:
- Create an issue on GitHub
- Check existing documentation
- Review GoReleaser documentation: https://goreleaser.com/
