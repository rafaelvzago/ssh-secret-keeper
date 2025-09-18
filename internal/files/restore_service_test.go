package files

import (
	"bytes"
	"crypto/md5"
	"fmt"
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

	t.Run("empty target directory defaults to user SSH dir", func(t *testing.T) {
		backup := &ssh.BackupData{
			Files: map[string]*ssh.FileData{},
		}

		err := service.RestoreFiles(backup, "", ssh.RestoreOptions{})
		if err != nil {
			t.Errorf("RestoreFiles() should succeed with empty target directory (defaults to ~/.ssh): %v", err)
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

// TestRestoreService_PermissionPreservation_EndToEnd tests that file permissions
// are preserved exactly during backup and restore operations, simulating the
// complete workflow including vault storage serialization/deserialization
func TestRestoreService_PermissionPreservation_EndToEnd(t *testing.T) {
	service := NewRestoreService()

	// Define test cases with various permission scenarios that should be preserved
	testFiles := map[string]struct {
		filename    string
		content     []byte
		permissions os.FileMode
		description string
	}{
		"private_key_0600": {
			filename:    "id_rsa",
			content:     []byte("-----BEGIN PRIVATE KEY-----\nMIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQC7\n-----END PRIVATE KEY-----\n"),
			permissions: 0600,
			description: "Private key with strict permissions (0600)",
		},
		"public_key_0644": {
			filename:    "id_rsa.pub",
			content:     []byte("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC7... user@hostname"),
			permissions: 0644,
			description: "Public key with standard permissions (0644)",
		},
		"config_file_0644": {
			filename:    "config",
			content:     []byte("Host github.com\n    HostName github.com\n    User git\n    IdentityFile ~/.ssh/id_rsa\n"),
			permissions: 0644,
			description: "SSH config file with readable permissions (0644)",
		},
		"restricted_key_0400": {
			filename:    "restricted_key",
			content:     []byte("-----BEGIN OPENSSH PRIVATE KEY-----\nb3BlbnNzaC1rZXktdjEAAAAABG5vbmU=\n-----END OPENSSH PRIVATE KEY-----\n"),
			permissions: 0400,
			description: "Read-only private key (0400)",
		},
		"known_hosts_0644": {
			filename:    "known_hosts",
			content:     []byte("github.com ssh-rsa AAAAB3NzaC1yc2EAAAABI...\n"),
			permissions: 0644,
			description: "Known hosts file with standard permissions (0644)",
		},
		"executable_script_0755": {
			filename:    "ssh-wrapper.sh",
			content:     []byte("#!/bin/bash\nssh -o StrictHostKeyChecking=no \"$@\"\n"),
			permissions: 0755,
			description: "Executable script with full permissions (0755)",
		},
	}

	// Phase 1: Create original files with specific permissions
	originalDir := t.TempDir()
	for _, testFile := range testFiles {
		filePath := filepath.Join(originalDir, testFile.filename)

		// Write file with specific permissions
		err := os.WriteFile(filePath, testFile.content, testFile.permissions)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", testFile.filename, err)
		}

		// Ensure permissions are set correctly (os.WriteFile behavior can vary)
		err = os.Chmod(filePath, testFile.permissions)
		if err != nil {
			t.Fatalf("Failed to set permissions for %s: %v", testFile.filename, err)
		}

		// Verify original permissions
		stat, err := os.Stat(filePath)
		if err != nil {
			t.Fatalf("Failed to stat original file %s: %v", testFile.filename, err)
		}

		actualPerm := stat.Mode().Perm()
		if actualPerm != testFile.permissions {
			t.Fatalf("Original file %s has incorrect permissions: expected %04o, got %04o",
				testFile.filename, testFile.permissions, actualPerm)
		}

		t.Logf("Created original file %s with permissions %04o (%s)",
			testFile.filename, actualPerm, testFile.description)
	}

	// Phase 2: Simulate backup creation (reading files and capturing metadata)
	backup := &ssh.BackupData{
		Version:   "1.0",
		Timestamp: time.Now(),
		Hostname:  "test-host",
		Username:  "test-user",
		SSHDir:    originalDir,
		Files:     make(map[string]*ssh.FileData),
		Metadata:  make(map[string]interface{}),
	}

	// Read files and capture their permissions (simulating backup process)
	for _, testFile := range testFiles {
		filePath := filepath.Join(originalDir, testFile.filename)
		stat, err := os.Stat(filePath)
		if err != nil {
			t.Fatalf("Failed to stat file during backup simulation %s: %v", testFile.filename, err)
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read file during backup simulation %s: %v", testFile.filename, err)
		}

		// Create FileData structure as it would be during backup
		backup.Files[testFile.filename] = &ssh.FileData{
			Filename:    testFile.filename,
			Content:     content,
			Permissions: stat.Mode(), // This captures the full file mode including permissions
			Size:        stat.Size(),
			ModTime:     stat.ModTime(),
			Checksum:    fmt.Sprintf("md5-%x", md5.Sum(content)),
			KeyInfo: &analyzer.KeyInfo{
				Type:     analyzer.KeyTypePrivate, // Simplified for test
				Format:   analyzer.FormatOpenSSH,
				Filename: testFile.filename,
			},
		}

		t.Logf("Captured file %s in backup with permissions %04o",
			testFile.filename, stat.Mode().Perm())
	}

	// Phase 3: Simulate vault storage serialization/deserialization
	// This simulates what happens when data is stored in and retrieved from Vault
	vaultData := make(map[string]interface{})
	vaultData["version"] = backup.Version
	vaultData["timestamp"] = backup.Timestamp.Format(time.RFC3339)
	vaultData["hostname"] = backup.Hostname
	vaultData["username"] = backup.Username
	vaultData["ssh_dir"] = backup.SSHDir

	// Serialize files as they would be stored in Vault
	filesData := make(map[string]interface{})
	for filename, fileData := range backup.Files {
		// Store permissions as integer (as done in real backup process)
		permissionsToStore := int(fileData.Permissions & os.ModePerm)

		filesData[filename] = map[string]interface{}{
			"filename":    fileData.Filename,
			"content":     string(fileData.Content),
			"permissions": permissionsToStore, // This is how permissions are stored in Vault
			"size":        fileData.Size,
			"mod_time":    fileData.ModTime.Format(time.RFC3339),
			"checksum":    fileData.Checksum,
		}

		t.Logf("Serialized file %s with permissions %d (octal: %04o)",
			filename, permissionsToStore, permissionsToStore)
	}
	vaultData["files"] = filesData

	// Phase 4: Simulate vault data deserialization (parseVaultBackup equivalent)
	restoredBackup := &ssh.BackupData{
		Version:  vaultData["version"].(string),
		Hostname: vaultData["hostname"].(string),
		Username: vaultData["username"].(string),
		SSHDir:   vaultData["ssh_dir"].(string),
		Files:    make(map[string]*ssh.FileData),
		Metadata: make(map[string]interface{}),
	}

	// Parse timestamp
	if timestampStr, ok := vaultData["timestamp"].(string); ok {
		if timestamp, err := time.Parse(time.RFC3339, timestampStr); err == nil {
			restoredBackup.Timestamp = timestamp
		}
	}

	// Deserialize files as they would be parsed from Vault
	if filesInterface, ok := vaultData["files"].(map[string]interface{}); ok {
		for filename, fileInterface := range filesInterface {
			fileMap := fileInterface.(map[string]interface{})

			fileData := &ssh.FileData{
				Filename: filename,
				Content:  []byte(fileMap["content"].(string)),
				Size:     int64(fileMap["size"].(int64)),
				Checksum: fileMap["checksum"].(string),
			}

			// Parse permissions as they would be parsed from Vault (int from our test data)
			if permsInt, ok := fileMap["permissions"].(int); ok {
				fileData.Permissions = os.FileMode(permsInt)
				t.Logf("Deserialized file %s with permissions %04o",
					filename, fileData.Permissions&os.ModePerm)
			}

			// Parse mod_time
			if modTimeStr, ok := fileMap["mod_time"].(string); ok {
				if modTime, err := time.Parse(time.RFC3339, modTimeStr); err == nil {
					fileData.ModTime = modTime
				}
			}

			restoredBackup.Files[filename] = fileData
		}
	}

	// Phase 5: Restore files to a new directory
	restoreDir := t.TempDir()

	options := ssh.RestoreOptions{
		DryRun:    false,
		Overwrite: true,
	}

	err := service.RestoreFiles(restoredBackup, restoreDir, options)
	if err != nil {
		t.Fatalf("RestoreFiles failed: %v", err)
	}

	// Phase 6: Verify that all permissions are preserved exactly
	for _, testFile := range testFiles {
		originalPath := filepath.Join(originalDir, testFile.filename)
		restoredPath := filepath.Join(restoreDir, testFile.filename)

		// Get original file permissions
		originalStat, err := os.Stat(originalPath)
		if err != nil {
			t.Errorf("Failed to stat original file %s: %v", testFile.filename, err)
			continue
		}
		originalPerms := originalStat.Mode().Perm()

		// Get restored file permissions
		restoredStat, err := os.Stat(restoredPath)
		if err != nil {
			t.Errorf("Failed to stat restored file %s: %v", testFile.filename, err)
			continue
		}
		restoredPerms := restoredStat.Mode().Perm()

		// Verify permissions match exactly
		if originalPerms != restoredPerms {
			t.Errorf("Permission mismatch for %s (%s):\n  Original:  %04o\n  Restored:  %04o\n  Expected:  %04o",
				testFile.filename, testFile.description, originalPerms, restoredPerms, testFile.permissions)
		} else {
			t.Logf("✓ Permissions preserved for %s: %04o (%s)",
				testFile.filename, restoredPerms, testFile.description)
		}

		// Verify content is also preserved
		originalContent, err := os.ReadFile(originalPath)
		if err != nil {
			t.Errorf("Failed to read original file content %s: %v", testFile.filename, err)
			continue
		}

		restoredContent, err := os.ReadFile(restoredPath)
		if err != nil {
			t.Errorf("Failed to read restored file content %s: %v", testFile.filename, err)
			continue
		}

		if !bytes.Equal(originalContent, restoredContent) {
			t.Errorf("Content mismatch for %s", testFile.filename)
		} else {
			t.Logf("✓ Content preserved for %s (%d bytes)", testFile.filename, len(restoredContent))
		}

		// Additional verification: check that the restored permissions match expected
		if restoredPerms != testFile.permissions {
			t.Errorf("Restored file %s permissions %04o do not match expected %04o",
				testFile.filename, restoredPerms, testFile.permissions)
		}
	}

	// Phase 7: Verify SSH directory permissions
	sshDirStat, err := os.Stat(restoreDir)
	if err != nil {
		t.Errorf("Failed to stat SSH directory: %v", err)
	} else {
		sshDirPerms := sshDirStat.Mode().Perm()
		if sshDirPerms != 0700 {
			t.Errorf("SSH directory permissions incorrect: expected 0700, got %04o", sshDirPerms)
		} else {
			t.Logf("✓ SSH directory permissions correct: %04o", sshDirPerms)
		}
	}

	// Phase 8: Run permission verification
	err = service.VerifyRestorePermissions(restoredBackup, restoreDir)
	if err != nil {
		t.Errorf("Permission verification failed: %v", err)
	} else {
		t.Logf("✓ All permission verification checks passed")
	}

	t.Logf("Permission preservation test completed successfully for %d files", len(testFiles))
}

// TestRestoreService_PermissionPreservation_EdgeCases tests edge cases and error scenarios
func TestRestoreService_PermissionPreservation_EdgeCases(t *testing.T) {
	service := NewRestoreService()

	t.Run("zero permissions fallback", func(t *testing.T) {
		// Test that 0000 permissions use fallback
		backup := &ssh.BackupData{
			Files: map[string]*ssh.FileData{
				"corrupted_file": {
					Filename:    "corrupted_file",
					Content:     []byte("test content"),
					Permissions: 0000, // Corrupted/zero permissions
					Size:        12,
					ModTime:     time.Now(),
					Checksum:    "test-checksum",
				},
			},
		}

		tmpDir := t.TempDir()
		options := ssh.RestoreOptions{DryRun: false}

		err := service.RestoreFiles(backup, tmpDir, options)
		if err != nil {
			t.Fatalf("RestoreFiles failed: %v", err)
		}

		// Verify fallback permissions (0600) were applied
		restoredPath := filepath.Join(tmpDir, "corrupted_file")
		stat, err := os.Stat(restoredPath)
		if err != nil {
			t.Fatalf("Failed to stat restored file: %v", err)
		}

		actualPerms := stat.Mode().Perm()
		expectedPerms := os.FileMode(0600) // Fallback
		if actualPerms != expectedPerms {
			t.Errorf("Expected fallback permissions %04o, got %04o", expectedPerms, actualPerms)
		}
	})

	t.Run("permission preservation with overwrite", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create existing file with different permissions
		existingFile := filepath.Join(tmpDir, "existing_key")
		err := os.WriteFile(existingFile, []byte("old content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create existing file: %v", err)
		}

		// Backup data with different permissions
		backup := &ssh.BackupData{
			Files: map[string]*ssh.FileData{
				"existing_key": {
					Filename:    "existing_key",
					Content:     []byte("new content"),
					Permissions: 0600, // Different from existing file
					Size:        11,
					ModTime:     time.Now(),
					Checksum:    "new-checksum",
				},
			},
		}

		options := ssh.RestoreOptions{
			DryRun:    false,
			Overwrite: true,
		}

		err = service.RestoreFiles(backup, tmpDir, options)
		if err != nil {
			t.Fatalf("RestoreFiles failed: %v", err)
		}

		// Verify new permissions were applied
		stat, err := os.Stat(existingFile)
		if err != nil {
			t.Fatalf("Failed to stat overwritten file: %v", err)
		}

		actualPerms := stat.Mode().Perm()
		expectedPerms := os.FileMode(0600)
		if actualPerms != expectedPerms {
			t.Errorf("Expected overwritten file permissions %04o, got %04o", expectedPerms, actualPerms)
		}

		// Verify content was also updated
		content, err := os.ReadFile(existingFile)
		if err != nil {
			t.Fatalf("Failed to read overwritten file: %v", err)
		}
		if string(content) != "new content" {
			t.Errorf("Expected content 'new content', got '%s'", string(content))
		}
	})

	t.Run("various permission combinations", func(t *testing.T) {
		// Test various permission combinations that should be preserved exactly
		permissionTests := []struct {
			name        string
			permissions os.FileMode
		}{
			{"owner_read_only", 0400},
			{"owner_write_only", 0200},
			{"owner_execute_only", 0100},
			{"owner_read_write", 0600},
			{"owner_read_execute", 0500},
			{"owner_all_group_read", 0640},
			{"owner_all_world_read", 0644},
			{"all_permissions", 0777},
			{"group_write", 0620},
			{"world_write", 0602},
		}

		for _, tt := range permissionTests {
			t.Run(tt.name, func(t *testing.T) {
				tmpDir := t.TempDir()
				filename := "test_" + tt.name

				backup := &ssh.BackupData{
					Files: map[string]*ssh.FileData{
						filename: {
							Filename:    filename,
							Content:     []byte("test content for " + tt.name),
							Permissions: tt.permissions,
							Size:        int64(len("test content for " + tt.name)),
							ModTime:     time.Now(),
							Checksum:    "test-checksum-" + tt.name,
						},
					},
				}

				options := ssh.RestoreOptions{DryRun: false}
				err := service.RestoreFiles(backup, tmpDir, options)
				if err != nil {
					t.Fatalf("RestoreFiles failed for %s: %v", tt.name, err)
				}

				// Verify exact permission preservation
				restoredPath := filepath.Join(tmpDir, filename)
				stat, err := os.Stat(restoredPath)
				if err != nil {
					t.Fatalf("Failed to stat restored file %s: %v", filename, err)
				}

				actualPerms := stat.Mode().Perm()
				if actualPerms != tt.permissions {
					t.Errorf("Permission mismatch for %s: expected %04o, got %04o",
						tt.name, tt.permissions, actualPerms)
				}
			})
		}
	})
}

// TestRestoreService_PermissionPreservation_RealWorldScenario tests with realistic SSH file scenarios
func TestRestoreService_PermissionPreservation_RealWorldScenario(t *testing.T) {
	service := NewRestoreService()

	// Real-world SSH directory scenario with typical files and their expected permissions
	realWorldFiles := map[string]struct {
		content     []byte
		permissions os.FileMode
		description string
	}{
		"id_rsa": {
			content: []byte(`-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAFwAAAAdzc2gtcn
NhAAAAAwEAAQAAAQEAwJKJFJkJFJkJFJkJFJkJFJkJFJkJFJkJFJkJFJkJFJkJFJkJFJkJ
-----END OPENSSH PRIVATE KEY-----`),
			permissions: 0600,
			description: "Private RSA key (must be 0600 for security)",
		},
		"id_rsa.pub": {
			content:     []byte("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC7... user@hostname"),
			permissions: 0644,
			description: "Public RSA key (can be world-readable)",
		},
		"id_ed25519": {
			content: []byte(`-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW
QyNTUxOQAAACBkJFJkJFJkJFJkJFJkJFJkJFJkJFJkJFJkJFJkJFJkJFJkJFJkJFJkJFJk
-----END OPENSSH PRIVATE KEY-----`),
			permissions: 0600,
			description: "Private Ed25519 key (must be 0600)",
		},
		"id_ed25519.pub": {
			content:     []byte("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIGQkUmQkUmQkUmQkUmQkUmQkUmQkUmQkUmQkUmQkUmQk user@hostname"),
			permissions: 0644,
			description: "Public Ed25519 key",
		},
		"config": {
			content: []byte(`Host github.com
    HostName github.com
    User git
    IdentityFile ~/.ssh/id_rsa

Host gitlab.com
    HostName gitlab.com
    User git
    IdentityFile ~/.ssh/id_ed25519`),
			permissions: 0644,
			description: "SSH client configuration",
		},
		"known_hosts": {
			content: []byte(`github.com ssh-rsa AAAAB3NzaC1yc2EAAAABI...
gitlab.com ecdsa-sha2-nistp256 AAAAE2VjZHNh...`),
			permissions: 0644,
			description: "Known hosts file",
		},
		"authorized_keys": {
			content: []byte(`ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC7... admin@server1
ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIGQkUmQkUmQkUmQkUmQkUmQkUmQkUmQkUmQkUmQkUmQk admin@server2`),
			permissions: 0600,
			description: "Authorized keys (should be 0600 for security)",
		},
	}

	// Create original directory with real-world SSH files
	originalDir := t.TempDir()
	for filename, fileInfo := range realWorldFiles {
		filePath := filepath.Join(originalDir, filename)

		err := os.WriteFile(filePath, fileInfo.content, fileInfo.permissions)
		if err != nil {
			t.Fatalf("Failed to create %s: %v", filename, err)
		}

		// Ensure permissions are set correctly
		err = os.Chmod(filePath, fileInfo.permissions)
		if err != nil {
			t.Fatalf("Failed to set permissions for %s: %v", filename, err)
		}
	}

	// Create backup data structure
	backup := &ssh.BackupData{
		Version:   "1.0",
		Timestamp: time.Now(),
		Hostname:  "test-host",
		Username:  "test-user",
		SSHDir:    originalDir,
		Files:     make(map[string]*ssh.FileData),
		Metadata:  make(map[string]interface{}),
	}

	// Read files into backup structure
	for filename := range realWorldFiles {
		filePath := filepath.Join(originalDir, filename)
		stat, err := os.Stat(filePath)
		if err != nil {
			t.Fatalf("Failed to stat %s: %v", filename, err)
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read %s: %v", filename, err)
		}

		backup.Files[filename] = &ssh.FileData{
			Filename:    filename,
			Content:     content,
			Permissions: stat.Mode(),
			Size:        stat.Size(),
			ModTime:     stat.ModTime(),
			Checksum:    fmt.Sprintf("md5-%x", md5.Sum(content)),
		}
	}

	// Restore to new directory
	restoreDir := t.TempDir()
	options := ssh.RestoreOptions{
		DryRun:    false,
		Overwrite: true,
	}

	err := service.RestoreFiles(backup, restoreDir, options)
	if err != nil {
		t.Fatalf("RestoreFiles failed: %v", err)
	}

	// Verify all permissions and content
	for filename, expectedInfo := range realWorldFiles {
		originalPath := filepath.Join(originalDir, filename)
		restoredPath := filepath.Join(restoreDir, filename)

		// Check permissions
		originalStat, err := os.Stat(originalPath)
		if err != nil {
			t.Errorf("Failed to stat original %s: %v", filename, err)
			continue
		}

		restoredStat, err := os.Stat(restoredPath)
		if err != nil {
			t.Errorf("Failed to stat restored %s: %v", filename, err)
			continue
		}

		originalPerms := originalStat.Mode().Perm()
		restoredPerms := restoredStat.Mode().Perm()

		if originalPerms != restoredPerms {
			t.Errorf("Permission mismatch for %s (%s): original=%04o, restored=%04o",
				filename, expectedInfo.description, originalPerms, restoredPerms)
		}

		// Verify against expected permissions
		if restoredPerms != expectedInfo.permissions {
			t.Errorf("Restored %s has incorrect permissions: expected=%04o, actual=%04o (%s)",
				filename, expectedInfo.permissions, restoredPerms, expectedInfo.description)
		}

		// Check content integrity
		originalContent, _ := os.ReadFile(originalPath)
		restoredContent, _ := os.ReadFile(restoredPath)

		if !bytes.Equal(originalContent, restoredContent) {
			t.Errorf("Content mismatch for %s", filename)
		}

		t.Logf("✓ %s: permissions=%04o, size=%d bytes (%s)",
			filename, restoredPerms, len(restoredContent), expectedInfo.description)
	}

	// Run comprehensive permission verification
	err = service.VerifyRestorePermissions(backup, restoreDir)
	if err != nil {
		t.Errorf("Permission verification failed: %v", err)
	}

	t.Logf("Real-world scenario test completed successfully with %d SSH files", len(realWorldFiles))
}
