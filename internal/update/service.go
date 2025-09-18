package update

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

// Service handles the update process
type Service struct {
	github     *GitHubClient
	downloader *Downloader
	platform   Platform
	config     *UpdateConfig
}

// NewService creates a new update service
func NewService(config *UpdateConfig) *Service {
	if config == nil {
		config = &UpdateConfig{
			GitHubRepo: defaultGitHubRepo,
		}
	}

	return &Service{
		github:     NewGitHubClient(config.GitHubRepo),
		downloader: NewDownloader(),
		platform:   DetectPlatform(),
		config:     config,
	}
}

// CheckForUpdates checks if a new version is available
func (s *Service) CheckForUpdates(currentVersion string, includePreRelease bool) (*UpdateStatus, error) {
	log.Info().
		Str("current_version", currentVersion).
		Bool("include_prerelease", includePreRelease).
		Msg("Checking for updates")

	release, err := s.github.GetLatestRelease(includePreRelease)
	if err != nil {
		return nil, fmt.Errorf("failed to check for updates: %w", err)
	}

	latestVersion := FormatVersion(release.TagName)
	currentVersion = FormatVersion(currentVersion)

	status := &UpdateStatus{
		CurrentVersion:  currentVersion,
		LatestVersion:   latestVersion,
		UpdateAvailable: IsNewer(latestVersion, currentVersion),
		ReleaseNotes:    release.Body,
		PublishedAt:     release.PublishedAt,
	}

	log.Info().
		Str("current", currentVersion).
		Str("latest", latestVersion).
		Bool("update_available", status.UpdateAvailable).
		Msg("Update check completed")

	return status, nil
}

// Update performs the update to a specific version or latest
func (s *Service) Update(currentVersion string, options UpdateOptions) error {
	// Check permissions first
	if err := s.checkPermissions(); err != nil {
		return err
	}

	// Get the target release
	var release *Release
	var err error

	if options.Version != "" {
		log.Info().Str("version", options.Version).Msg("Updating to specific version")
		release, err = s.github.GetReleaseByTag(options.Version)
	} else {
		log.Info().Bool("prerelease", options.PreRelease).Msg("Updating to latest version")
		release, err = s.github.GetLatestRelease(options.PreRelease)
	}

	if err != nil {
		return fmt.Errorf("failed to get release: %w", err)
	}

	targetVersion := FormatVersion(release.TagName)
	currentVersion = FormatVersion(currentVersion)

	// Check if update is needed
	if !options.Force && !IsNewer(targetVersion, currentVersion) {
		log.Info().
			Str("current", currentVersion).
			Str("target", targetVersion).
			Msg("Already on the latest version")
		return fmt.Errorf("already on version %s (use --force to reinstall)", currentVersion)
	}

	// Find asset for platform
	asset, err := s.github.FindAssetForPlatform(release, s.platform)
	if err != nil {
		return fmt.Errorf("failed to find release for platform: %w", err)
	}

	// Get checksum if available
	var checksum string
	if !options.SkipChecksum {
		checksum, err = s.github.GetChecksumForAsset(release, asset.Name)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to get checksum, continuing without verification")
		}
	}

	// Download the asset
	log.Info().Str("version", targetVersion).Msg("Downloading update")
	archivePath, err := s.downloader.DownloadAsset(asset, s.progressCallback)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer os.Remove(archivePath)

	// Verify checksum
	if checksum != "" && !options.SkipChecksum {
		log.Info().Msg("Verifying checksum")
		if err := s.downloader.VerifyChecksum(archivePath, checksum); err != nil {
			return fmt.Errorf("checksum verification failed: %w", err)
		}
	}

	// Extract binary
	log.Info().Msg("Extracting update")
	newBinaryPath, err := s.downloader.ExtractBinary(archivePath, s.platform)
	if err != nil {
		return fmt.Errorf("failed to extract binary: %w", err)
	}
	defer os.Remove(newBinaryPath)

	// Verify the new binary works
	if !options.SkipVerify {
		log.Info().Msg("Verifying new binary")
		if err := s.verifyBinary(newBinaryPath); err != nil {
			return fmt.Errorf("binary verification failed: %w", err)
		}
	}

	// Create backup of current binary
	var backupPath string
	if !options.NoBackup {
		backupPath, err = s.createBackup()
		if err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
		log.Info().Str("backup", backupPath).Msg("Created backup of current binary")
	}

	// Replace the binary
	log.Info().Msg("Installing new binary")
	if err := s.replaceBinary(newBinaryPath); err != nil {
		// Try to restore backup
		if backupPath != "" {
			log.Error().Err(err).Msg("Update failed, restoring backup")
			if restoreErr := s.restoreBackup(backupPath); restoreErr != nil {
				log.Error().Err(restoreErr).Msg("Failed to restore backup")
			}
		}
		return fmt.Errorf("failed to install update: %w", err)
	}

	// Clean up backup after successful update
	if backupPath != "" {
		os.Remove(backupPath)
	}

	// Clean up temp files
	s.downloader.CleanupTempFiles()

	log.Info().
		Str("version", targetVersion).
		Msg("Successfully updated SSH Secret Keeper")

	return nil
}

// checkPermissions verifies we have permission to update
func (s *Service) checkPermissions() error {
	// Check if we can write to /usr/local/bin/
	targetPath := "/usr/local/bin/sshsk"
	targetDir := "/usr/local/bin"

	// Try to create a test file in the target directory
	testFile := filepath.Join(targetDir, ".sshsk-update-test")
	file, err := os.Create(testFile)
	if err != nil {
		// Can't write to /usr/local/bin/, need sudo
		return fmt.Errorf("insufficient permissions to update %s: try running with sudo", targetPath)
	}
	file.Close()
	os.Remove(testFile)

	return nil
}

// progressCallback handles download progress reporting
func (s *Service) progressCallback(downloaded, total int64) {
	if total <= 0 {
		return
	}

	percent := int((downloaded * 100) / total)
	mb := float64(downloaded) / (1024 * 1024)
	totalMB := float64(total) / (1024 * 1024)

	// Log progress at intervals
	if percent%10 == 0 || downloaded == total {
		log.Debug().
			Int("percent", percent).
			Str("progress", fmt.Sprintf("%.1f/%.1f MB", mb, totalMB)).
			Msg("Download progress")
	}
}

// verifyBinary checks that the new binary works
func (s *Service) verifyBinary(binaryPath string) error {
	cmd := exec.Command(binaryPath, "version")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("binary verification failed: %w (output: %s)", err, string(output))
	}

	// Check that output contains version info
	outputStr := string(output)
	if !strings.Contains(outputStr, "SSH Secret Keeper") {
		return fmt.Errorf("binary verification failed: unexpected output: %s", outputStr)
	}

	log.Debug().Str("output", strings.TrimSpace(outputStr)).Msg("New binary verified")
	return nil
}

// createBackup creates a backup of the current binary
func (s *Service) createBackup() (string, error) {
	// Always target /usr/local/bin/sshsk for system-wide updates
	targetPath := "/usr/local/bin/sshsk"

	// Check if system binary exists
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		// If /usr/local/bin/sshsk doesn't exist, fall back to current binary
		currentPath, err := GetCurrentBinaryPath()
		if err != nil {
			return "", fmt.Errorf("failed to get current binary path: %w", err)
		}
		targetPath = currentPath
	}

	backupPath := targetPath + ".backup"

	// Remove old backup if it exists
	os.Remove(backupPath)

	// Only create backup if source exists
	if _, err := os.Stat(targetPath); err == nil {
		// Copy current binary to backup
		if err := s.copyFile(targetPath, backupPath); err != nil {
			return "", fmt.Errorf("failed to create backup: %w", err)
		}
	}

	return backupPath, nil
}

// replaceBinary replaces the current binary with the new one
func (s *Service) replaceBinary(newBinaryPath string) error {
	// Always target /usr/local/bin/sshsk for system-wide updates
	targetPath := "/usr/local/bin/sshsk"

	// Check if target exists, if not use current binary path as fallback
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		// If /usr/local/bin/sshsk doesn't exist, fall back to current binary location
		currentPath, err := GetCurrentBinaryPath()
		if err != nil {
			return fmt.Errorf("failed to get current binary path: %w", err)
		}
		targetPath = currentPath
		log.Debug().Str("path", targetPath).Msg("System binary not found, updating current binary")
	} else {
		log.Info().Str("path", targetPath).Msg("Updating system binary")
	}

	// Get target binary permissions (or use default if new installation)
	var fileMode os.FileMode = 0755
	if info, err := os.Stat(targetPath); err == nil {
		fileMode = info.Mode()
	}

	// On both Windows and Unix systems, we need to rename the old binary first
	// Windows: because you can't delete a running executable
	// Linux/Unix: because you get "text file busy" when writing to a running executable
	tempPath := targetPath + ".old"
	os.Remove(tempPath) // Remove if exists

	// Only rename if the target exists
	if _, err := os.Stat(targetPath); err == nil {
		if err := os.Rename(targetPath, tempPath); err != nil {
			return fmt.Errorf("failed to move old binary: %w", err)
		}
		defer os.Remove(tempPath)
	}

	// Copy new binary to target location
	if err := s.copyFile(newBinaryPath, targetPath); err != nil {
		// Try to restore the old binary if copy fails
		if _, err := os.Stat(tempPath); err == nil {
			os.Rename(tempPath, targetPath)
		}
		return fmt.Errorf("failed to copy new binary: %w", err)
	}

	// Preserve permissions
	if err := os.Chmod(targetPath, fileMode); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	return nil
}

// restoreBackup restores the backup binary
func (s *Service) restoreBackup(backupPath string) error {
	// Always target /usr/local/bin/sshsk for system-wide updates
	targetPath := "/usr/local/bin/sshsk"

	// Check if we should restore to system location or current binary
	if _, err := os.Stat(backupPath); err == nil {
		// If backup is for system location, restore there
		if strings.HasPrefix(backupPath, "/usr/local/bin/") {
			targetPath = strings.TrimSuffix(backupPath, ".backup")
		} else {
			// Otherwise use current binary location
			currentPath, err := GetCurrentBinaryPath()
			if err != nil {
				return fmt.Errorf("failed to get current binary path: %w", err)
			}
			targetPath = currentPath
		}
	}

	// On Linux/Unix, we need to handle "text file busy" error
	// First try to rename the current binary if it exists
	if _, err := os.Stat(targetPath); err == nil {
		tempPath := targetPath + ".failed"
		os.Remove(tempPath) // Remove if exists

		// Try to rename the failed binary
		if err := os.Rename(targetPath, tempPath); err != nil {
			// If rename fails, it might be because the binary is still running
			// In this case, we can't restore the backup while the process is running
			return fmt.Errorf("cannot restore backup: current binary is still running: %w", err)
		}
		defer os.Remove(tempPath)
	}

	return s.copyFile(backupPath, targetPath)
}

// copyFile copies a file from src to dst
func (s *Service) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = destFile.ReadFrom(sourceFile)
	return err
}

// GetPlatformInfo returns information about the current platform
func (s *Service) GetPlatformInfo() string {
	return fmt.Sprintf("%s-%s", s.platform.OS, s.platform.Arch)
}

// SetPlatform allows overriding the detected platform (for testing)
func (s *Service) SetPlatform(platform Platform) {
	s.platform = platform
}
