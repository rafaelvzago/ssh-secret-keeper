package ssh

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/rzago/ssh-vault-keeper/internal/analyzer"
	"github.com/rzago/ssh-vault-keeper/internal/crypto"
)

// FileData represents SSH file data with metadata
type FileData struct {
	Filename    string                 `json:"filename"`
	Content     []byte                 `json:"-"`                  // Raw content (not serialized)
	Permissions os.FileMode            `json:"permissions"`
	Size        int64                  `json:"size"`
	ModTime     time.Time              `json:"mod_time"`
	Checksum    string                 `json:"checksum"`
	KeyInfo     *analyzer.KeyInfo      `json:"key_info,omitempty"`
	Encrypted   *crypto.EncryptedData  `json:"encrypted_data,omitempty"`
}

// BackupData represents a complete SSH backup
type BackupData struct {
	Version     string                    `json:"version"`
	Timestamp   time.Time                 `json:"timestamp"`
	Hostname    string                    `json:"hostname"`
	Username    string                    `json:"username"`
	SSHDir      string                    `json:"ssh_dir"`
	Files       map[string]*FileData      `json:"files"`
	Analysis    *analyzer.DetectionResult `json:"analysis"`
	Metadata    map[string]interface{}    `json:"metadata"`
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

	backup := &BackupData{
		Version:   "1.0",
		Timestamp: time.Now(),
		Hostname:  hostname,
		Username:  username,
		SSHDir:    sshDir,
		Files:     files,
		Analysis:  analysis,
		Metadata: map[string]interface{}{
			"total_files":     len(files),
			"total_size":      h.calculateTotalSize(files),
			"key_pair_count":  len(analysis.KeyPairs),
			"service_count":   len(analysis.Categories["service"]),
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

		// Calculate checksum
		checksum := fmt.Sprintf("%x", sha256.Sum256(content))

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
			Str("checksum", checksum[:8]).
			Msg("File read successfully")
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

	// Ensure target directory exists
	if !options.DryRun {
		if err := os.MkdirAll(targetDir, 0700); err != nil {
			return fmt.Errorf("failed to create target directory: %w", err)
		}
	}

	for filename, fileData := range backup.Files {
		// Skip if file filtered out
		if !h.shouldRestoreFile(filename, options) {
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

		// Write file content
		if err := os.WriteFile(targetPath, fileData.Content, fileData.Permissions); err != nil {
			return fmt.Errorf("failed to write file %s: %w", filename, err)
		}

		// Restore modification time
		if err := os.Chtimes(targetPath, time.Now(), fileData.ModTime); err != nil {
			log.Warn().Err(err).Str("file", filename).Msg("Failed to restore modification time")
		}

		log.Info().
			Str("file", filename).
			Str("permissions", fileData.Permissions.String()).
			Msg("File restored successfully")
	}

	log.Info().Msg("File restoration completed")
	return nil
}

// RestoreOptions configure the restoration process
type RestoreOptions struct {
	DryRun      bool
	Overwrite   bool
	Interactive bool
	FileFilter  []string  // Only restore these files
	TypeFilter  []string  // Only restore these file types
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

// VerifyBackup verifies the integrity of a backup
func (h *Handler) VerifyBackup(backup *BackupData) error {
	log.Info().Msg("Verifying backup integrity")

	for filename, fileData := range backup.Files {
		if fileData.Content == nil {
			log.Warn().Str("file", filename).Msg("File has no content to verify")
			continue
		}

		// Verify checksum
		currentChecksum := fmt.Sprintf("%x", sha256.Sum256(fileData.Content))
		if currentChecksum != fileData.Checksum {
			return fmt.Errorf("checksum mismatch for file %s: expected %s, got %s", 
				filename, fileData.Checksum, currentChecksum)
		}

		log.Debug().Str("file", filename).Msg("Checksum verified")
	}

	log.Info().Msg("Backup integrity verified")
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
