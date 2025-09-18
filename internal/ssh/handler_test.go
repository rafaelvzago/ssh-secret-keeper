package ssh

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rzago/ssh-secret-keeper/internal/analyzer"
)

func TestNew(t *testing.T) {
	handler := New()
	if handler == nil {
		t.Fatal("New() returned nil")
	}

	if handler.analyzer == nil {
		t.Error("Handler analyzer is nil")
	}

	if handler.encryptor == nil {
		t.Error("Handler encryptor is nil")
	}
}

func TestHandler_ReadDirectory(t *testing.T) {
	handler := New()

	t.Run("nonexistent directory", func(t *testing.T) {
		_, err := handler.ReadDirectory("/nonexistent/directory")
		if err == nil {
			t.Error("ReadDirectory() should fail for nonexistent directory")
		}
	})

	t.Run("empty directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		backup, err := handler.ReadDirectory(tmpDir)
		if err != nil {
			t.Errorf("ReadDirectory() failed for empty directory: %v", err)
		}

		if backup == nil {
			t.Fatal("ReadDirectory() returned nil backup")
		}

		if len(backup.Files) != 0 {
			t.Errorf("Expected 0 files in empty directory, got %d", len(backup.Files))
		}

		if backup.Version != "1.0" {
			t.Errorf("Expected version 1.0, got %s", backup.Version)
		}
	})

	t.Run("directory with SSH files", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create test SSH files
		testFiles := map[string][]byte{
			"id_rsa": []byte(`-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAFwAAAAdzc2gtcn
NhAAAAAwEAAQAAAQEA1234567890
-----END OPENSSH PRIVATE KEY-----`),
			"id_rsa.pub": []byte("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDT1234567890 user@host"),
			"config": []byte(`Host github.com
    HostName github.com
    User git
    IdentityFile ~/.ssh/id_rsa`),
		}

		for filename, content := range testFiles {
			filePath := filepath.Join(tmpDir, filename)
			err := os.WriteFile(filePath, content, 0600)
			if err != nil {
				t.Fatalf("Failed to create test file %s: %v", filename, err)
			}
		}

		backup, err := handler.ReadDirectory(tmpDir)
		if err != nil {
			t.Errorf("ReadDirectory() failed: %v", err)
		}

		if backup == nil {
			t.Fatal("ReadDirectory() returned nil backup")
		}

		// Check that files were read
		if len(backup.Files) == 0 {
			t.Error("No files were read from SSH directory")
		}

		// Verify backup metadata
		if backup.SSHDir != tmpDir {
			t.Errorf("Backup SSH dir = %s, want %s", backup.SSHDir, tmpDir)
		}

		if backup.Analysis == nil {
			t.Error("Backup analysis is nil")
		}

		// Check metadata
		if backup.Metadata == nil {
			t.Error("Backup metadata is nil")
		} else {
			if totalFiles, ok := backup.Metadata["total_files"].(int); ok {
				if totalFiles != len(backup.Files) {
					t.Errorf("Metadata total_files = %d, want %d", totalFiles, len(backup.Files))
				}
			}
		}

		// Verify file data integrity
		for filename, fileData := range backup.Files {
			if fileData.Filename != filename {
				t.Errorf("File %s has wrong filename in data: %s", filename, fileData.Filename)
			}

			if len(fileData.Content) == 0 {
				t.Errorf("File %s has no content", filename)
			}

			if fileData.Checksum == "" {
				t.Errorf("File %s has no checksum", filename)
			}

			if len(fileData.Checksum) != 32 {
				t.Errorf("File %s checksum length = %d, want 32", filename, len(fileData.Checksum))
			}

			if fileData.Size == 0 {
				t.Errorf("File %s has zero size", filename)
			}
		}
	})
}

func TestHandler_EncryptDecryptBackup(t *testing.T) {
	handler := New()
	passphrase := "test-passphrase-123"

	// Create test backup
	backup := &BackupData{
		Version:   "1.0",
		Timestamp: time.Now(),
		Files: map[string]*FileData{
			"test_key": {
				Filename: "test_key",
				Content:  []byte("private key content"),
				Size:     19,
			},
			"test_key.pub": {
				Filename: "test_key.pub",
				Content:  []byte("public key content"),
				Size:     18,
			},
		},
	}

	// Store original content for verification
	originalContent := make(map[string][]byte)
	for filename, fileData := range backup.Files {
		originalContent[filename] = make([]byte, len(fileData.Content))
		copy(originalContent[filename], fileData.Content)
	}

	t.Run("encrypt backup", func(t *testing.T) {
		err := handler.EncryptBackup(backup, passphrase)
		if err != nil {
			t.Errorf("EncryptBackup() failed: %v", err)
		}

		// Verify encryption
		for filename, fileData := range backup.Files {
			if fileData.Encrypted == nil {
				t.Errorf("File %s was not encrypted", filename)
			}

			if fileData.Content != nil {
				t.Errorf("File %s still has plaintext content after encryption", filename)
			}
		}
	})

	t.Run("decrypt backup", func(t *testing.T) {
		err := handler.DecryptBackup(backup, passphrase)
		if err != nil {
			t.Errorf("DecryptBackup() failed: %v", err)
		}

		// Verify decryption
		for filename, fileData := range backup.Files {
			if fileData.Content == nil {
				t.Errorf("File %s has no content after decryption", filename)
			}

			// Verify content matches original
			if string(fileData.Content) != string(originalContent[filename]) {
				t.Errorf("File %s content mismatch after decrypt", filename)
			}
		}
	})

	t.Run("decrypt with wrong passphrase", func(t *testing.T) {
		// Encrypt again
		err := handler.EncryptBackup(backup, passphrase)
		if err != nil {
			t.Fatalf("EncryptBackup() failed: %v", err)
		}

		// Try to decrypt with wrong passphrase
		err = handler.DecryptBackup(backup, "wrong-passphrase")
		if err == nil {
			t.Error("DecryptBackup() should fail with wrong passphrase")
		}
	})
}

func TestHandler_VerifyBackup(t *testing.T) {
	handler := New()

	t.Run("valid backup", func(t *testing.T) {
		content := []byte("test content")
		checksum := "9473fdd0d880a43c21b7778d34872157" // MD5 of "test content"

		backup := &BackupData{
			Files: map[string]*FileData{
				"test_file": {
					Filename: "test_file",
					Content:  content,
					Checksum: checksum,
				},
			},
		}

		err := handler.VerifyBackup(backup)
		if err != nil {
			t.Errorf("VerifyBackup() failed for valid backup: %v", err)
		}
	})

	t.Run("corrupted backup", func(t *testing.T) {
		backup := &BackupData{
			Files: map[string]*FileData{
				"test_file": {
					Filename: "test_file",
					Content:  []byte("test content"),
					Checksum: "wrong-checksum",
				},
			},
		}

		err := handler.VerifyBackup(backup)
		if err == nil {
			t.Error("VerifyBackup() should fail for corrupted backup")
		}
	})

	t.Run("backup with nil content", func(t *testing.T) {
		backup := &BackupData{
			Files: map[string]*FileData{
				"test_file": {
					Filename: "test_file",
					Content:  nil,
					Checksum: "some-checksum",
				},
			},
		}

		err := handler.VerifyBackup(backup)
		if err != nil {
			t.Errorf("VerifyBackup() should not fail for nil content: %v", err)
		}
	})
}

func TestHandler_RestoreFiles_DryRun(t *testing.T) {
	handler := New()
	tmpDir := t.TempDir()

	backup := &BackupData{
		Files: map[string]*FileData{
			"test_key": {
				Filename:    "test_key",
				Content:     []byte("key content"),
				Permissions: 0600,
			},
		},
	}

	options := RestoreOptions{
		DryRun: true,
	}

	err := handler.RestoreFiles(backup, tmpDir, options)
	if err != nil {
		t.Errorf("RestoreFiles() dry run failed: %v", err)
	}

	// Verify no files were created
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("Dry run should not create files, found %d", len(files))
	}
}

func TestHandler_RestoreFiles_Actual(t *testing.T) {
	handler := New()
	tmpDir := t.TempDir()

	testContent := []byte("key content")
	backup := &BackupData{
		Files: map[string]*FileData{
			"test_key": {
				Filename:    "test_key",
				Content:     testContent,
				Permissions: 0600,
				Size:        int64(len(testContent)),
				ModTime:     time.Now(),
			},
		},
	}

	options := RestoreOptions{
		DryRun: false,
	}

	err := handler.RestoreFiles(backup, tmpDir, options)
	if err != nil {
		t.Errorf("RestoreFiles() failed: %v", err)
	}

	// Verify file was created
	filePath := filepath.Join(tmpDir, "test_key")
	stat, err := os.Stat(filePath)
	if err != nil {
		t.Errorf("Restored file not found: %v", err)
	}

	// Check permissions
	if perm := stat.Mode().Perm(); perm != 0600 {
		t.Errorf("File permissions = %04o, want 0600", perm)
	}

	// Check content
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Errorf("Failed to read restored file: %v", err)
	}

	if string(content) != string(testContent) {
		t.Error("Restored file content mismatch")
	}
}

func TestHandler_RestoreFiles_WithFilters(t *testing.T) {
	handler := New()
	tmpDir := t.TempDir()

	backup := &BackupData{
		Files: map[string]*FileData{
			"id_rsa": {
				Filename:    "id_rsa",
				Content:     []byte("private key"),
				Permissions: 0600,
			},
			"id_rsa.pub": {
				Filename:    "id_rsa.pub",
				Content:     []byte("public key"),
				Permissions: 0644,
			},
			"config": {
				Filename:    "config",
				Content:     []byte("ssh config"),
				Permissions: 0644,
			},
		},
	}

	t.Run("file filter", func(t *testing.T) {
		testDir := filepath.Join(tmpDir, "file_filter")
		os.MkdirAll(testDir, 0755)

		options := RestoreOptions{
			FileFilter: []string{"id_rsa*"},
		}

		err := handler.RestoreFiles(backup, testDir, options)
		if err != nil {
			t.Errorf("RestoreFiles() with file filter failed: %v", err)
		}

		// Check that only matching files were restored
		if _, err := os.Stat(filepath.Join(testDir, "id_rsa")); err != nil {
			t.Error("id_rsa should be restored")
		}
		if _, err := os.Stat(filepath.Join(testDir, "id_rsa.pub")); err != nil {
			t.Error("id_rsa.pub should be restored")
		}
		if _, err := os.Stat(filepath.Join(testDir, "config")); err == nil {
			t.Error("config should not be restored")
		}
	})
}

func TestHandler_RestoreFiles_ExistingFiles(t *testing.T) {
	handler := New()
	tmpDir := t.TempDir()

	// Create existing file
	existingFile := filepath.Join(tmpDir, "test_key")
	err := os.WriteFile(existingFile, []byte("existing content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create existing file: %v", err)
	}

	backup := &BackupData{
		Files: map[string]*FileData{
			"test_key": {
				Filename:    "test_key",
				Content:     []byte("new content"),
				Permissions: 0600,
			},
		},
	}

	t.Run("skip existing files", func(t *testing.T) {
		options := RestoreOptions{
			Overwrite: false,
		}

		err := handler.RestoreFiles(backup, tmpDir, options)
		if err != nil {
			t.Errorf("RestoreFiles() failed: %v", err)
		}

		// File should still have original content
		content, err := os.ReadFile(existingFile)
		if err != nil {
			t.Errorf("Failed to read file: %v", err)
		}

		if string(content) != "existing content" {
			t.Error("File should not be overwritten")
		}
	})

	t.Run("overwrite existing files", func(t *testing.T) {
		options := RestoreOptions{
			Overwrite: true,
		}

		err := handler.RestoreFiles(backup, tmpDir, options)
		if err != nil {
			t.Errorf("RestoreFiles() failed: %v", err)
		}

		// File should have new content
		content, err := os.ReadFile(existingFile)
		if err != nil {
			t.Errorf("Failed to read file: %v", err)
		}

		if string(content) != "new content" {
			t.Error("File should be overwritten")
		}
	})
}

func TestHandler_CalculateTotalSize(t *testing.T) {
	handler := New()

	files := map[string]*FileData{
		"file1": {Size: 100},
		"file2": {Size: 200},
		"file3": {Size: 300},
	}

	total := handler.CalculateTotalSize(files)
	expected := int64(600)

	if total != expected {
		t.Errorf("CalculateTotalSize() = %d, want %d", total, expected)
	}
}

func TestHandler_VerifyRestorePermissions(t *testing.T) {
	handler := New()
	tmpDir := t.TempDir()

	// Create SSH directory with correct permissions
	err := os.MkdirAll(tmpDir, 0700)
	if err != nil {
		t.Fatalf("Failed to create SSH directory: %v", err)
	}

	// Ensure permissions are set correctly
	err = os.Chmod(tmpDir, 0700)
	if err != nil {
		t.Fatalf("Failed to create SSH directory: %v", err)
	}

	// Create test file with correct permissions
	testFile := filepath.Join(tmpDir, "test_key")
	err = os.WriteFile(testFile, []byte("test content"), 0600)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	backup := &BackupData{
		Files: map[string]*FileData{
			"test_key": {
				Filename:    "test_key",
				Content:     []byte("test content"),
				Permissions: 0600,
				KeyInfo: &analyzer.KeyInfo{
					Type: analyzer.KeyTypePrivate,
				},
			},
		},
	}

	err = handler.VerifyRestorePermissions(backup, tmpDir)
	if err != nil {
		t.Errorf("VerifyRestorePermissions() failed: %v", err)
	}
}

func TestHandler_VerifyRestorePermissions_Errors(t *testing.T) {
	handler := New()
	tmpDir := t.TempDir()

	// Create SSH directory with wrong permissions
	err := os.MkdirAll(tmpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create SSH directory: %v", err)
	}

	backup := &BackupData{
		Files: map[string]*FileData{},
	}

	err = handler.VerifyRestorePermissions(backup, tmpDir)
	if err == nil {
		t.Error("VerifyRestorePermissions() should fail with wrong SSH directory permissions")
	}
}

func TestHandler_GetAppropriatePermissions(t *testing.T) {
	handler := New()

	tests := []struct {
		name         string
		fileData     *FileData
		expectedPerm os.FileMode
	}{
		{
			name: "normal permissions",
			fileData: &FileData{
				Filename:    "id_rsa",
				Permissions: 0600,
			},
			expectedPerm: 0600,
		},
		{
			name: "public key permissions",
			fileData: &FileData{
				Filename:    "id_rsa.pub",
				Permissions: 0644,
			},
			expectedPerm: 0644,
		},
		{
			name: "zero permissions fallback",
			fileData: &FileData{
				Filename:    "corrupted",
				Permissions: 0000,
			},
			expectedPerm: 0600,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			perm := handler.getAppropriatePermissions(tt.fileData)
			if perm != tt.expectedPerm {
				t.Errorf("getAppropriatePermissions() = %04o, want %04o",
					perm, tt.expectedPerm)
			}
		})
	}
}

func TestHandler_ShouldRestoreFile(t *testing.T) {
	handler := New()

	tests := []struct {
		name     string
		filename string
		options  RestoreOptions
		want     bool
	}{
		{
			name:     "no filters",
			filename: "id_rsa",
			options:  RestoreOptions{},
			want:     true,
		},
		{
			name:     "matching file filter",
			filename: "id_rsa",
			options:  RestoreOptions{FileFilter: []string{"id_rsa*"}},
			want:     true,
		},
		{
			name:     "non-matching file filter",
			filename: "config",
			options:  RestoreOptions{FileFilter: []string{"id_rsa*"}},
			want:     false,
		},
		{
			name:     "multiple filters - match",
			filename: "config",
			options:  RestoreOptions{FileFilter: []string{"id_rsa*", "config"}},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.shouldRestoreFile(tt.filename, tt.options)
			if result != tt.want {
				t.Errorf("shouldRestoreFile() = %v, want %v", result, tt.want)
			}
		})
	}
}

// Benchmark tests
func BenchmarkHandler_ReadDirectory(b *testing.B) {
	handler := New()
	tmpDir := b.TempDir()

	// Create test files
	testFiles := map[string][]byte{
		"id_rsa":     []byte("private key content"),
		"id_rsa.pub": []byte("public key content"),
		"config":     []byte("ssh config content"),
	}

	for filename, content := range testFiles {
		filePath := filepath.Join(tmpDir, filename)
		os.WriteFile(filePath, content, 0600)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := handler.ReadDirectory(tmpDir)
		if err != nil {
			b.Fatalf("ReadDirectory failed: %v", err)
		}
	}
}

func BenchmarkHandler_EncryptDecrypt(b *testing.B) {
	handler := New()
	passphrase := "test-passphrase"

	backup := &BackupData{
		Files: map[string]*FileData{
			"test_key": {
				Filename: "test_key",
				Content:  []byte("test key content for benchmarking"),
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset content
		backup.Files["test_key"].Content = []byte("test key content for benchmarking")
		backup.Files["test_key"].Encrypted = nil

		err := handler.EncryptBackup(backup, passphrase)
		if err != nil {
			b.Fatalf("EncryptBackup failed: %v", err)
		}

		err = handler.DecryptBackup(backup, passphrase)
		if err != nil {
			b.Fatalf("DecryptBackup failed: %v", err)
		}
	}
}

// Test for Issue #17 defensive validation fix
func TestHandler_VerifyRestorePermissions_Issue17_DefensiveFix(t *testing.T) {
	handler := New()

	// Test cases simulating old backups with incorrect KeyInfo
	testCases := []struct {
		name        string
		filename    string
		permissions os.FileMode
		storedType  analyzer.KeyType // Incorrect type from old backup
		expectError bool
		description string
	}{
		{
			name:        "pub_file_stored_as_private_0644",
			filename:    "bitbucket_rsa.pub",
			permissions: 0644,
			storedType:  analyzer.KeyTypePrivate, // WRONG - old backup mistake
			expectError: false,                   // Should be fixed by defensive logic
			description: "0644 permissions on .pub file incorrectly stored as private",
		},
		{
			name:        "pub_file_stored_as_private_0600",
			filename:    "cci.pub",
			permissions: 0600,
			storedType:  analyzer.KeyTypePrivate, // WRONG - old backup mistake
			expectError: false,                   // Should be fixed by defensive logic
			description: "0600 permissions on .pub file incorrectly stored as private",
		},
		{
			name:        "actual_private_key_0644",
			filename:    "id_rsa",
			permissions: 0644,
			storedType:  analyzer.KeyTypePrivate, // CORRECT
			expectError: false,                   // VerifyRestorePermissions logs warnings but doesn't fail
			description: "Actual private key with world-readable permissions",
		},
		{
			name:        "pub_file_correctly_stored",
			filename:    "id_rsa.pub",
			permissions: 0644,
			storedType:  analyzer.KeyTypePublic, // CORRECT
			expectError: false,
			description: "Public key correctly stored and validated",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create KeyInfo with the stored type (potentially incorrect)
			keyInfo := &analyzer.KeyInfo{
				Filename: tc.filename,
				Type:     tc.storedType,
				Format:   analyzer.FormatRSA,
			}

			// Create minimal backup structure
			backup := &BackupData{
				Files: map[string]*FileData{
					tc.filename: {
						Filename:    tc.filename,
						Permissions: tc.permissions,
						KeyInfo:     keyInfo,
						Content:     []byte("test content"),
					},
				},
			}

			// Create temp directory and file for testing
			tmpDir := t.TempDir()
			// Fix SSH directory permissions for test
			if err := os.Chmod(tmpDir, 0700); err != nil {
				t.Fatalf("Failed to set directory permissions: %v", err)
			}
			tmpFile := filepath.Join(tmpDir, tc.filename)
			if err := os.WriteFile(tmpFile, []byte("test content"), tc.permissions); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Test the validation
			err := handler.VerifyRestorePermissions(backup, tmpDir)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error for %s but got none", tc.description)
				} else {
					t.Logf("✅ Expected error for %s: %v", tc.description, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for %s but got: %v", tc.description, err)
				} else {
					t.Logf("✅ No error for %s as expected", tc.description)
				}
			}
		})
	}
}
