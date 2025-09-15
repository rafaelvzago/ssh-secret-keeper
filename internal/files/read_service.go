package files

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/rzago/ssh-vault-keeper/internal/analyzer"
	"github.com/rzago/ssh-vault-keeper/internal/ssh"
)

// ReadService provides file reading functionality following SRP
type ReadService struct{}

// NewReadService creates a new file read service
func NewReadService() *ReadService {
	return &ReadService{}
}

// ReadSSHDirectory reads all files from an SSH directory
func (s *ReadService) ReadSSHDirectory(sshDir string) (map[string]*ssh.FileData, error) {
	if err := s.ValidateDirectory(sshDir); err != nil {
		return nil, fmt.Errorf("directory validation failed: %w", err)
	}

	files, err := os.ReadDir(sshDir)
	if err != nil {
		return nil, fmt.Errorf("cannot read directory: %w", err)
	}

	fileData := make(map[string]*ssh.FileData)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(sshDir, file.Name())
		data, err := s.readSingleFile(filePath)
		if err != nil {
			log.Warn().
				Err(err).
				Str("file", file.Name()).
				Msg("Failed to read file, skipping")
			continue
		}

		fileData[file.Name()] = data
	}

	log.Info().
		Int("files_read", len(fileData)).
		Str("directory", sshDir).
		Msg("SSH directory read completed")

	return fileData, nil
}

// ValidateDirectory validates that a directory exists and is accessible
func (s *ReadService) ValidateDirectory(sshDir string) error {
	if sshDir == "" {
		return fmt.Errorf("directory path cannot be empty")
	}

	// Clean and resolve path
	cleanDir := filepath.Clean(sshDir)

	// Check if directory exists
	stat, err := os.Stat(cleanDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory does not exist: %s", cleanDir)
		}
		return fmt.Errorf("cannot access directory: %w", err)
	}

	// Verify it's actually a directory
	if !stat.IsDir() {
		return fmt.Errorf("path is not a directory: %s", cleanDir)
	}

	// Test read permissions
	f, err := os.Open(cleanDir)
	if err != nil {
		return fmt.Errorf("cannot read directory: %w", err)
	}
	f.Close()

	log.Debug().
		Str("directory", cleanDir).
		Str("permissions", stat.Mode().Perm().String()).
		Msg("Directory validated")

	return nil
}

// CalculateChecksum calculates SHA-256 checksum for data
func (s *ReadService) CalculateChecksum(data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}

// VerifyFilePermissions checks if file has expected permissions
func (s *ReadService) VerifyFilePermissions(filePath string, expectedMode os.FileMode) error {
	stat, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("cannot stat file: %w", err)
	}

	actualMode := stat.Mode().Perm()
	expectedPerm := expectedMode.Perm()

	if actualMode != expectedPerm {
		return fmt.Errorf("file %s has permissions %04o, expected %04o",
			filePath, actualMode, expectedPerm)
	}

	return nil
}

// Private helper methods

func (s *ReadService) readSingleFile(filePath string) (*ssh.FileData, error) {
	// Get file information first
	stat, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot stat file: %w", err)
	}

	// Check file size (warn if very large)
	const maxRecommendedSize = 10 * 1024 * 1024 // 10MB
	if stat.Size() > maxRecommendedSize {
		log.Warn().
			Str("file", filepath.Base(filePath)).
			Int64("size", stat.Size()).
			Msg("File is unusually large for SSH directory")
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot read file content: %w", err)
	}

	// Calculate checksum
	checksum := s.CalculateChecksum(content)

	// Create file data structure
	fileData := &ssh.FileData{
		Filename:    filepath.Base(filePath),
		Content:     content,
		Permissions: stat.Mode(),
		Size:        stat.Size(),
		ModTime:     stat.ModTime(),
		Checksum:    checksum,
		// KeyInfo will be populated by the analyzer
	}

	log.Debug().
		Str("file", fileData.Filename).
		Int("size", len(content)).
		Str("checksum", checksum[:8]+"...").
		Str("permissions", stat.Mode().Perm().String()).
		Msg("File read successfully")

	return fileData, nil
}

// AddKeyInfoToFiles adds key analysis information to file data
func (s *ReadService) AddKeyInfoToFiles(files map[string]*ssh.FileData, analysisResult *analyzer.DetectionResult) {
	// Create a map for quick lookup of key info by filename
	keyInfoMap := make(map[string]*analyzer.KeyInfo)
	for i := range analysisResult.Keys {
		key := &analysisResult.Keys[i]
		keyInfoMap[key.Filename] = key
	}

	// Add key info to corresponding file data
	for filename, fileData := range files {
		if keyInfo, exists := keyInfoMap[filename]; exists {
			fileData.KeyInfo = keyInfo
		}
	}

	log.Debug().
		Int("files_with_keyinfo", len(keyInfoMap)).
		Int("total_files", len(files)).
		Msg("Key info added to file data")
}

// ValidateFileIntegrity verifies file content matches stored checksum
func (s *ReadService) ValidateFileIntegrity(fileData *ssh.FileData) error {
	if fileData == nil {
		return fmt.Errorf("file data is nil")
	}

	if fileData.Content == nil {
		return fmt.Errorf("file content is nil")
	}

	currentChecksum := s.CalculateChecksum(fileData.Content)
	if currentChecksum != fileData.Checksum {
		return fmt.Errorf("checksum mismatch for file %s: expected %s, got %s",
			fileData.Filename, fileData.Checksum, currentChecksum)
	}

	return nil
}

// GetFileStats returns statistics about files in a directory
func (s *ReadService) GetFileStats(files map[string]*ssh.FileData) map[string]interface{} {
	stats := make(map[string]interface{})

	var totalSize int64
	var oldestFile, newestFile time.Time
	permissionCounts := make(map[string]int)

	first := true
	for _, fileData := range files {
		totalSize += fileData.Size

		if first {
			oldestFile = fileData.ModTime
			newestFile = fileData.ModTime
			first = false
		} else {
			if fileData.ModTime.Before(oldestFile) {
				oldestFile = fileData.ModTime
			}
			if fileData.ModTime.After(newestFile) {
				newestFile = fileData.ModTime
			}
		}

		permStr := fmt.Sprintf("%04o", fileData.Permissions&os.ModePerm)
		permissionCounts[permStr]++
	}

	stats["file_count"] = len(files)
	stats["total_size"] = totalSize
	stats["average_size"] = float64(totalSize) / float64(len(files))
	stats["oldest_file"] = oldestFile.Format(time.RFC3339)
	stats["newest_file"] = newestFile.Format(time.RFC3339)
	stats["permission_counts"] = permissionCounts

	return stats
}
