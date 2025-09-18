package ssh

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/rzago/ssh-secret-keeper/internal/analyzer"
	"github.com/rzago/ssh-secret-keeper/internal/crypto"
	"github.com/rzago/ssh-secret-keeper/internal/utils"
)

// FileData represents SSH file data with metadata
type FileData struct {
	Filename    string                `json:"filename"`
	Content     []byte                `json:"-"` // Raw content (not serialized)
	Permissions os.FileMode           `json:"permissions"`
	Size        int64                 `json:"size"`
	ModTime     time.Time             `json:"mod_time"`
	Checksum    string                `json:"checksum"`
	KeyInfo     *analyzer.KeyInfo     `json:"key_info,omitempty"`
	Encrypted   *crypto.EncryptedData `json:"encrypted_data,omitempty"`
}

// BackupData represents a complete SSH backup
type BackupData struct {
	Version      string                    `json:"version"`
	Timestamp    time.Time                 `json:"timestamp"`
	Hostname     string                    `json:"hostname"`
	Username     string                    `json:"username"`
	SSHDir       string                    `json:"ssh_dir"`                      // Keep for backward compatibility
	SSHDirNorm   string                    `json:"ssh_dir_normalized,omitempty"` // New normalized path for cross-user compatibility
	OriginalUser string                    `json:"original_user,omitempty"`      // For informational purposes
	PathVersion  string                    `json:"path_version,omitempty"`       // Track path normalization version
	Files        map[string]*FileData      `json:"files"`
	Analysis     *analyzer.DetectionResult `json:"analysis"`
	Metadata     map[string]interface{}    `json:"metadata"`
}

// Handler manages SSH file operations
type Handler struct {
	analyzer  *analyzer.Analyzer
	encryptor *crypto.Encryptor
}

// New creates a new SSH handler
func New() *Handler {
	return &Handler{
		analyzer:  analyzer.New(),
		encryptor: crypto.NewEncryptor(crypto.DefaultIterations),
	}
}

// ReadDirectory reads and analyzes an SSH directory
func (h *Handler) ReadDirectory(sshDir string) (*BackupData, error) {
	log.Info().Str("dir", sshDir).Msg("Reading SSH directory")

	// Verify directory exists
	if _, err := os.Stat(sshDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("SSH directory does not exist: %s", sshDir)
	}

	// Analyze directory structure
	analysis, err := h.analyzer.AnalyzeDirectory(sshDir)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze SSH directory: %w", err)
	}

	// Read file contents
	files, err := h.readFiles(sshDir, analysis.Keys)
	if err != nil {
		return nil, fmt.Errorf("failed to read SSH files: %w", err)
	}

	// Get system information
	hostname, _ := os.Hostname()
	username := os.Getenv("USER")
	if username == "" {
		username = "unknown"
	}

	// Normalize SSH directory path for cross-user compatibility
	pathNormalizer := utils.NewPathNormalizer()
	normalizedSSHDir, err := pathNormalizer.NormalizePath(sshDir)
	if err != nil {
		log.Warn().
			Err(err).
			Str("original_path", sshDir).
			Msg("Failed to normalize SSH directory path, using original")
		normalizedSSHDir = sshDir
	}

	backup := &BackupData{
		Version:      "1.0",
		Timestamp:    time.Now(),
		Hostname:     hostname,
		Username:     username,
		SSHDir:       sshDir,           // Keep original for backward compatibility
		SSHDirNorm:   normalizedSSHDir, // New normalized path
		OriginalUser: username,         // Store original user for reference
		PathVersion:  "2.0",            // Version indicating path normalization support
		Files:        files,
		Analysis:     analysis,
		Metadata: map[string]interface{}{
			"total_files":           len(files),
			"total_size":            h.calculateTotalSize(files),
			"key_pair_count":        len(analysis.KeyPairs),
			"service_count":         len(analysis.Categories["service"]),
			"normalized_path":       normalizedSSHDir,
			"cross_user_compatible": true,
		},
	}

	log.Info().
		Int("files", len(files)).
		Int64("total_size", h.calculateTotalSize(files)).
		Msg("SSH directory read completed")

	return backup, nil
}

// readFiles reads the content of SSH files
func (h *Handler) readFiles(sshDir string, keys []analyzer.KeyInfo) (map[string]*FileData, error) {
	files := make(map[string]*FileData)

	for _, keyInfo := range keys {
		filePath := filepath.Join(sshDir, keyInfo.Filename)

		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Warn().
				Err(err).
				Str("file", keyInfo.Filename).
				Msg("Failed to read file")
			continue
		}

		// Calculate MD5 checksum
		checksum := fmt.Sprintf("%x", md5.Sum(content))

		fileData := &FileData{
			Filename:    keyInfo.Filename,
			Content:     content,
			Permissions: keyInfo.Permissions,
			Size:        keyInfo.Size,
			ModTime:     keyInfo.ModTime,
			Checksum:    checksum,
			KeyInfo:     &keyInfo,
		}

		files[keyInfo.Filename] = fileData

		log.Debug().
			Str("file", keyInfo.Filename).
			Int("size", len(content)).
			Str("md5", checksum[:8]).
			Msg("File read successfully with MD5 checksum")
	}

	return files, nil
}

// EncryptBackup encrypts a backup with the given passphrase
func (h *Handler) EncryptBackup(backup *BackupData, passphrase string) error {
	log.Info().Msg("Encrypting backup data")

	for filename, fileData := range backup.Files {
		if fileData.Content == nil {
			continue
		}

		encrypted, err := h.encryptor.Encrypt(fileData.Content, passphrase)
		if err != nil {
			return fmt.Errorf("failed to encrypt file %s: %w", filename, err)
		}

		fileData.Encrypted = encrypted
		// Clear plaintext content after encryption
		fileData.Content = nil

		log.Debug().
			Str("file", filename).
			Msg("File encrypted")
	}

	log.Info().Int("files", len(backup.Files)).Msg("Backup encryption completed")
	return nil
}

// DecryptBackup decrypts a backup with the given passphrase
func (h *Handler) DecryptBackup(backup *BackupData, passphrase string) error {
	log.Info().Msg("Decrypting backup data")

	for filename, fileData := range backup.Files {
		if fileData.Encrypted == nil {
			continue
		}

		content, err := h.encryptor.Decrypt(fileData.Encrypted, passphrase)
		if err != nil {
			return fmt.Errorf("failed to decrypt file %s: %w", filename, err)
		}

		fileData.Content = content
		// Keep encrypted data for verification if needed

		log.Debug().
			Str("file", filename).
			Msg("File decrypted")
	}

	log.Info().Int("files", len(backup.Files)).Msg("Backup decryption completed")
	return nil
}

// RestoreFiles restores files to the SSH directory
func (h *Handler) RestoreFiles(backup *BackupData, targetDir string, options RestoreOptions) error {
	log.Info().
		Str("target", targetDir).
		Bool("dry_run", options.DryRun).
		Msg("Starting file restoration")

	// Ensure target directory exists with correct SSH directory permissions
	if !options.DryRun {
		if err := os.MkdirAll(targetDir, 0700); err != nil {
			return fmt.Errorf("failed to create target directory: %w", err)
		}

		// Verify SSH directory permissions
		if err := h.verifySSHDirectoryPermissions(targetDir); err != nil {
			log.Warn().Err(err).Msg("SSH directory permission warning")
		}
	}

	for filename, fileData := range backup.Files {
		// Skip if file filtered out
		if !h.shouldRestoreFile(filename, options) {
			log.Debug().Str("file", filename).Msg("File filtered out, skipping")
			continue
		}

		targetPath := filepath.Join(targetDir, filename)

		if options.DryRun {
			log.Info().
				Str("file", filename).
				Str("target", targetPath).
				Str("permissions", fileData.Permissions.String()).
				Msg("[DRY RUN] Would restore file")
			continue
		}

		// Debug: Log file data details
		log.Debug().
			Str("file", filename).
			Int64("content_length", int64(len(fileData.Content))).
			Bool("content_is_nil", fileData.Content == nil).
			Str("target_path", targetPath).
			Bool("overwrite_enabled", options.Overwrite).
			Msg("Processing file for restore")

		// Check if file exists and handle conflicts
		if _, err := os.Stat(targetPath); err == nil && !options.Overwrite {
			if options.Interactive {
				if !h.promptOverwrite(filename) {
					log.Info().Str("file", filename).Msg("Skipped by user")
					continue
				}
			} else {
				log.Warn().Str("file", filename).Msg("File exists, skipping")
				continue
			}
		}

		// Check for empty content
		if fileData.Content == nil {
			log.Error().Str("file", filename).Msg("CRITICAL: File content is nil - cannot restore file")
			continue
		}

		if len(fileData.Content) == 0 {
			log.Warn().Str("file", filename).Msg("File content is empty - creating empty file")
		}

		// Determine appropriate permissions for the file
		permissions := h.getAppropriatePermissions(fileData)

		// Write file content with appropriate permissions
		log.Debug().
			Str("file", filename).
			Str("target_path", targetPath).
			Int("content_bytes", len(fileData.Content)).
			Str("permissions", fmt.Sprintf("%04o", permissions&os.ModePerm)).
			Msg("Writing file to disk")

		if err := os.WriteFile(targetPath, fileData.Content, permissions); err != nil {
			return fmt.Errorf("failed to write file %s: %w", filename, err)
		}

		log.Info().
			Str("file", filename).
			Str("target_path", targetPath).
			Int("bytes_written", len(fileData.Content)).
			Msg("✓ File successfully written to disk")

		// Ensure permissions are set correctly (os.WriteFile only sets permissions for new files)
		if err := os.Chmod(targetPath, permissions); err != nil {
			log.Warn().
				Err(err).
				Str("file", filename).
				Str("permissions", fmt.Sprintf("%04o", permissions&os.ModePerm)).
				Msg("Failed to set file permissions")
		} else {
			log.Debug().
				Str("file", filename).
				Str("permissions", fmt.Sprintf("%04o", permissions&os.ModePerm)).
				Msg("File permissions set successfully")
		}

		// Verify and warn about critical permission issues
		if err := h.validateFilePermissions(filename, permissions, fileData.KeyInfo); err != nil {
			log.Warn().Err(err).Str("file", filename).Msg("Permission validation warning")
		}

		// Restore modification time
		if err := os.Chtimes(targetPath, time.Now(), fileData.ModTime); err != nil {
			log.Warn().Err(err).Str("file", filename).Msg("Failed to restore modification time")
		}

		log.Info().
			Str("file", filename).
			Str("permissions", permissions.String()).
			Str("octal", fmt.Sprintf("%04o", permissions&os.ModePerm)).
			Msg("File restored with appropriate permissions")
	}

	log.Info().Msg("File restoration completed")
	return nil
}

// RestoreOptions configure the restoration process
type RestoreOptions struct {
	DryRun      bool
	Overwrite   bool
	Interactive bool
	FileFilter  []string // Only restore these files
	TypeFilter  []string // Only restore these file types
}

// shouldRestoreFile checks if a file should be restored based on options
func (h *Handler) shouldRestoreFile(filename string, options RestoreOptions) bool {
	// Check file filter
	if len(options.FileFilter) > 0 {
		found := false
		for _, filter := range options.FileFilter {
			if matched, _ := filepath.Match(filter, filename); matched {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Type filtering would go here (based on key info)
	// This is a placeholder for future enhancement

	return true
}

// promptOverwrite prompts the user for file overwrite confirmation
func (h *Handler) promptOverwrite(filename string) bool {
	fmt.Printf("File %s already exists. Overwrite? [y/N]: ", filename)

	var response string
	fmt.Scanln(&response)

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// VerifyBackup verifies the integrity of a backup using MD5 checksums
func (h *Handler) VerifyBackup(backup *BackupData) error {
	log.Info().Msg("Verifying backup integrity with MD5 checksums")

	for filename, fileData := range backup.Files {
		if fileData.Content == nil {
			log.Warn().Str("file", filename).Msg("File has no content to verify")
			continue
		}

		// Verify MD5 checksum
		currentChecksum := fmt.Sprintf("%x", md5.Sum(fileData.Content))
		if currentChecksum != fileData.Checksum {
			return fmt.Errorf("MD5 checksum mismatch for file %s: expected %s, got %s",
				filename, fileData.Checksum, currentChecksum)
		}

		log.Debug().Str("file", filename).Str("md5", currentChecksum[:8]).Msg("MD5 checksum verified")
	}

	log.Info().Msg("Backup integrity verified with MD5 checksums")
	return nil
}

// calculateTotalSize calculates the total size of all files in the backup
func (h *Handler) calculateTotalSize(files map[string]*FileData) int64 {
	var total int64
	for _, fileData := range files {
		total += fileData.Size
	}
	return total
}

// CalculateTotalSize is a public wrapper for calculateTotalSize
func (h *Handler) CalculateTotalSize(files map[string]*FileData) int64 {
	return h.calculateTotalSize(files)
}

// verifySSHDirectoryPermissions checks SSH directory has correct permissions
func (h *Handler) verifySSHDirectoryPermissions(sshDir string) error {
	stat, err := os.Stat(sshDir)
	if err != nil {
		return fmt.Errorf("cannot stat SSH directory: %w", err)
	}

	perm := stat.Mode().Perm()
	if perm != 0700 {
		return fmt.Errorf("SSH directory has incorrect permissions %04o (should be 0700)", perm)
	}

	log.Info().
		Str("directory", sshDir).
		Str("permissions", "0700").
		Msg("SSH directory permissions verified")

	return nil
}

// validateFilePermissions validates SSH file permissions and warns about issues
func (h *Handler) validateFilePermissions(filename string, perms os.FileMode, keyInfo *analyzer.KeyInfo) error {
	perm := perms & os.ModePerm

	// Define expected permissions for different file types
	var expectedPerms []os.FileMode
	var fileType string

	if keyInfo != nil {
		switch keyInfo.Type {
		case analyzer.KeyTypePrivate:
			expectedPerms = []os.FileMode{0600}
			fileType = "private key"
		case analyzer.KeyTypePublic:
			expectedPerms = []os.FileMode{0644, 0600}
			fileType = "public key"
		case analyzer.KeyTypeConfig:
			expectedPerms = []os.FileMode{0600, 0644}
			fileType = "SSH config"
		case analyzer.KeyTypeHosts:
			expectedPerms = []os.FileMode{0600, 0644}
			fileType = "known_hosts"
		case analyzer.KeyTypeAuthorized:
			expectedPerms = []os.FileMode{0600, 0644}
			fileType = "authorized_keys"
		default:
			expectedPerms = []os.FileMode{0600, 0644}
			fileType = "SSH file"
		}
	} else {
		// Fallback for files without key info
		expectedPerms = []os.FileMode{0600, 0644}
		fileType = "SSH file"
	}

	// Check if current permissions are acceptable
	permissionOK := false
	for _, expectedPerm := range expectedPerms {
		if perm == expectedPerm {
			permissionOK = true
			break
		}
	}

	if !permissionOK {
		// Log warning for problematic permissions
		log.Warn().
			Str("file", filename).
			Str("type", fileType).
			Str("current_perms", fmt.Sprintf("%04o", perm)).
			Strs("recommended_perms", permissionsToStrings(expectedPerms)).
			Msg("File permissions may be insecure")

		// Special warning for overly permissive private keys
		if keyInfo != nil && keyInfo.Type == analyzer.KeyTypePrivate && (perm&0077) != 0 {
			return fmt.Errorf("CRITICAL: Private key %s has world/group readable permissions (%04o) - SSH will reject this key",
				filename, perm)
		}
	}

	return nil
}

// permissionsToStrings converts permission modes to string representations
func permissionsToStrings(perms []os.FileMode) []string {
	result := make([]string, len(perms))
	for i, perm := range perms {
		result[i] = fmt.Sprintf("%04o", perm)
	}
	return result
}

// VerifyRestorePermissions performs post-restore permission verification
func (h *Handler) VerifyRestorePermissions(backup *BackupData, targetDir string) error {
	log.Info().Str("target", targetDir).Msg("Verifying restored file permissions")

	permissionIssues := 0

	// Check SSH directory itself
	if err := h.verifySSHDirectoryPermissions(targetDir); err != nil {
		log.Error().Err(err).Msg("SSH directory permission issue")
		permissionIssues++
	}

	// Check each restored file
	for filename, fileData := range backup.Files {
		targetPath := filepath.Join(targetDir, filename)

		// Check if file exists
		stat, err := os.Stat(targetPath)
		if err != nil {
			log.Warn().Err(err).Str("file", filename).Msg("Cannot verify file permissions (file not found)")
			continue
		}

		actualPerms := stat.Mode().Perm()
		expectedPerms := h.getAppropriatePermissions(fileData) & os.ModePerm

		if actualPerms != expectedPerms {
			log.Error().
				Str("file", filename).
				Str("expected", fmt.Sprintf("%04o", expectedPerms)).
				Str("actual", fmt.Sprintf("%04o", actualPerms)).
				Msg("Permission mismatch after restore")
			permissionIssues++
		} else {
			log.Debug().
				Str("file", filename).
				Str("permissions", fmt.Sprintf("%04o", actualPerms)).
				Msg("File permissions verified")
		}

		// Validate security of permissions
		if err := h.validateFilePermissions(filename, stat.Mode(), fileData.KeyInfo); err != nil {
			log.Warn().Err(err).Str("file", filename).Msg("Permission security warning")
		}
	}

	if permissionIssues > 0 {
		return fmt.Errorf("found %d permission issues after restore", permissionIssues)
	}

	log.Info().Msg("All file permissions verified successfully")
	return nil
}

// getAppropriatePermissions determines the correct permissions for a file
func (h *Handler) getAppropriatePermissions(fileData *FileData) os.FileMode {
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
