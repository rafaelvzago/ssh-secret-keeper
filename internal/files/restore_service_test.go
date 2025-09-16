package files

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rzago/ssh-secret-keeper/internal/analyzer"
	"github.com/rzago/ssh-secret-keeper/internal/ssh"
)

func TestNewRestoreService(t *testing.T) {
	service := NewRestoreService()
	if service == nil {
		t.Fatal("NewRestoreService() returned nil")
	}
}

func TestRestoreService_ValidateRestoreTarget(t *testing.T) {
	service := NewRestoreService()

	tests := []struct {
		name      string
		setupDir  func() string
		wantError bool
	}{
		{
			name: "empty path",
			setupDir: func() string {
				return ""
			},
			wantError: true,
		},
		{
			name: "existing directory",
			setupDir: func() string {
				return t.TempDir()
			},
			wantError: false,
		},
		{
			name: "new directory with existing parent",
			setupDir: func() string {
				tmpDir := t.TempDir()
				return filepath.Join(tmpDir, "new_ssh_dir")
			},
			wantError: false,
		},
		{
			name: "file instead of directory",
			setupDir: func() string {
				tmpDir := t.TempDir()
				filePath := filepath.Join(tmpDir, "notadir")
				os.WriteFile(filePath, []byte("test"), 0644)
				return filePath
			},
			wantError: true,
		},
		{
			name: "nonexistent parent directory",
			setupDir: func() string {
				return "/nonexistent/parent/ssh"
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setupDir()
			err := service.ValidateRestoreTarget(dir)

			if tt.wantError && err == nil {
				t.Errorf("ValidateRestoreTarget() expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("ValidateRestoreTarget() unexpected error: %v", err)
			}
		})
	}
}

func TestRestoreService_CreateSSHDirectory(t *testing.T) {
	service := NewRestoreService()

	t.Run("create new directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		sshDir := filepath.Join(tmpDir, "new_ssh")

		err := service.CreateSSHDirectory(sshDir)
		if err != nil {
			t.Errorf("CreateSSHDirectory() unexpected error: %v", err)
		}

		// Verify directory exists
		stat, err := os.Stat(sshDir)
		if err != nil {
			t.Errorf("SSH directory was not created: %v", err)
		}

		if !stat.IsDir() {
			t.Error("Created path is not a directory")
		}

		// Verify permissions
		if perm := stat.Mode().Perm(); perm != 0700 {
			t.Errorf("SSH directory permissions = %04o, want 0700", perm)
		}
	})

	t.Run("existing directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create directory with wrong permissions
		err := os.MkdirAll(tmpDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		err = service.CreateSSHDirectory(tmpDir)
		if err != nil {
			t.Errorf("CreateSSHDirectory() unexpected error: %v", err)
		}

		// Verify permissions were fixed
		stat, err := os.Stat(tmpDir)
		if err != nil {
			t.Errorf("Cannot stat directory: %v", err)
		}

		if perm := stat.Mode().Perm(); perm != 0700 {
			t.Errorf("SSH directory permissions = %04o, want 0700", perm)
		}
	})
}

func TestRestoreService_RestoreFiles_DryRun(t *testing.T) {
	service := NewRestoreService()
	tmpDir := t.TempDir()

	// Create test backup data
	backup := &ssh.BackupData{
		Files: map[string]*ssh.FileData{
			"id_rsa": {
				Filename:    "id_rsa",
				Content:     []byte("private key content"),
				Permissions: 0600,
				Size:        19,
				ModTime:     time.Now(),
				Checksum:    "test-checksum-1",
			},
			"id_rsa.pub": {
				Filename:    "id_rsa.pub",
				Content:     []byte("public key content"),
				Permissions: 0644,
				Size:        18,
				ModTime:     time.Now(),
				Checksum:    "test-checksum-2",
			},
		},
	}

	options := ssh.RestoreOptions{
		DryRun: true,
	}

	err := service.RestoreFiles(backup, tmpDir, options)
	if err != nil {
		t.Errorf("RestoreFiles() dry run failed: %v", err)
	}

	// Verify no files were actually created in dry run mode
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("Dry run created files: %d files found", len(files))
	}
}

func TestRestoreService_RestoreFiles_Actual(t *testing.T) {
	service := NewRestoreService()
	tmpDir := t.TempDir()

	testContent1 := []byte("private key content")
	testContent2 := []byte("public key content")

	// Create test backup data
	backup := &ssh.BackupData{
		Files: map[string]*ssh.FileData{
			"id_rsa": {
				Filename:    "id_rsa",
				Content:     testContent1,
				Permissions: 0600,
				Size:        int64(len(testContent1)),
				ModTime:     time.Now(),
				Checksum:    "test-checksum-1",
			},
			"id_rsa.pub": {
				Filename:    "id_rsa.pub",
				Content:     testContent2,
				Permissions: 0644,
				Size:        int64(len(testContent2)),
				ModTime:     time.Now(),
				Checksum:    "test-checksum-2",
			},
		},
	}

	options := ssh.RestoreOptions{
		DryRun: false,
	}

	err := service.RestoreFiles(backup, tmpDir, options)
	if err != nil {
		t.Errorf("RestoreFiles() failed: %v", err)
	}

	// Verify files were created
	for filename, expectedData := range backup.Files {
		filePath := filepath.Join(tmpDir, filename)

		// Check file exists
		stat, err := os.Stat(filePath)
		if err != nil {
			t.Errorf("File %s was not created: %v", filename, err)
			continue
		}

		// Check permissions
		if perm := stat.Mode().Perm(); perm != (expectedData.Permissions & os.ModePerm) {
			t.Errorf("File %s permissions = %04o, want %04o",
				filename, perm, expectedData.Permissions&os.ModePerm)
		}

		// Check content
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Errorf("Failed to read restored file %s: %v", filename, err)
			continue
		}

		if string(content) != string(expectedData.Content) {
			t.Errorf("File %s content mismatch", filename)
		}
	}
}

func TestRestoreService_RestoreFiles_WithFilters(t *testing.T) {
	service := NewRestoreService()
	tmpDir := t.TempDir()

	// Create test backup with key info
	backup := &ssh.BackupData{
		Files: map[string]*ssh.FileData{
			"id_rsa": {
				Filename:    "id_rsa",
				Content:     []byte("private key"),
				Permissions: 0600,
				KeyInfo: &analyzer.KeyInfo{
					Type: analyzer.KeyTypePrivate,
				},
			},
			"id_rsa.pub": {
				Filename:    "id_rsa.pub",
				Content:     []byte("public key"),
				Permissions: 0644,
				KeyInfo: &analyzer.KeyInfo{
					Type: analyzer.KeyTypePublic,
				},
			},
			"config": {
				Filename:    "config",
				Content:     []byte("ssh config"),
				Permissions: 0644,
			},
		},
	}

	tests := []struct {
		name          string
		options       ssh.RestoreOptions
		expectedFiles []string
		excludedFiles []string
	}{
		{
			name: "file filter",
			options: ssh.RestoreOptions{
				FileFilter: []string{"id_rsa*"},
			},
			expectedFiles: []string{"id_rsa", "id_rsa.pub"},
			excludedFiles: []string{"config"},
		},
		{
			name: "type filter - private keys only",
			options: ssh.RestoreOptions{
				TypeFilter: []string{string(analyzer.KeyTypePrivate)},
			},
			expectedFiles: []string{"id_rsa"},
			excludedFiles: []string{"id_rsa.pub"},
		},
		{
			name: "type filter - public keys only",
			options: ssh.RestoreOptions{
				TypeFilter: []string{string(analyzer.KeyTypePublic)},
			},
			expectedFiles: []string{"id_rsa.pub"},
			excludedFiles: []string{"id_rsa"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create separate directory for this test
			testDir := filepath.Join(tmpDir, tt.name)
			os.MkdirAll(testDir, 0755)

			err := service.RestoreFiles(backup, testDir, tt.options)
			if err != nil {
				t.Errorf("RestoreFiles() failed: %v", err)
			}

			// Check expected files exist
			for _, filename := range tt.expectedFiles {
				filePath := filepath.Join(testDir, filename)
				if _, err := os.Stat(filePath); err != nil {
					t.Errorf("Expected file %s not found: %v", filename, err)
				}
			}

			// Check excluded files don't exist
			for _, filename := range tt.excludedFiles {
				filePath := filepath.Join(testDir, filename)
				if _, err := os.Stat(filePath); err == nil {
					t.Errorf("Excluded file %s was created", filename)
				}
			}
		})
	}
}

func TestRestoreService_RestoreFiles_ExistingFiles(t *testing.T) {
	service := NewRestoreService()
	tmpDir := t.TempDir()

	// Create existing file
	existingFile := filepath.Join(tmpDir, "id_rsa")
	err := os.WriteFile(existingFile, []byte("existing content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create existing file: %v", err)
	}

	backup := &ssh.BackupData{
		Files: map[string]*ssh.FileData{
			"id_rsa": {
				Filename:    "id_rsa",
				Content:     []byte("new content"),
				Permissions: 0600,
			},
		},
	}

	tests := []struct {
		name          string
		options       ssh.RestoreOptions
		wantContent   string
		wantOverwrite bool
	}{
		{
			name: "skip existing files (default)",
			options: ssh.RestoreOptions{
				Overwrite: false,
			},
			wantContent:   "existing content",
			wantOverwrite: false,
		},
		{
			name: "overwrite existing files",
			options: ssh.RestoreOptions{
				Overwrite: true,
			},
			wantContent:   "new content",
			wantOverwrite: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset file content
			os.WriteFile(existingFile, []byte("existing content"), 0644)

			err := service.RestoreFiles(backup, tmpDir, tt.options)
			if err != nil {
				t.Errorf("RestoreFiles() failed: %v", err)
			}

			// Check file content
			content, err := os.ReadFile(existingFile)
			if err != nil {
				t.Errorf("Failed to read file: %v", err)
			}

			if string(content) != tt.wantContent {
				t.Errorf("File content = %s, want %s", string(content), tt.wantContent)
			}
		})
	}
}

func TestRestoreService_VerifyRestorePermissions(t *testing.T) {
	service := NewRestoreService()
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

	// Create test files with various permissions
	testFiles := map[string]os.FileMode{
		"private_key": 0600,
		"public_key":  0644,
		"config":      0644,
	}

	backup := &ssh.BackupData{
		Files: make(map[string]*ssh.FileData),
	}

	for filename, perm := range testFiles {
		filePath := filepath.Join(tmpDir, filename)
		content := []byte("test content for " + filename)

		err := os.WriteFile(filePath, content, perm)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}

		// Add to backup data
		backup.Files[filename] = &ssh.FileData{
			Filename:    filename,
			Content:     content,
			Permissions: perm,
			KeyInfo: &analyzer.KeyInfo{
				Type: analyzer.KeyTypePrivate, // Mark as private for stricter checking
			},
		}
	}

	err = service.VerifyRestorePermissions(backup, tmpDir)
	if err != nil {
		t.Errorf("VerifyRestorePermissions() failed: %v", err)
	}
}

func TestRestoreService_VerifyRestorePermissions_Errors(t *testing.T) {
	service := NewRestoreService()
	tmpDir := t.TempDir()

	// Create SSH directory with wrong permissions
	err := os.MkdirAll(tmpDir, 0755) // Wrong permissions
	if err != nil {
		t.Fatalf("Failed to create SSH directory: %v", err)
	}

	backup := &ssh.BackupData{
		Files: make(map[string]*ssh.FileData),
	}

	err = service.VerifyRestorePermissions(backup, tmpDir)
	if err == nil {
		t.Error("VerifyRestorePermissions() should fail with wrong SSH directory permissions")
	}
}

func TestRestoreService_GetAppropriatePermissions(t *testing.T) {
	service := NewRestoreService()

	tests := []struct {
		name         string
		fileData     *ssh.FileData
		expectedPerm os.FileMode
	}{
		{
			name: "private key permissions",
			fileData: &ssh.FileData{
				Filename:    "id_rsa",
				Permissions: 0600,
			},
			expectedPerm: 0600,
		},
		{
			name: "public key permissions",
			fileData: &ssh.FileData{
				Filename:    "id_rsa.pub",
				Permissions: 0644,
			},
			expectedPerm: 0644,
		},
		{
			name: "config file permissions",
			fileData: &ssh.FileData{
				Filename:    "config",
				Permissions: 0644,
			},
			expectedPerm: 0644,
		},
		{
			name: "zero permissions fallback",
			fileData: &ssh.FileData{
				Filename:    "corrupted",
				Permissions: 0000,
			},
			expectedPerm: 0600, // Fallback
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			perm := service.getAppropriatePermissions(tt.fileData)
			if perm != tt.expectedPerm {
				t.Errorf("getAppropriatePermissions() = %04o, want %04o",
					perm, tt.expectedPerm)
			}
		})
	}
}

func TestRestoreService_ErrorHandling(t *testing.T) {
	service := NewRestoreService()

	t.Run("nil backup data", func(t *testing.T) {
		err := service.RestoreFiles(nil, t.TempDir(), ssh.RestoreOptions{})
		if err == nil {
			t.Error("RestoreFiles() should fail with nil backup data")
		}
	})

	t.Run("invalid target directory", func(t *testing.T) {
		backup := &ssh.BackupData{
			Files: map[string]*ssh.FileData{},
		}

		err := service.RestoreFiles(backup, "", ssh.RestoreOptions{})
		if err == nil {
			t.Error("RestoreFiles() should fail with empty target directory")
		}
	})

	t.Run("file with nil content", func(t *testing.T) {
		tmpDir := t.TempDir()
		backup := &ssh.BackupData{
			Files: map[string]*ssh.FileData{
				"broken": {
					Filename:    "broken",
					Content:     nil, // Nil content
					Permissions: 0600,
				},
			},
		}

		err := service.RestoreFiles(backup, tmpDir, ssh.RestoreOptions{})
		if err == nil {
			t.Error("RestoreFiles() should fail with nil file content")
		}
	})
}

func TestRestoreService_HomeDirectoryExpansion(t *testing.T) {
	service := NewRestoreService()

	// Test home directory expansion in validation
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot get home directory")
	}

	// Create a test directory in home
	testDir := filepath.Join(homeDir, ".test-ssh-restore")
	defer os.RemoveAll(testDir) // Clean up

	// Test with tilde path
	tildeDir := "~/.test-ssh-restore"

	err = service.ValidateRestoreTarget(tildeDir)
	if err != nil {
		t.Errorf("ValidateRestoreTarget() failed with tilde path: %v", err)
	}
}

// Benchmark tests
func BenchmarkRestoreService_RestoreSmallFiles(b *testing.B) {
	service := NewRestoreService()

	// Create test backup
	backup := &ssh.BackupData{
		Files: map[string]*ssh.FileData{
			"test_key": {
				Filename:    "test_key",
				Content:     []byte("small key content"),
				Permissions: 0600,
			},
		},
	}

	options := ssh.RestoreOptions{DryRun: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tmpDir := b.TempDir()
		err := service.RestoreFiles(backup, tmpDir, options)
		if err != nil {
			b.Fatalf("RestoreFiles failed: %v", err)
		}
	}
}
