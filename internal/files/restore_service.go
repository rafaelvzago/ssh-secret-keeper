package files

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/rzago/ssh-secret-keeper/internal/analyzer"
	"github.com/rzago/ssh-secret-keeper/internal/ssh"
)

// RestoreService provides file restoration functionality following SRP
type RestoreService struct{}

// NewRestoreService creates a new file restore service
func NewRestoreService() *RestoreService {
	return &RestoreService{}
}

// RestoreFiles restores files from backup to target directory
func (s *RestoreService) RestoreFiles(backup *ssh.BackupData, targetDir string, options ssh.RestoreOptions) error {
	if backup == nil {
		return fmt.Errorf("backup data is nil")
	}

	if err := s.ValidateRestoreTarget(targetDir); err != nil {
		return fmt.Errorf("target validation failed: %w", err)
	}

	log.Info().
		Str("target", targetDir).
		Bool("dry_run", options.DryRun).
		Int("total_files", len(backup.Files)).
		Msg("Starting file restoration")

	// Create SSH directory if it doesn't exist
	if !options.DryRun {
		if err := s.CreateSSHDirectory(targetDir); err != nil {
			return fmt.Errorf("failed to create SSH directory: %w", err)
		}
	}

	restoredCount := 0
	skippedFiles := make([]string, 0)

	for filename, fileData := range backup.Files {
		// Check if file should be restored based on filters
		if !s.shouldRestoreFile(filename, fileData, options) {
			skippedFiles = append(skippedFiles, filename)
			continue
		}

		targetPath := filepath.Join(targetDir, filename)

		if options.DryRun {
			s.logDryRunRestore(filename, targetPath, fileData)
			restoredCount++
			continue
		}

		// Handle existing files
		if s.fileExists(targetPath) {
			action, err := s.handleExistingFile(filename, targetPath, options)
			if err != nil {
				return fmt.Errorf("error handling existing file %s: %w", filename, err)
			}

			if action == "skip" {
				skippedFiles = append(skippedFiles, filename)
				continue
			}
		}

		// Restore the file
		if err := s.restoreSingleFile(fileData, targetPath); err != nil {
			return fmt.Errorf("failed to restore file %s: %w", filename, err)
		}

		restoredCount++

		log.Info().
			Str("file", filename).
			Str("target", targetPath).
			Str("permissions", fmt.Sprintf("%04o", s.getAppropriatePermissions(fileData)&os.ModePerm)).
			Msg("File restored successfully")
	}

	log.Info().
		Int("restored", restoredCount).
		Int("skipped", len(skippedFiles)).
		Msg("File restoration completed")

	return nil
}

// ValidateRestoreTarget validates the target directory for restoration
func (s *RestoreService) ValidateRestoreTarget(targetDir string) error {
	if targetDir == "" {
		return fmt.Errorf("target directory cannot be empty")
	}

	// Clean the path
	cleanDir := filepath.Clean(targetDir)

	// Expand home directory if needed
	if strings.HasPrefix(targetDir, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("cannot resolve home directory: %w", err)
		}
		cleanDir = filepath.Join(homeDir, targetDir[2:])
	}

	// Check if target already exists
	if stat, err := os.Stat(cleanDir); err == nil {
		if !stat.IsDir() {
			return fmt.Errorf("target path exists but is not a directory: %s", cleanDir)
		}
		log.Debug().
			Str("target", cleanDir).
			Msg("Target directory already exists")
		return nil
	}

	// Check if parent directory exists and is writable
	parentDir := filepath.Dir(cleanDir)
	if _, err := os.Stat(parentDir); err != nil {
		return fmt.Errorf("parent directory does not exist: %s", parentDir)
	}

	// Test write permission by creating a temp file
	testFile := filepath.Join(parentDir, ".ssh-secret-keeper-test")
	f, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("cannot write to parent directory: %w", err)
	}
	f.Close()
	os.Remove(testFile)

	return nil
}

// CreateSSHDirectory creates SSH directory with correct permissions
func (s *RestoreService) CreateSSHDirectory(targetDir string) error {
	// Clean and resolve path
	cleanDir := filepath.Clean(targetDir)
	if strings.HasPrefix(targetDir, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("cannot resolve home directory: %w", err)
		}
		cleanDir = filepath.Join(homeDir, targetDir[2:])
	}

	// Create directory with SSH-appropriate permissions (700)
	if err := os.MkdirAll(cleanDir, 0700); err != nil {
		return fmt.Errorf("cannot create directory: %w", err)
	}

	// Verify permissions were set correctly
	stat, err := os.Stat(cleanDir)
	if err != nil {
		return fmt.Errorf("cannot stat created directory: %w", err)
	}

	if perm := stat.Mode().Perm(); perm != 0700 {
		// Try to fix permissions
		if err := os.Chmod(cleanDir, 0700); err != nil {
			log.Warn().
				Str("directory", cleanDir).
				Str("actual_perm", fmt.Sprintf("%04o", perm)).
				Msg("SSH directory has incorrect permissions, but chmod failed")
		} else {
			log.Info().
				Str("directory", cleanDir).
				Msg("Fixed SSH directory permissions to 0700")
		}
	}

	log.Info().
		Str("directory", cleanDir).
		Str("permissions", "0700").
		Msg("SSH directory created")

	return nil
}

// VerifyRestorePermissions verifies that restored files have correct permissions
func (s *RestoreService) VerifyRestorePermissions(backup *ssh.BackupData, targetDir string) error {
	log.Info().
		Str("target", targetDir).
		Msg("Verifying restored file permissions")

	issues := 0
	warnings := 0

	// Verify SSH directory permissions
	if err := s.verifySSHDirectoryPermissions(targetDir); err != nil {
		log.Error().Err(err).Msg("SSH directory permission issue")
		issues++
	}

	// Verify each restored file
	for filename, fileData := range backup.Files {
		targetPath := filepath.Join(targetDir, filename)

		if err := s.verifyFilePermissions(targetPath, fileData); err != nil {
			if s.isCriticalPermissionError(err, fileData) {
				log.Error().Err(err).Str("file", filename).Msg("Critical permission error")
				issues++
			} else {
				log.Warn().Err(err).Str("file", filename).Msg("Permission warning")
				warnings++
			}
		}
	}

	if issues > 0 {
		return fmt.Errorf("found %d critical permission issues after restore", issues)
	}

	if warnings > 0 {
		log.Warn().
			Int("warnings", warnings).
			Msg("Permission verification completed with warnings")
	} else {
		log.Info().Msg("All file permissions verified successfully")
	}

	return nil
}

// Private helper methods

func (s *RestoreService) shouldRestoreFile(filename string, fileData *ssh.FileData, options ssh.RestoreOptions) bool {
	// Check file filter
	if len(options.FileFilter) > 0 {
		matched := false
		for _, pattern := range options.FileFilter {
			if match, _ := filepath.Match(pattern, filename); match {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Check type filter
	if len(options.TypeFilter) > 0 && fileData.KeyInfo != nil {
		matched := false
		keyType := string(fileData.KeyInfo.Type)
		for _, filterType := range options.TypeFilter {
			if keyType == filterType {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	return true
}

func (s *RestoreService) logDryRunRestore(filename, targetPath string, fileData *ssh.FileData) {
	log.Info().
		Str("file", filename).
		Str("target", targetPath).
		Str("permissions", fmt.Sprintf("%04o", s.getAppropriatePermissions(fileData)&os.ModePerm)).
		Int64("size", fileData.Size).
		Msg("[DRY RUN] Would restore file")
}

func (s *RestoreService) fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (s *RestoreService) handleExistingFile(filename, targetPath string, options ssh.RestoreOptions) (string, error) {
	if options.Overwrite {
		return "overwrite", nil
	}

	if options.Interactive {
		return s.promptOverwrite(filename)
	}

	// Default: skip existing files
	log.Info().
		Str("file", filename).
		Msg("File exists, skipping (use --overwrite to replace)")
	return "skip", nil
}

func (s *RestoreService) promptOverwrite(filename string) (string, error) {
	fmt.Printf("File %s already exists. Overwrite? [y/N]: ", filename)

	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		return "skip", err
	}

	response = strings.ToLower(strings.TrimSpace(response))
	if response == "y" || response == "yes" {
		return "overwrite", nil
	}

	return "skip", nil
}

func (s *RestoreService) restoreSingleFile(fileData *ssh.FileData, targetPath string) error {
	if fileData.Content == nil {
		return fmt.Errorf("file content is nil")
	}

	// Check if file is empty and warn
	if len(fileData.Content) == 0 {
		log.Warn().
			Str("file", fileData.Filename).
			Msg("Restoring empty file - this may indicate a backup issue")
	}

	// Determine appropriate permissions for the file
	permissions := s.getAppropriatePermissions(fileData)

	// If file exists, fix its permissions first to allow overwriting
	if s.fileExists(targetPath) {
		log.Info().
			Str("file", filepath.Base(targetPath)).
			Msg("Fixing permissions of existing file before overwrite")

		// Use 0666 to make file writable, then set correct permissions later
		if err := os.Chmod(targetPath, 0666); err != nil {
			log.Warn().
				Err(err).
				Str("file", filepath.Base(targetPath)).
				Msg("Failed to fix permissions of existing file before overwrite")
		} else {
			log.Info().
				Str("file", filepath.Base(targetPath)).
				Msg("Successfully fixed permissions of existing file")
		}
	}

	// Write file with appropriate permissions
	if err := os.WriteFile(targetPath, fileData.Content, permissions); err != nil {
		return fmt.Errorf("cannot write file: %w", err)
	}

	// Ensure permissions are set correctly (os.WriteFile only sets permissions for new files)
	log.Debug().
		Str("file", filepath.Base(targetPath)).
		Str("permissions", fmt.Sprintf("%04o", permissions&os.ModePerm)).
		Msg("Setting file permissions")

	if err := os.Chmod(targetPath, permissions); err != nil {
		log.Warn().
			Err(err).
			Str("file", filepath.Base(targetPath)).
			Str("permissions", fmt.Sprintf("%04o", permissions&os.ModePerm)).
			Msg("Failed to set file permissions")
	} else {
		// Verify the permissions were set correctly
		if stat, err := os.Stat(targetPath); err == nil {
			actualPerms := stat.Mode().Perm()
			expectedPerms := permissions & os.ModePerm
			if actualPerms != expectedPerms {
				log.Error().
					Str("file", filepath.Base(targetPath)).
					Str("expected", fmt.Sprintf("%04o", expectedPerms)).
					Str("actual", fmt.Sprintf("%04o", actualPerms)).
					Msg("Permission verification failed after restore")
			} else {
				log.Debug().
					Str("file", filepath.Base(targetPath)).
					Str("permissions", fmt.Sprintf("%04o", actualPerms)).
					Msg("File permissions verified successfully")
			}
		} else {
			log.Warn().
				Err(err).
				Str("file", filepath.Base(targetPath)).
				Msg("Could not verify file permissions after restore")
		}
	}

	// Restore modification time
	if err := os.Chtimes(targetPath, time.Now(), fileData.ModTime); err != nil {
		log.Warn().
			Err(err).
			Str("file", filepath.Base(targetPath)).
			Msg("Failed to restore modification time")
	}

	return nil
}

// getAppropriatePermissions determines the correct permissions for a file
func (s *RestoreService) getAppropriatePermissions(fileData *ssh.FileData) os.FileMode {
	originalPerms := fileData.Permissions & os.ModePerm

	// ALWAYS use original permissions - they should never be 0000 if backup was created correctly
	if originalPerms == 0000 {
		log.Error().
			Str("file", fileData.Filename).
			Str("raw_permissions", fmt.Sprintf("%04o", fileData.Permissions)).
			Msg("CRITICAL: Original permissions are 0000 - this indicates a backup corruption issue")
		// This should not happen if backup was created correctly
		// Return 0600 as a fallback, but log this as an error
		return 0600
	}

	// Use original permissions exactly as they were
	log.Debug().
		Str("file", fileData.Filename).
		Str("permissions", fmt.Sprintf("%04o", originalPerms)).
		Msg("Using original permissions from backup")

	return originalPerms
}

func (s *RestoreService) verifySSHDirectoryPermissions(sshDir string) error {
	stat, err := os.Stat(sshDir)
	if err != nil {
		return fmt.Errorf("cannot stat SSH directory: %w", err)
	}

	perm := stat.Mode().Perm()
	if perm != 0700 {
		return fmt.Errorf("SSH directory has incorrect permissions %04o (should be 0700)", perm)
	}

	return nil
}

func (s *RestoreService) verifyFilePermissions(targetPath string, fileData *ssh.FileData) error {
	stat, err := os.Stat(targetPath)
	if err != nil {
		return fmt.Errorf("cannot stat file: %w", err)
	}

	actualPerm := stat.Mode().Perm()
	expectedPerm := s.getAppropriatePermissions(fileData) & os.ModePerm

	if actualPerm != expectedPerm {
		return fmt.Errorf("permission mismatch: expected %04o, got %04o",
			expectedPerm, actualPerm)
	}

	return nil
}

func (s *RestoreService) isCriticalPermissionError(err error, fileData *ssh.FileData) bool {
	if fileData.KeyInfo == nil {
		return false
	}

	// Private keys with world/group readable permissions are critical
	if fileData.KeyInfo.Type == analyzer.KeyTypePrivate {
		actualPerm := s.getAppropriatePermissions(fileData) & os.ModePerm
		if (actualPerm & 0077) != 0 {
			return true
		}
	}

	return false
}
