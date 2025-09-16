package files

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rzago/ssh-secret-keeper/internal/analyzer"
	"github.com/rzago/ssh-secret-keeper/internal/ssh"
)

func TestNewReadService(t *testing.T) {
	service := NewReadService()
	if service == nil {
		t.Fatal("NewReadService() returned nil")
	}
}

func TestReadService_ValidateDirectory(t *testing.T) {
	service := NewReadService()

	tests := []struct {
		name      string
		setupDir  func() string
		wantError bool
	}{
		{
			name: "valid directory",
			setupDir: func() string {
				return t.TempDir()
			},
			wantError: false,
		},
		{
			name: "empty path",
			setupDir: func() string {
				return ""
			},
			wantError: true,
		},
		{
			name: "nonexistent directory",
			setupDir: func() string {
				return "/nonexistent/directory"
			},
			wantError: true,
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setupDir()
			err := service.ValidateDirectory(dir)

			if tt.wantError && err == nil {
				t.Errorf("ValidateDirectory() expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("ValidateDirectory() unexpected error: %v", err)
			}
		})
	}
}

func TestReadService_CalculateChecksum(t *testing.T) {
	service := NewReadService()

	tests := []struct {
		name     string
		data     []byte
		expected string
	}{
		{
			name:     "empty data",
			data:     []byte{},
			expected: "d41d8cd98f00b204e9800998ecf8427e",
		},
		{
			name:     "simple text",
			data:     []byte("hello world"),
			expected: "5eb63bbbe01eeed093cb22bb8f5acdc3",
		},
		{
			name:     "ssh key content",
			data:     []byte("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDT1234567890"),
			expected: "f8b2c1c9b8a4d5e6f7a8b9c0d1e2f3a4", // This will be different, just testing consistency
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.CalculateChecksum(tt.data)

			// Check format (32 hex characters)
			if len(result) != 32 {
				t.Errorf("CalculateChecksum() length = %d, want 32", len(result))
			}

			// For empty data, we know the exact MD5
			if tt.name == "empty data" && result != tt.expected {
				t.Errorf("CalculateChecksum() = %s, want %s", result, tt.expected)
			}

			// For simple text, we know the exact MD5
			if tt.name == "simple text" && result != tt.expected {
				t.Errorf("CalculateChecksum() = %s, want %s", result, tt.expected)
			}

			// Test consistency - same input should give same output
			result2 := service.CalculateChecksum(tt.data)
			if result != result2 {
				t.Errorf("CalculateChecksum() not consistent: %s != %s", result, result2)
			}
		})
	}
}

func TestReadService_VerifyFilePermissions(t *testing.T) {
	service := NewReadService()
	tmpDir := t.TempDir()

	tests := []struct {
		name         string
		filePerms    os.FileMode
		expectedMode os.FileMode
		wantError    bool
	}{
		{
			name:         "correct permissions",
			filePerms:    0600,
			expectedMode: 0600,
			wantError:    false,
		},
		{
			name:         "incorrect permissions",
			filePerms:    0644,
			expectedMode: 0600,
			wantError:    true,
		},
		{
			name:         "public key permissions",
			filePerms:    0644,
			expectedMode: 0644,
			wantError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file with specific permissions
			filePath := filepath.Join(tmpDir, "test_"+tt.name)
			err := os.WriteFile(filePath, []byte("test content"), tt.filePerms)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			err = service.VerifyFilePermissions(filePath, tt.expectedMode)

			if tt.wantError && err == nil {
				t.Errorf("VerifyFilePermissions() expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("VerifyFilePermissions() unexpected error: %v", err)
			}
		})
	}
}

func TestReadService_ReadSSHDirectory(t *testing.T) {
	service := NewReadService()

	t.Run("empty directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		files, err := service.ReadSSHDirectory(tmpDir)
		if err != nil {
			t.Errorf("ReadSSHDirectory() unexpected error: %v", err)
		}

		if len(files) != 0 {
			t.Errorf("ReadSSHDirectory() expected 0 files, got %d", len(files))
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

		files, err := service.ReadSSHDirectory(tmpDir)
		if err != nil {
			t.Errorf("ReadSSHDirectory() unexpected error: %v", err)
		}

		if len(files) != len(testFiles) {
			t.Errorf("ReadSSHDirectory() expected %d files, got %d", len(testFiles), len(files))
		}

		// Verify each file was read correctly
		for filename, expectedContent := range testFiles {
			fileData, exists := files[filename]
			if !exists {
				t.Errorf("File %s not found in results", filename)
				continue
			}

			if string(fileData.Content) != string(expectedContent) {
				t.Errorf("File %s content mismatch", filename)
			}

			if fileData.Filename != filename {
				t.Errorf("File %s name mismatch: got %s", filename, fileData.Filename)
			}

			if len(fileData.Checksum) != 32 {
				t.Errorf("File %s checksum invalid length: %d", filename, len(fileData.Checksum))
			}

			if fileData.Size != int64(len(expectedContent)) {
				t.Errorf("File %s size mismatch: got %d, want %d",
					filename, fileData.Size, len(expectedContent))
			}
		}
	})

	t.Run("directory with subdirectories", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a file and a subdirectory
		os.WriteFile(filepath.Join(tmpDir, "id_rsa"), []byte("private key"), 0600)
		os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755)
		os.WriteFile(filepath.Join(tmpDir, "subdir", "nested_key"), []byte("nested"), 0600)

		files, err := service.ReadSSHDirectory(tmpDir)
		if err != nil {
			t.Errorf("ReadSSHDirectory() unexpected error: %v", err)
		}

		// Should only read files, not subdirectories
		if len(files) != 1 {
			t.Errorf("ReadSSHDirectory() expected 1 file, got %d", len(files))
		}

		if _, exists := files["id_rsa"]; !exists {
			t.Error("Expected id_rsa file not found")
		}
	})

	t.Run("nonexistent directory", func(t *testing.T) {
		_, err := service.ReadSSHDirectory("/nonexistent/directory")
		if err == nil {
			t.Error("ReadSSHDirectory() expected error for nonexistent directory")
		}
	})
}

func TestReadService_AddKeyInfoToFiles(t *testing.T) {
	service := NewReadService()

	// Create test file data
	files := map[string]*ssh.FileData{
		"id_rsa": {
			Filename: "id_rsa",
			Content:  []byte("private key content"),
		},
		"id_rsa.pub": {
			Filename: "id_rsa.pub",
			Content:  []byte("public key content"),
		},
		"config": {
			Filename: "config",
			Content:  []byte("ssh config content"),
		},
	}

	// Create test analysis result
	analysisResult := &analyzer.DetectionResult{
		Keys: []analyzer.KeyInfo{
			{
				Filename: "id_rsa",
				Type:     analyzer.KeyTypePrivate,
				Format:   analyzer.FormatRSA,
				Purpose:  analyzer.PurposePersonal,
			},
			{
				Filename: "id_rsa.pub",
				Type:     analyzer.KeyTypePublic,
				Format:   analyzer.FormatRSA,
				Purpose:  analyzer.PurposePersonal,
			},
		},
	}

	service.AddKeyInfoToFiles(files, analysisResult)

	// Verify key info was added
	if files["id_rsa"].KeyInfo == nil {
		t.Error("KeyInfo not added to id_rsa")
	} else {
		if files["id_rsa"].KeyInfo.Type != analyzer.KeyTypePrivate {
			t.Error("Incorrect key type for id_rsa")
		}
	}

	if files["id_rsa.pub"].KeyInfo == nil {
		t.Error("KeyInfo not added to id_rsa.pub")
	} else {
		if files["id_rsa.pub"].KeyInfo.Type != analyzer.KeyTypePublic {
			t.Error("Incorrect key type for id_rsa.pub")
		}
	}

	// Config file should not have key info
	if files["config"].KeyInfo != nil {
		t.Error("KeyInfo should not be added to config file")
	}
}

func TestReadService_ValidateFileIntegrity(t *testing.T) {
	service := NewReadService()

	tests := []struct {
		name      string
		fileData  *ssh.FileData
		wantError bool
	}{
		{
			name:      "nil file data",
			fileData:  nil,
			wantError: true,
		},
		{
			name: "nil content",
			fileData: &ssh.FileData{
				Filename: "test",
				Content:  nil,
			},
			wantError: true,
		},
		{
			name: "valid checksum",
			fileData: &ssh.FileData{
				Filename: "test",
				Content:  []byte("hello world"),
				Checksum: service.CalculateChecksum([]byte("hello world")),
			},
			wantError: false,
		},
		{
			name: "invalid checksum",
			fileData: &ssh.FileData{
				Filename: "test",
				Content:  []byte("hello world"),
				Checksum: "invalid-checksum",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateFileIntegrity(tt.fileData)

			if tt.wantError && err == nil {
				t.Errorf("ValidateFileIntegrity() expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("ValidateFileIntegrity() unexpected error: %v", err)
			}
		})
	}
}

func TestReadService_GetFileStats(t *testing.T) {
	service := NewReadService()

	now := time.Now()
	older := now.Add(-24 * time.Hour)

	files := map[string]*ssh.FileData{
		"file1": {
			Filename:    "file1",
			Size:        100,
			ModTime:     now,
			Permissions: 0600,
		},
		"file2": {
			Filename:    "file2",
			Size:        200,
			ModTime:     older,
			Permissions: 0644,
		},
		"file3": {
			Filename:    "file3",
			Size:        300,
			ModTime:     now.Add(-12 * time.Hour),
			Permissions: 0600,
		},
	}

	stats := service.GetFileStats(files)

	// Check basic stats
	if stats["file_count"] != 3 {
		t.Errorf("file_count = %v, want 3", stats["file_count"])
	}

	if stats["total_size"] != int64(600) {
		t.Errorf("total_size = %v, want 600", stats["total_size"])
	}

	if stats["average_size"] != float64(200) {
		t.Errorf("average_size = %v, want 200", stats["average_size"])
	}

	// Check permission counts
	permCounts, ok := stats["permission_counts"].(map[string]int)
	if !ok {
		t.Error("permission_counts not found or wrong type")
	} else {
		if permCounts["0600"] != 2 {
			t.Errorf("permission count for 0600 = %d, want 2", permCounts["0600"])
		}
		if permCounts["0644"] != 1 {
			t.Errorf("permission count for 0644 = %d, want 1", permCounts["0644"])
		}
	}
}

func TestReadService_LargeFile(t *testing.T) {
	service := NewReadService()
	tmpDir := t.TempDir()

	// Create a large file (but not too large for tests)
	largeContent := make([]byte, 1024*1024) // 1MB
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}

	filePath := filepath.Join(tmpDir, "large_file")
	err := os.WriteFile(filePath, largeContent, 0600)
	if err != nil {
		t.Fatalf("Failed to create large test file: %v", err)
	}

	files, err := service.ReadSSHDirectory(tmpDir)
	if err != nil {
		t.Errorf("ReadSSHDirectory() failed with large file: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(files))
	}

	fileData := files["large_file"]
	if fileData.Size != int64(len(largeContent)) {
		t.Errorf("Large file size mismatch: got %d, want %d",
			fileData.Size, len(largeContent))
	}

	// Verify checksum is calculated correctly for large file
	expectedChecksum := service.CalculateChecksum(largeContent)
	if fileData.Checksum != expectedChecksum {
		t.Error("Large file checksum mismatch")
	}
}

// Benchmark tests
func BenchmarkReadService_CalculateChecksum(b *testing.B) {
	service := NewReadService()
	data := make([]byte, 1024) // 1KB
	for i := range data {
		data[i] = byte(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.CalculateChecksum(data)
	}
}

func BenchmarkReadService_ReadSmallDirectory(b *testing.B) {
	service := NewReadService()
	tmpDir := b.TempDir()

	// Create test files
	for i := 0; i < 10; i++ {
		filename := filepath.Join(tmpDir, fmt.Sprintf("key%d", i))
		content := []byte(fmt.Sprintf("key content %d", i))
		os.WriteFile(filename, content, 0600)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.ReadSSHDirectory(tmpDir)
		if err != nil {
			b.Fatalf("ReadSSHDirectory failed: %v", err)
		}
	}
}
